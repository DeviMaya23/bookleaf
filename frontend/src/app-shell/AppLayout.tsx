import { useState, useCallback, useEffect } from 'react'
import { Plus, UploadCloud, ChevronDown, Images, Focus, MousePointerClick } from 'lucide-react'
import { DndContext } from '@dnd-kit/core'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
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
import ImageLightbox from '@/features/viewer/components/ImageLightbox'
import UploadModal from '@/features/upload/components/UploadModal'
import BatchUploadModal from '@/features/upload/components/BatchUploadModal'
import RightPanel from '@/features/right-panel/components/RightPanel'
import MobileTopBar from './components/MobileTopBar'
import FloatingUploadButton from './components/FloatingUploadButton'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { useMaintenanceActive } from '@/lib/maintenanceStore'
import MaintenancePage from '@/components/MaintenancePage'
import { useVisionSuggestion } from './useVisionSuggestion'
import { useSSEEvents } from './useSSEEvents'
import { getMe } from '@/features/auth/lib/me'
import { handleFileAutoUpload } from './lib/dragHandlers'
import { bulkAddImagesToFolder, bulkTrashImages } from '@/lib/images'
import type { Image } from '@/lib/images'
import { getTags } from '@/lib/tags'
import { useAppView } from './useAppView'
import { useAppDragAndDrop } from './useAppDragAndDrop'
import { useIsCoarsePointer } from '@/hooks/useIsCoarsePointer'

