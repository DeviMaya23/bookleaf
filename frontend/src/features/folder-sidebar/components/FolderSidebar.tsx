import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useDndContext } from '@dnd-kit/core'
import { Button } from '@/components/ui/button'
import { PlusIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import type { AppView } from '@/lib/view'
import { getMe } from '@/features/auth/lib/me'
import FolderNameDialog from './FolderNameDialog'
import DeleteFolderDialog from './DeleteFolderDialog'
import FolderItem from './FolderItem'
import UnsortedEntry from './UnsortedEntry'
import TrashEntry from './TrashEntry'
import RootDropZone from './RootDropZone'
import ProfileMenu from '@/features/auth/components/ProfileMenu'
import { buildFolderTree, filterFolderTree, type FolderNode } from '../lib/folderTree'
import { useFolderMutations } from '../hooks/useFolderMutations'
import { isImageDragData, isFolderDragData } from '@/app-shell/lib/dragHandlers'
import { FOLDER_ICONS, SYSTEM_ICON_KEYS } from '../lib/folderIcons'

const AllIcon = FOLDER_ICONS[SYSTEM_ICON_KEYS.all]

type NameDialogState =
  | { mode: 'create-root' }
  | { mode: 'create-sub'; parent: Folder }
  | { mode: 'rename'; target: Folder }
  | null

interface FolderSidebarProps {
  view: AppView
  onFolderSelect?: () => void
  onFolderViewDetails?: () => void
  mobileOpen?: boolean
  onMobileClose?: () => void
}

export default function FolderSidebar({ view, onFolderSelect, onFolderViewDetails, mobileOpen, onMobileClose }: FolderSidebarProps) {
  const { getToken } = useKindeAuth()
  const navigate = useNavigate()
  const { active } = useDndContext()

  const dragData = active?.data.current
  const activeDragType = isImageDragData(dragData) ? 'image' : isFolderDragData(dragData) ? 'folder' : null
  const activeDragFolderId = isFolderDragData(dragData) ? dragData.folderId : null

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
  })

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })
  const iconsEnabled = me?.folder_icons_enabled ?? true

  const { createMutation, renameMutation, deleteMutation, changeIconMutation } = useFolderMutations()

  const [folderFilter, setFolderFilter] = useState('')
  const [nameDialog, setNameDialog] = useState<NameDialogState>(null)
  const [deleteTarget, setDeleteTarget] = useState<Folder | null>(null)

  function handleDelete() {
    if (!deleteTarget) return
    const targetId = deleteTarget.id
    const isViewingTarget = view.type === 'folder' && view.id === targetId
    deleteMutation.mutate(targetId, {
      onSuccess: () => {
        if (isViewingTarget) navigate('/app')
      },
    })
    setDeleteTarget(null)
  }

  function handleFolderSelect(folder: FolderNode) {
    const isActive = view.type === 'folder' && view.id === folder.id
    if (!isActive) onFolderSelect?.()
    navigate(`/app/folders/${folder.id}`)
    onMobileClose?.()
  }

  function handleViewDetails(folder: Folder) {
    const isActive = view.type === 'folder' && view.id === folder.id
    if (!isActive) navigate(`/app/folders/${folder.id}`)
    onFolderViewDetails?.()
    onMobileClose?.()
  }

  function handleNameDialogSubmit(name: string) {
    if (!nameDialog) return
    if (nameDialog.mode === 'create-root') {
      createMutation.mutate({ name })
    } else if (nameDialog.mode === 'create-sub') {
      createMutation.mutate({ name, parentId: nameDialog.parent.id })
    } else {
      renameMutation.mutate({ id: nameDialog.target.id, name })
    }
  }

  const nameDialogTitle =
    nameDialog?.mode === 'create-sub'
      ? 'New subfolder'
      : nameDialog?.mode === 'rename'
      ? 'Rename folder'
      : 'New folder'
  const nameDialogInitialValue = nameDialog?.mode === 'rename' ? nameDialog.target.name : ''

  const tree = buildFolderTree(folders)
  const trimmedFolderFilter = folderFilter.trim()
  const visibleTree = trimmedFolderFilter ? filterFolderTree(tree, trimmedFolderFilter) : tree

  return (
    <aside
      className={cn(
        'fixed inset-y-0 left-0 w-[240px] flex flex-col border-r bg-background z-30 transform transition-transform duration-200 sm:translate-x-0',
        mobileOpen ? 'translate-x-0' : '-translate-x-full',
      )}
    >
      <div className="p-4 pb-3">
        <span className="text-sm font-semibold tracking-tight font-serif">Bookleaf</span>
      </div>

      <nav className="flex-1 overflow-y-auto px-2 py-1 space-y-0.5">
        <div
          className={`flex items-center gap-1.5 px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
            view.type === 'all'
              ? 'bg-accent text-accent-foreground font-medium'
              : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
          }`}
          onClick={() => { navigate('/app'); onMobileClose?.() }}
        >
          {iconsEnabled && <AllIcon className="w-3.5 h-3.5 shrink-0" />}
          All
        </div>
        <UnsortedEntry
          active={view.type === 'unsorted'}
          activeDragType={activeDragType}
          iconsEnabled={iconsEnabled}
          onClick={() => { navigate('/app/unsorted'); onMobileClose?.() }}
        />
        <TrashEntry active={view.type === 'trash'} iconsEnabled={iconsEnabled} onClick={() => { navigate('/app/trash'); onMobileClose?.() }} />

        <div className="pt-2 pb-1">
          <div className="border-t mb-2" />
          <div className="flex items-center justify-between px-3">
            <p className="text-[10px] font-semibold tracking-widest uppercase text-muted-foreground/50">
              Folders
            </p>
            <Button
              variant="ghost"
              size="icon-xs"
              aria-label="New folder"
              onClick={() => setNameDialog({ mode: 'create-root' })}
            >
              <PlusIcon />
            </Button>
          </div>
        </div>

        {visibleTree.map((folder) => (
          <FolderItem
            key={folder.id}
            folder={folder}
            depth={0}
            view={view}
            folders={folders}
            activeDragType={activeDragType}
            activeDragFolderId={activeDragFolderId}
            iconsEnabled={iconsEnabled}
            onSelect={handleFolderSelect}
            onRename={(target) => setNameDialog({ mode: 'rename', target })}
            onDelete={setDeleteTarget}
            onNewSubfolder={(parent) => setNameDialog({ mode: 'create-sub', parent })}
            onChangeIcon={(target, icon) => changeIconMutation.mutate({ id: target.id, icon })}
            onViewDetails={handleViewDetails}
          />
        ))}

        <RootDropZone />
      </nav>

      <div className="p-2 border-t space-y-1">
        {folders.length === 0 && (
          <button
            className="w-full rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground text-left"
            onClick={() => setNameDialog({ mode: 'create-root' })}
          >
            + New folder
          </button>
        )}
        <input
          value={folderFilter}
          onChange={(e) => setFolderFilter(e.target.value)}
          placeholder="Filter folders…"
          className="w-full rounded-md border bg-background px-2.5 py-1 text-xs outline-none focus:ring-1 focus:ring-primary/40"
        />
        <ProfileMenu />
      </div>

      <FolderNameDialog
        open={!!nameDialog}
        onOpenChange={(open) => { if (!open) setNameDialog(null) }}
        title={nameDialogTitle}
        initialValue={nameDialogInitialValue}
        onSubmit={handleNameDialogSubmit}
      />

      <DeleteFolderDialog
        folder={deleteTarget}
        onCancel={() => setDeleteTarget(null)}
        onConfirm={handleDelete}
      />
    </aside>
  )
}
