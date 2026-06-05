import { useState, useCallback, useRef } from 'react'
import { useParams, useLocation } from 'react-router-dom'
import { Plus, UploadCloud, ChevronDown, Images } from 'lucide-react'
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { snapCenterToCursor } from '@dnd-kit/modifiers'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import { ScrollArea } from '@/components/ui/scroll-area'
import { buttonVariants } from '@/components/ui/button-variants'
import { cn } from '@/lib/utils'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import FolderSidebar from './FolderSidebar'
import ImageGrid from './ImageGrid'
import UploadModal from './UploadModal'
import BatchUploadModal from './BatchUploadModal'
import RightPanel from './RightPanel'
import { useQuery } from '@tanstack/react-query'
import { getFolders } from '@/lib/folders'
import { useVisionSuggestion } from '@/hooks/useVisionSuggestion'
import { handleImageDrop, handleFolderDrop, handleFileAutoUpload } from '@/lib/dragHandlers'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import type { SortEndTrigger } from '@/components/ImageGrid'

function useAppView(): AppView {
  const { folderId } = useParams<{ folderId: string }>()
  const { pathname } = useLocation()

  if (folderId) return { type: 'folder', id: folderId }
  if (pathname === '/unsorted') return { type: 'unsorted' }
  if (pathname === '/trash') return { type: 'trash' }
  return { type: 'all' }
}

function ImageDragOverlayCard({ thumbnailUrl }: { thumbnailUrl: string | null }) {
  return (
    <div className="w-20 h-20 rounded-lg overflow-hidden bg-card shadow-xl ring-1 ring-black/10 opacity-95">
      {thumbnailUrl ? (
        <img src={thumbnailUrl} className="w-full h-full object-cover" alt="" />
      ) : (
        <div className="w-full h-full bg-muted flex items-center justify-center">
          <UploadCloud className="w-6 h-6 text-muted-foreground" />
        </div>
      )}
    </div>
  )
}

