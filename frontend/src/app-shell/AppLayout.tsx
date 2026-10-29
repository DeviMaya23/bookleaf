import { useState, useEffect } from 'react'
import { Plus, UploadCloud, ChevronDown, Images, Focus, MousePointerClick } from 'lucide-react'
import { DndContext } from '@dnd-kit/core'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useLocation, useNavigate as useRouterNavigate } from 'react-router-dom'
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
import RightPanel, { type PanelContent } from '@/features/right-panel/components/RightPanel'
import MobileTopBar from './components/MobileTopBar'
import FloatingUploadButton from './components/FloatingUploadButton'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { useMaintenanceActive } from '@/lib/maintenanceStore'
import MaintenancePage from '@/components/MaintenancePage'
import { useSSEEvents } from './useSSEEvents'
import { getMe } from '@/features/auth/lib/me'
import { getTags } from '@/lib/tags'
import { useAppView } from './useAppView'
import TagsView from '@/features/tags/components/TagsView'
import { useAppDragAndDrop } from './useAppDragAndDrop'
import { useIsCoarsePointer } from '@/hooks/useIsCoarsePointer'
import { useViewerState } from './useViewerState'
import { useSelectionMode } from './useSelectionMode'
import { useUploadTriggers } from './useUploadTriggers'
import { useFileDrop } from './useFileDrop'