export default function AppLayout() {
  const maintenanceActive = useMaintenanceActive()
  const view = useAppView()
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [uploadOpen, setUploadOpen] = useState(false)
  const [uploadInitialFile, setUploadInitialFile] = useState<File | null>(null)
  const [batchUploadOpen, setBatchUploadOpen] = useState(false)
  const [batchInitialFiles, setBatchInitialFiles] = useState<File[]>([])
  const [selectedImage, setSelectedImage] = useState<Image | null>(null)
  const [folderPanelOpen, setFolderPanelOpen] = useState(false)
  const [focusMode, setFocusMode] = useState(false)
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false)
  const [isFileDragOver, setIsFileDragOver] = useState(false)
  const [isAutoUploading, setIsAutoUploading] = useState(false)
  const [autoFocusTitle, setAutoFocusTitle] = useState(false)
  const [viewerImage, setViewerImage] = useState<Image | null>(null)
  const [lightboxImage, setLightboxImage] = useState<Image | null>(null)
  const [selectMode, setSelectMode] = useState(false)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [mainSelectedId, setMainSelectedId] = useState<string | null>(null)
  const isCoarsePointer = useIsCoarsePointer()

  const folderId = view.type === 'folder' ? view.id : null
  const viewKey = view.type === 'folder' ? `folder:${view.id}` : view.type

  useEffect(() => {
    setViewerImage(null)
    setSelectedImage(null)
    setAutoFocusTitle(false)
    setLightboxImage(null)
    setSelectMode(false)
    setSelectedIds(new Set())
    setMainSelectedId(null)
  }, [viewKey])

  const handleSelectModeToggle = useCallback((pressed: boolean) => {
    setSelectMode(pressed)
    if (pressed) {
      setSelectedImage(null)
      setFolderPanelOpen(false)
      setAutoFocusTitle(false)
    } else {
      setSelectedIds(new Set())
      setMainSelectedId(null)
    }
  }, [])

  const exitSelectMode = useCallback(() => {
    setSelectMode(false)
    setSelectedIds(new Set())
    setMainSelectedId(null)
  }, [])

  const bulkAddToFolderMutation = useMutation({
    mutationFn: ({ imageIds, folderId }: { imageIds: string[]; folderId: string }) =>
      bulkAddImagesToFolder(getToken, imageIds, folderId),
    onSuccess: (result, { imageIds }) => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      if (result.succeeded_count < imageIds.length) {
        toast.warning(`Added ${result.succeeded_count} of ${imageIds.length} images to folder`)
      } else {
        toast.success(`Added ${result.succeeded_count} image${result.succeeded_count === 1 ? '' : 's'} to folder`)
      }
      exitSelectMode()
    },
    onError: () => {
      toast.error('Failed to add images to folder')
    },
  })

  const bulkTrashMutation = useMutation({
    mutationFn: (imageIds: string[]) => bulkTrashImages(getToken, imageIds),
    onSuccess: (result, imageIds) => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      if (result.succeeded_count < imageIds.length) {
        toast.warning(`Moved ${result.succeeded_count} of ${imageIds.length} images to trash`)
      } else {
        toast.success(`Moved ${result.succeeded_count} image${result.succeeded_count === 1 ? '' : 's'} to trash`)
      }
      exitSelectMode()
    },
    onError: () => {
      toast.error('Failed to move images to trash')
    },
  })

  const handleAddSelectionToFolder = useCallback((targetFolderId: string) => {
    bulkAddToFolderMutation.mutate({ imageIds: Array.from(selectedIds), folderId: targetFolderId })
  }, [bulkAddToFolderMutation, selectedIds])

  const handleMoveSelectionToTrash = useCallback(() => {
    bulkTrashMutation.mutate(Array.from(selectedIds))
  }, [bulkTrashMutation, selectedIds])

  const handleSelectionChange = useCallback((ids: Set<string>, anchorId: string | null) => {
    setSelectedIds(ids)
    setMainSelectedId(anchorId)
  }, [])

  useEffect(() => {
    function handlePaste(event: ClipboardEvent) {
      const tag = (event.target as HTMLElement).tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA') return
      if (uploadOpen) return
      const items = Array.from(event.clipboardData?.items ?? [])
      const imageItem = items.find((item) => item.kind === 'file' && item.type.startsWith('image/'))
      if (!imageItem) return
      const file = imageItem.getAsFile()
      if (!file) return
      setUploadInitialFile(file)
      setUploadOpen(true)
    }
    document.addEventListener('paste', handlePaste)
    return () => document.removeEventListener('paste', handlePaste)
  }, [uploadOpen])

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

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  const gallery = useGalleryControls(view, tags, folders, me?.vision_enabled ?? false)
  const { sensors, handleDragStart, handleDragEnd, sortEndTrigger, dragOverlay } = useAppDragAndDrop(folders)

  const activeFolder: Folder | null = view.type === 'folder'
    ? folders.find((f) => f.id === view.id) ?? null
    : null

  const { checkVision } = useVisionSuggestion()
  useSSEEvents()

  const handleImageSelect = useCallback((img: Image) => {
    if (isCoarsePointer) {
      setLightboxImage(img)
      return
    }
    setAutoFocusTitle(false)
    setSelectedImage(img)
    setFolderPanelOpen(false)
  }, [isCoarsePointer])

  const handleViewDetails = useCallback((img: Image) => {
    setAutoFocusTitle(false)
    setSelectedImage(img)
    setFolderPanelOpen(false)
  }, [])

  const handleFolderViewDetails = useCallback(() => {
    setFolderPanelOpen(true)
    setSelectedImage(null)
    setAutoFocusTitle(false)
  }, [])

  const handleImageDoubleClick = useCallback((img: Image) => {
    if (isCoarsePointer) return
    setSelectedImage(img)
    setFolderPanelOpen(false)
    setViewerImage(img)
  }, [isCoarsePointer])

  const handleImageDeleted = useCallback((id: string) => {
    if (selectedImage?.id === id) { setSelectedImage(null); setAutoFocusTitle(false) }
    if (viewerImage?.id === id) setViewerImage(null)
    if (lightboxImage?.id === id) setLightboxImage(null)
  }, [selectedImage, viewerImage, lightboxImage])

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
      const { image: imageDetail, duplicates } = await handleFileAutoUpload(getToken, file, folderId)
      queryClient.invalidateQueries({ queryKey: ['images'] })
      if (duplicates.length > 0) {
        toast.warning(`Possible duplicate of "${duplicates[0].title}"`)
      }
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

  if (maintenanceActive) {
    return <MaintenancePage />
  }

  return (
    <DndContext sensors={sensors} onDragStart={handleDragStart} onDragEnd={handleDragEnd}>
      <div className="flex h-screen">
        {!focusMode && (
          <>
            <MobileTopBar onMenuClick={() => setMobileDrawerOpen(true)} />
            {mobileDrawerOpen && (
              <div
                data-testid="mobile-drawer-backdrop"
                className="sm:hidden fixed inset-0 z-[25] bg-black/35"
                onClick={() => setMobileDrawerOpen(false)}
              />
            )}
            <FolderSidebar
              view={view}
              onFolderSelect={() => {
                if (!isCoarsePointer) setFolderPanelOpen(true)
                setSelectedImage(null)
                setAutoFocusTitle(false)
              }}
              onFolderViewDetails={handleFolderViewDetails}
              mobileOpen={mobileDrawerOpen}
              onMobileClose={() => setMobileDrawerOpen(false)}
            />
          </>
        )}
        <main
          className={cn(
            'flex-1 h-screen min-w-0 relative pt-12 sm:pt-0',
            focusMode ? 'ml-0' : 'ml-0 sm:ml-[240px]',
          )}
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
          {lightboxImage !== null ? (
            <ImageLightbox
              image={lightboxImage}
              onClose={() => setLightboxImage(null)}
            />
          ) : viewerImage !== null ? (
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
                  controlsDisabled={selectMode}
                  focusToggle={
                    <div className="hidden sm:flex">
                      <Toggle
                        aria-label="Focus mode"
                        aria-pressed={focusMode}
                        pressed={focusMode}
                        onPressedChange={setFocusMode}
                        className="aria-pressed:bg-secondary aria-pressed:text-secondary-foreground"
                      >
                        <Focus className="w-3.5 h-3.5" />
                      </Toggle>
                    </div>
                  }
                  selectModeToggle={
                    !isCoarsePointer && view.type !== 'trash' ? (
                      <Toggle
                        aria-label="Select mode"
                        aria-pressed={selectMode}
                        pressed={selectMode}
                        onPressedChange={handleSelectModeToggle}
                        className={cn(buttonVariants({ variant: selectMode ? 'secondary' : 'outline', size: 'icon' }))}
                      >
                        <MousePointerClick className="w-3.5 h-3.5" />
                      </Toggle>
                    ) : null
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
                  searchLabels={gallery.searchLabels}
                  onImageSelect={handleImageSelect}
                  onImageDoubleClick={handleImageDoubleClick}
                  onImageDeleted={handleImageDeleted}
                  onViewDetails={handleViewDetails}
                  sortEndTrigger={sortEndTrigger}
                  selectMode={selectMode}
                  selectedIds={selectedIds}
                  mainSelectedId={mainSelectedId}
                  onSelectionChange={handleSelectionChange}
                />
              </div>
            </ScrollArea>
          )}
        </main>
        {selectedIds.size > 0 ? (
          <RightPanel
            mode="selection"
            selectedCount={selectedIds.size}
            onAddToFolder={handleAddSelectionToFolder}
            onMoveToTrash={handleMoveSelectionToTrash}
            onClose={exitSelectMode}
          />
        ) : selectedImage && !focusMode ? (
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
        <FloatingUploadButton onClick={() => setUploadOpen(true)} />
        <UploadModal
          open={uploadOpen}
          onOpenChange={(open) => {
            setUploadOpen(open)
            if (!open) setUploadInitialFile(null)
          }}
          folderId={folderId}
          initialFile={uploadInitialFile ?? undefined}
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
