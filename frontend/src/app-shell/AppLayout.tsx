import { useState, useCallback, useEffect } from 'react'
import { Plus, UploadCloud, ChevronDown, Images, Focus } from 'lucide-react'
import { DndContext } from '@dnd-kit/core'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
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
import { Toggle } from '@/components/ui/toggle'
import FolderSidebar from '@/features/folder-sidebar/components/FolderSidebar'
import ImageGrid from '@/features/gallery/components/ImageGrid'
import GalleryToolbar from '@/features/gallery/components/GalleryToolbar'
import { useGalleryControls } from '@/features/gallery/hooks/useGalleryControls'
import ImageViewer from '@/features/viewer/components/ImageViewer'
import UploadModal from '@/features/upload/components/UploadModal'
import BatchUploadModal from '@/features/upload/components/BatchUploadModal'
import RightPanel from '@/features/right-panel/components/RightPanel'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { useVisionSuggestion } from '@/features/right-panel/hooks/useVisionSuggestion'
import { handleFileAutoUpload } from './lib/dragHandlers'
import type { Image } from '@/lib/images'
import { getTags } from '@/lib/tags'
import { useAppView } from './useAppView'
import { useAppDragAndDrop } from './useAppDragAndDrop'

export default function AppLayout() {
  const view = useAppView()
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [uploadOpen, setUploadOpen] = useState(false)
  const [batchUploadOpen, setBatchUploadOpen] = useState(false)
  const [batchInitialFiles, setBatchInitialFiles] = useState<File[]>([])
  const [selectedImage, setSelectedImage] = useState<Image | null>(null)
  const [folderPanelOpen, setFolderPanelOpen] = useState(false)
  const [focusMode, setFocusMode] = useState(false)
  const [isFileDragOver, setIsFileDragOver] = useState(false)
  const [isAutoUploading, setIsAutoUploading] = useState(false)
  const [autoFocusTitle, setAutoFocusTitle] = useState(false)
  const [viewerImage, setViewerImage] = useState<Image | null>(null)

  const folderId = view.type === 'folder' ? view.id : null
  const viewKey = view.type === 'folder' ? `folder:${view.id}` : view.type

  useEffect(() => {
    setViewerImage(null)
    setSelectedImage(null)
    setAutoFocusTitle(false)
  }, [viewKey])

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  const { data: tags = [] } = useQuery({
    queryKey: ['tags'],
    queryFn: () => getTags(getToken),
    staleTime: 60_000,
  })

  const gallery = useGalleryControls(view, tags, folders)
  const { sensors, handleDragStart, handleDragEnd, sortEndTrigger, dragOverlay } = useAppDragAndDrop(folders)

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
        {!focusMode && (
          <FolderSidebar
            view={view}
            onFolderSelect={() => { setFolderPanelOpen(true); setSelectedImage(null); setAutoFocusTitle(false) }}
          />
        )}
        <main
          className={cn('flex-1 h-screen min-w-0 relative', focusMode ? 'ml-0' : 'ml-[240px]')}
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
            <ImageViewer
              image={viewerImage}
              onClose={() => setViewerImage(null)}
              focusMode={focusMode}
              onToggleFocusMode={() => setFocusMode((v) => !v)}
            />
          ) : (
            <ScrollArea className="h-full">
              <div className="p-6">
                <GalleryToolbar
                  view={view}
                  controls={gallery}
                  focusToggle={
                    <Toggle
                      aria-label="Focus mode"
                      aria-pressed={focusMode}
                      pressed={focusMode}
                      onPressedChange={setFocusMode}
                      className="aria-pressed:bg-secondary aria-pressed:text-secondary-foreground"
                    >
                      <Focus className="w-3.5 h-3.5" />
                    </Toggle>
                  }
                  uploadActions={
                    <>
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
                    </>
                  }
                />
                <ImageGrid
                  view={view}
                  searchTerm={gallery.searchTerm}
                  debouncedSearchTerm={gallery.debouncedSearchTerm}
                  sortBy={gallery.sortBy}
                  sortDir={gallery.sortDir}
                  filterTagIds={gallery.filterTagIds}
                  filterMimeTypes={gallery.filterMimeTypes}
                  filterFolderIds={gallery.filterFolderIds}
                  onImageSelect={(img) => { setAutoFocusTitle(false); setSelectedImage(img); setFolderPanelOpen(false) }}
                  onImageDoubleClick={handleImageDoubleClick}
                  onImageDeleted={handleImageDeleted}
                  sortEndTrigger={sortEndTrigger}
                />
              </div>
            </ScrollArea>
          )}
        </main>
        {selectedImage && !focusMode ? (
          <RightPanel
            mode="image"
            image={selectedImage}
            onClose={() => { setSelectedImage(null); setAutoFocusTitle(false) }}
            autoFocusTitle={autoFocusTitle}
          />
        ) : folderPanelOpen && activeFolder && !focusMode ? (
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
      {dragOverlay}
    </DndContext>
  )
}