export default function AppLayout() {
  const view = useAppView()
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [uploadOpen, setUploadOpen] = useState(false)
  const [batchUploadOpen, setBatchUploadOpen] = useState(false)
  const [batchInitialFiles, setBatchInitialFiles] = useState<File[]>([])
  const [selectedImage, setSelectedImage] = useState<Image | null>(null)
  const [isFileDragOver, setIsFileDragOver] = useState(false)
  const [isAutoUploading, setIsAutoUploading] = useState(false)
  const [activeDragImage, setActiveDragImage] = useState<{ id: string; thumbnailUrl: string | null } | null>(null)
  const sortEndTriggerRef = useRef<number>(0)
  const [sortEndTrigger, setSortEndTrigger] = useState<SortEndTrigger | null>(null)
  const [autoFocusTitle, setAutoFocusTitle] = useState(false)

  const folderId = view.type === 'folder' ? view.id : null

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  const { checkVision } = useVisionSuggestion()

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  )

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const data = event.active.data.current
    if (data?.type === 'image') {
      setActiveDragImage({ id: data.imageId, thumbnailUrl: data.thumbnailUrl ?? null })
    }
  }, [])

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    setActiveDragImage(null)
    const { active, over } = event
    if (!over) return

    const dragData = active.data.current
    const dropData = over.data.current
    if (!dragData) return

    // Image dropped on another image → reorder (handled inside ImageGrid via sortEndTrigger)
    if (dragData.type === 'image' && dropData?.type === 'image') {
      setSortEndTrigger({ activeId: String(active.id), overId: String(over.id), ts: ++sortEndTriggerRef.current })
      return
    }

    if (!dropData) return

    if (dragData.type === 'image') {
      try {
        const result = await handleImageDrop(getToken, dragData as Parameters<typeof handleImageDrop>[1], dropData as Parameters<typeof handleImageDrop>[2])
        if (result === 'moved') {
          queryClient.invalidateQueries({ queryKey: ['images'] })
          toast.success('Image moved')
        }
      } catch {
        toast.error('Failed to move image')
        queryClient.invalidateQueries({ queryKey: ['images'] })
      }
    } else if (dragData.type === 'folder') {
      try {
        const result = await handleFolderDrop(getToken, dragData as Parameters<typeof handleFolderDrop>[1], dropData as Parameters<typeof handleFolderDrop>[2], folders)
        if (result === 'moved') {
          queryClient.invalidateQueries({ queryKey: ['folders'] })
          toast.success('Folder moved')
        }
      } catch {
        toast.error('Failed to move folder')
      }
    }
  }, [getToken, queryClient, folders])

  const handleMainDragOver = useCallback((e: React.DragEvent<HTMLElement>) => {
    if (e.dataTransfer.types.includes('Files')) {
      e.preventDefault()
      setIsFileDragOver(true)
    }
  }, [])

  const handleMainDragLeave = useCallback((e: React.DragEvent<HTMLElement>) => {
    if (e.currentTarget.contains(e.relatedTarget as Node)) return
    setIsFileDragOver(false)
  }, [])

  const handleMainDrop = useCallback(async (e: React.DragEvent<HTMLElement>) => {
    e.preventDefault()
    setIsFileDragOver(false)

    const files = Array.from(e.dataTransfer.files)
    if (files.length === 0) return

    if (files.length > 1) {
      setBatchInitialFiles(files)
      setBatchUploadOpen(true)
      return
    }

    const file = files[0]
    setIsAutoUploading(true)
    try {
      const imageDetail = await handleFileAutoUpload(getToken, file, folderId)
      queryClient.invalidateQueries({ queryKey: ['images'] })
      checkVision(imageDetail.id)
      setAutoFocusTitle(true)
      setSelectedImage(imageDetail)
    } catch (err) {
      if ((err as Error).message === 'heic_safari_only') {
        toast.error('HEIC uploads are only supported in Safari.')
      } else if ((err as Error).message === 'unsupported_type') {
        toast.error('Unsupported file type. Use JPEG, PNG, GIF, WEBP, or AVIF.')
      } else {
        toast.error('Upload failed. Please try again.')
      }
    } finally {
      setIsAutoUploading(false)
    }
  }, [getToken, folderId, queryClient, checkVision])

  return (
    <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="flex h-screen">
        <FolderSidebar view={view} />
        <main
          className="ml-[240px] flex-1 h-screen min-w-0 relative"
          onDragOver={handleMainDragOver}
          onDragLeave={handleMainDragLeave}
          onDrop={handleMainDrop}
        >
          {(isFileDragOver || isAutoUploading) && (
            <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-background/80 backdrop-blur-sm border-2 border-dashed border-primary rounded-none pointer-events-none">
              <UploadCloud className="w-12 h-12 text-primary mb-3" />
              <p className="text-sm font-medium text-primary">
                {isAutoUploading ? 'Uploading…' : 'Drop to upload'}
              </p>
            </div>
          )}
          <ScrollArea className="h-full">
            <div className="p-6">
              <div className="flex justify-end mb-4">
                <div className="flex">
                  <button
                    className={cn(buttonVariants(), 'rounded-r-none')}
                    onClick={() => setUploadOpen(true)}
                  >
                    <Plus className="w-4 h-4 mr-1" />
                    Image
                  </button>
                  <DropdownMenu>
                    <DropdownMenuTrigger className={cn(buttonVariants(), 'rounded-l-none border-l border-l-white/20 px-2')}>
                      <ChevronDown className="w-3.5 h-3.5" />
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                      <DropdownMenuItem onClick={() => setUploadOpen(true)}>
                        <UploadCloud className="w-4 h-4" />
                        Upload image
                      </DropdownMenuItem>
                      <DropdownMenuItem onClick={() => { setBatchInitialFiles([]); setBatchUploadOpen(true) }}>
                        <Images className="w-4 h-4" />
                        Upload multiple images
                      </DropdownMenuItem>
                    </DropdownMenuContent>
                  </DropdownMenu>
                </div>
              </div>
              <ImageGrid
                view={view}
                onImageSelect={(img) => { setAutoFocusTitle(false); setSelectedImage(img) }}
                onImageDeleted={(id) => { if (selectedImage?.id === id) { setSelectedImage(null); setAutoFocusTitle(false) } }}
                sortEndTrigger={sortEndTrigger}
              />
            </div>
          </ScrollArea>
        </main>
        {selectedImage && (
          <RightPanel
            image={selectedImage}
            onClose={() => { setSelectedImage(null); setAutoFocusTitle(false) }}
            autoFocusTitle={autoFocusTitle}
          />
        )}
        <UploadModal
          open={uploadOpen}
          onOpenChange={setUploadOpen}
          folderId={folderId}
          onUploadSuccess={checkVision}
        />
        <BatchUploadModal
          open={batchUploadOpen}
          onOpenChange={setBatchUploadOpen}
          folderId={folderId}
          initialFiles={batchInitialFiles}
        />
      </div>
      <DragOverlay dropAnimation={null} modifiers={[snapCenterToCursor]}>
        {activeDragImage && (
          <ImageDragOverlayCard thumbnailUrl={activeDragImage.thumbnailUrl} />
        )}
      </DragOverlay>
    </DndContext>
  )
}