export default function AppLayout() {
  const maintenanceActive = useMaintenanceActive()
  const view = useAppView()
  const { getToken } = useKindeAuth()
  const location = useLocation()
  const routerNavigate = useRouterNavigate()

  const [focusMode, setFocusMode] = useState(false)
  const [mobileDrawerOpen, setMobileDrawerOpen] = useState(false)
  const [coarsePointerPanelOpen, setCoarsePointerPanelOpen] = useState(false)
  const isCoarsePointer = useIsCoarsePointer()

  const folderId = view.type === 'folder' ? view.id : null
  const viewKey = view.type === 'folder' ? `folder:${view.id}` : view.type

  useEffect(() => {
    setCoarsePointerPanelOpen(false)
  }, [view.type])

  const viewer = useViewerState({ isCoarsePointer, viewKey })
  const upload = useUploadTriggers()
  const selection = useSelectionMode({ viewKey, clearImageSelection: viewer.clearImageSelection })
  const fileDrop = useFileDrop({
    folderId,
    onAutoUploaded: viewer.selectImageAutoFocus,
    onBatchFilesDropped: upload.openBatchUpload,
  })

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

  useEffect(() => {
    const state = location.state as Record<string, unknown> | null
    const tagIds = Array.isArray(state?.filterTagIds) ? (state.filterTagIds as string[]) : null
    if (tagIds && tagIds.length > 0) {
      gallery.setFilterTagIds(tagIds)
      routerNavigate(location.pathname, { replace: true, state: null })
    }
  // gallery.setFilterTagIds and routerNavigate are stable refs; firing only on
  // location.state change prevents the filter from re-applying on every render.
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [location.state])

  const { sensors, handleDragStart, handleDragEnd, sortEndTrigger, dragOverlay } = useAppDragAndDrop(folders)

  const activeFolder: Folder | null = view.type === 'folder'
    ? folders.find((f) => f.id === view.id) ?? null
    : null

  const viewLabel = view.type === 'unsorted'
    ? 'Unsorted'
    : view.type === 'trash'
      ? 'Trash'
      : view.type === 'folder'
        ? (activeFolder?.name ?? '')
        : 'All'

  const panelContent: PanelContent = (selection.selectMode || selection.selectedIds.size > 0)
    ? { mode: 'selection', selectedCount: selection.selectedIds.size, onApply: selection.handleApplySelection, onMoveToTrash: selection.handleMoveSelectionToTrash, onExitSelectMode: selection.exitSelectMode, onDownloadZip: selection.handleDownloadSelection }
    : viewer.selectedImage
      ? { mode: 'image', image: viewer.selectedImage, autoFocusTitle: viewer.autoFocusTitle }
      : activeFolder
        ? { mode: 'folder', folder: activeFolder }
        : { mode: 'neutral', viewLabel }

  useSSEEvents()

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
              onFolderSelect={viewer.clearImageSelection}
              onFolderViewDetails={() => {
                viewer.clearImageSelection()
                setCoarsePointerPanelOpen(true)
              }}
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
          onDragOver={fileDrop.dragHandlers.onDragOver}
          onDragLeave={fileDrop.dragHandlers.onDragLeave}
          onDrop={fileDrop.dragHandlers.onDrop}
        >
          {(fileDrop.isFileDragOver || fileDrop.isAutoUploading) && (
            <div className="absolute inset-0 z-50 flex flex-col items-center justify-center bg-background/80 backdrop-blur-sm border-2 border-dashed border-primary rounded-none pointer-events-none">
              <UploadCloud className="w-12 h-12 text-primary mb-3" />
              <p className="text-sm font-medium text-primary">
                {fileDrop.isAutoUploading ? 'Uploading…' : 'Drop to upload'}
              </p>
            </div>
          )}
          {viewer.lightboxImage !== null ? (
            <ImageLightbox
              image={viewer.lightboxImage}
              onClose={viewer.closeLightbox}
            />
          ) : viewer.viewerImage !== null ? (
            <ImageViewer
              image={viewer.viewerImage}
              onClose={viewer.closeViewer}
              focusMode={focusMode}
              onToggleFocusMode={() => setFocusMode((v) => !v)}
            />
          ) : (
            <ScrollArea className="h-full">
              <div className="p-6">
                {view.type === 'tags' ? <TagsView /> : null}
                {view.type !== 'tags' && <GalleryToolbar
                  view={view}
                  controls={gallery}
                  controlsDisabled={selection.selectMode}
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
                        aria-pressed={selection.selectMode}
                        pressed={selection.selectMode}
                        onPressedChange={selection.handleSelectModeToggle}
                        className={cn(buttonVariants({ variant: selection.selectMode ? 'secondary' : 'outline', size: 'icon' }))}
                      >
                        <MousePointerClick className="w-3.5 h-3.5" />
                      </Toggle>
                    ) : null
                  }
                  uploadActions={
                    <>
                      <button
                        className={cn(buttonVariants(), 'rounded-r-none')}
                        onClick={() => upload.openUpload()}
                      >
                        <Plus className="w-4 h-4 mr-1" />
                        Image
                      </button>
                      <DropdownMenu>
                        <DropdownMenuTrigger className={cn(buttonVariants(), 'rounded-l-none border-l border-l-white/20 px-2')}>
                          <ChevronDown className="w-3.5 h-3.5" />
                        </DropdownMenuTrigger>
                        <DropdownMenuContent align="end">
                          <DropdownMenuItem onClick={() => upload.openUpload()}>
                            <UploadCloud className="w-4 h-4" />
                            Upload image
                          </DropdownMenuItem>
                          <DropdownMenuItem onClick={() => upload.openBatchUpload()}>
                            <Images className="w-4 h-4" />
                            Upload multiple images
                          </DropdownMenuItem>
                        </DropdownMenuContent>
                      </DropdownMenu>
                    </>
                  }
                />}
                {view.type !== 'tags' && <ImageGrid
                  view={view}
                  searchTerm={gallery.searchTerm}
                  debouncedSearchTerm={gallery.debouncedSearchTerm}
                  sortBy={gallery.sortBy}
                  sortDir={gallery.sortDir}
                  filterTagIds={gallery.filterTagIds}
                  filterMimeTypes={gallery.filterMimeTypes}
                  filterFolderIds={gallery.filterFolderIds}
                  searchLabels={gallery.searchLabels}
                  onImageSelect={viewer.handleImageSelect}
                  onImageDoubleClick={viewer.handleImageDoubleClick}
                  onImageDeleted={viewer.handleImageDeleted}
                  onViewDetails={viewer.handleViewDetails}
                  sortEndTrigger={sortEndTrigger}
                  selectMode={selection.selectMode}
                  selectedIds={selection.selectedIds}
                  mainSelectedId={selection.mainSelectedId}
                  onSelectionChange={selection.handleSelectionChange}
                />}
              </div>
            </ScrollArea>
          )}
        </main>
        <RightPanel panelContent={panelContent} focusMode={focusMode} mobileOpen={coarsePointerPanelOpen} onMobileClose={() => setCoarsePointerPanelOpen(false)} />
        <FloatingUploadButton onClick={() => upload.openUpload()} />
        <UploadModal
          open={upload.uploadOpen}
          onOpenChange={(open) => { if (open) upload.openUpload(); else upload.closeUpload() }}
          folderId={folderId}
          initialFile={upload.uploadInitialFile ?? undefined}
        />
        <BatchUploadModal
          open={upload.batchUploadOpen}
          onOpenChange={(open) => { if (!open) upload.closeBatchUpload() }}
          folderId={folderId}
          initialFiles={upload.batchInitialFiles}
        />
      </div>
      {dragOverlay}
    </DndContext>
  )
}
