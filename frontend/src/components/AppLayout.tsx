import { useState, useCallback, useEffect, useRef } from 'react'
import { useParams, useLocation } from 'react-router-dom'
import { Plus, UploadCloud, ChevronDown, Images, Search, ArrowUpDown, ArrowUp, ArrowDown } from 'lucide-react'
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
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import FolderSidebar from './FolderSidebar'
import ImageGrid from './ImageGrid'
import ImageViewer from './ImageViewer'
import UploadModal from './UploadModal'
import BatchUploadModal from './BatchUploadModal'
import RightPanel from './RightPanel'
import { useQuery } from '@tanstack/react-query'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { useVisionSuggestion } from '@/hooks/useVisionSuggestion'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { handleImageDrop, handleFolderDrop, handleFileAutoUpload } from '@/lib/dragHandlers'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import type { SortEndTrigger } from '@/components/ImageGrid'

type SortBy = 'manual' | 'created_at' | 'title'
type SortDir = 'asc' | 'desc'

const SORT_FIELD_LABELS: Record<SortBy, string> = {
  manual: 'Manual',
  created_at: 'Date added',
  title: 'Name',
}

const FIELD_DEFAULT_DIRECTION: Record<'created_at' | 'title', SortDir> = {
  created_at: 'desc',
  title: 'asc',
}

const DIR_LABELS: Record<'created_at' | 'title', Record<SortDir, string>> = {
  created_at: { asc: 'Oldest first', desc: 'Newest first' },
  title: { asc: 'A → Z', desc: 'Z → A' },
}

function defaultSortForViewType(isFolder: boolean): { sortBy: SortBy; sortDir: SortDir | undefined } {
  return isFolder ? { sortBy: 'manual', sortDir: undefined } : { sortBy: 'created_at', sortDir: 'desc' }
}

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
  const [folderPanelOpen, setFolderPanelOpen] = useState(false)
  const [isFileDragOver, setIsFileDragOver] = useState(false)
  const [isAutoUploading, setIsAutoUploading] = useState(false)
  const [activeDragImage, setActiveDragImage] = useState<{ id: string; thumbnailUrl: string | null } | null>(null)
  const sortEndTriggerRef = useRef<number>(0)
  const [sortEndTrigger, setSortEndTrigger] = useState<SortEndTrigger | null>(null)
  const [autoFocusTitle, setAutoFocusTitle] = useState(false)
  const [viewerImage, setViewerImage] = useState<Image | null>(null)
  const [searchTerm, setSearchTerm] = useState('')
  const debouncedSearchTerm = useDebouncedValue(searchTerm, 300)
  const [sortBy, setSortBy] = useState<SortBy>(() => defaultSortForViewType(view.type === 'folder').sortBy)
  const [sortDir, setSortDir] = useState<SortDir | undefined>(() => defaultSortForViewType(view.type === 'folder').sortDir)

  const folderId = view.type === 'folder' ? view.id : null
  const viewKey = view.type === 'folder' ? `folder:${view.id}` : view.type
  const viewDefaultSort = defaultSortForViewType(view.type === 'folder')
  const sortFieldOptions: SortBy[] = view.type === 'folder' ? ['manual', 'created_at', 'title'] : ['created_at', 'title']
  const sortActive = sortBy !== viewDefaultSort.sortBy
    || (sortBy !== 'manual' && sortDir !== FIELD_DEFAULT_DIRECTION[sortBy])

  useEffect(() => {
    setViewerImage(null)
    setSelectedImage(null)
    setAutoFocusTitle(false)
  }, [viewKey])

  useEffect(() => {
    setSearchTerm('')
  }, [viewKey])

  useEffect(() => {
    const def = defaultSortForViewType(viewKey.startsWith('folder:'))
    setSortBy(def.sortBy)
    setSortDir(def.sortDir)
  }, [viewKey])

  const handleSortFieldChange = useCallback((value: SortBy) => {
    setSortBy(value)
    setSortDir(value === 'manual' ? undefined : FIELD_DEFAULT_DIRECTION[value])
  }, [])

  const handleSortDirToggle = useCallback(() => {
    setSortDir((prev) => (prev === 'asc' ? 'desc' : 'asc'))
  }, [])

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  const activeFolder: Folder | null = view.type === 'folder'
    ? folders.find((f) => f.id === view.id) ?? null
    : null

  const { checkVision } = useVisionSuggestion()

  const handleImageDoubleClick = useCallback((img: Image) => {
    setSelectedImage(img)
    setFolderPanelOpen(false)
    setViewerImage(img)
  }, [])

  const handleImageDeleted = useCallback((id: string) => {
    if (selectedImage?.id === id) { setSelectedImage(null); setAutoFocusTitle(false) }
    if (viewerImage?.id === id) setViewerImage(null)
  }, [selectedImage, viewerImage])

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
      setFolderPanelOpen(false)
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
        <FolderSidebar
          view={view}
          onFolderSelect={() => { setFolderPanelOpen(true); setSelectedImage(null); setAutoFocusTitle(false) }}
        />
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
          {viewerImage !== null ? (
            <ImageViewer image={viewerImage} onClose={() => setViewerImage(null)} />
          ) : (
            <ScrollArea className="h-full">
              <div className="p-6">
                <div className="flex items-center justify-between gap-2 mb-4">
                  <div className="flex items-center gap-2 w-full max-w-xs">
                    <div className="relative flex-1">
                      <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
                      <input
                        value={searchTerm}
                        onChange={(e) => setSearchTerm(e.target.value)}
                        placeholder="Search images by name…"
                        className="w-full rounded-md border bg-background pl-8 pr-3 py-1.5 text-sm outline-none focus:ring-1 focus:ring-primary/40"
                      />
                    </div>
                    <DropdownMenu>
                      <DropdownMenuTrigger
                        aria-label="Sort"
                        className={cn(buttonVariants({ variant: sortActive ? 'default' : 'outline', size: 'icon' }))}
                      >
                        <ArrowUpDown className="w-3.5 h-3.5" />
                      </DropdownMenuTrigger>
                      <DropdownMenuContent align="start" className="w-48">
                        <DropdownMenuRadioGroup
                          value={sortBy}
                          onValueChange={(value) => handleSortFieldChange(value as SortBy)}
                        >
                          {sortFieldOptions.map((field) => (
                            <DropdownMenuRadioItem key={field} value={field}>
                              {SORT_FIELD_LABELS[field]}
                            </DropdownMenuRadioItem>
                          ))}
                        </DropdownMenuRadioGroup>
                        {sortBy !== 'manual' && (
                          <>
                            <DropdownMenuSeparator />
                            <DropdownMenuItem closeOnClick={false} onClick={handleSortDirToggle}>
                              {sortDir === 'asc' ? <ArrowUp className="w-4 h-4" /> : <ArrowDown className="w-4 h-4" />}
                              {DIR_LABELS[sortBy][sortDir ?? FIELD_DEFAULT_DIRECTION[sortBy]]}
                            </DropdownMenuItem>
                          </>
                        )}
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </div>
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
                  searchTerm={searchTerm}
                  debouncedSearchTerm={debouncedSearchTerm}
                  sortBy={sortBy}
                  sortDir={sortDir}
                  onImageSelect={(img) => { setAutoFocusTitle(false); setSelectedImage(img); setFolderPanelOpen(false) }}
                  onImageDoubleClick={handleImageDoubleClick}
                  onImageDeleted={handleImageDeleted}
                  sortEndTrigger={sortEndTrigger}
                />
              </div>
            </ScrollArea>
          )}
        </main>
        {selectedImage ? (
          <RightPanel
            mode="image"
            image={selectedImage}
            onClose={() => { setSelectedImage(null); setAutoFocusTitle(false) }}
            autoFocusTitle={autoFocusTitle}
          />
        ) : folderPanelOpen && activeFolder ? (
          <RightPanel
            mode="folder"
            folder={activeFolder}
            onClose={() => setFolderPanelOpen(false)}
          />
        ) : null}
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
