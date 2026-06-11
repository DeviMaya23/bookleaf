import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useDndContext } from '@dnd-kit/core'
import { Button } from '@/components/ui/button'
import { PlusIcon } from 'lucide-react'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import type { AppView } from '@/lib/view'
import FolderNameDialog from './FolderNameDialog'
import DeleteFolderDialog from './DeleteFolderDialog'
import FolderItem from './FolderItem'
import UnsortedEntry from './UnsortedEntry'
import TrashEntry from './TrashEntry'
import RootDropZone from './RootDropZone'
import ProfileMenu from '@/features/auth/components/ProfileMenu'
import { buildFolderTree, filterFolderTree, type FolderNode } from '../lib/folderTree'
import { useFolderMutations } from '../hooks/useFolderMutations'

type NameDialogState =
  | { mode: 'create-root' }
  | { mode: 'create-sub'; parent: Folder }
  | { mode: 'rename'; target: Folder }
  | null

interface FolderSidebarProps {
  view: AppView
  onFolderSelect?: () => void
}

export default function FolderSidebar({ view, onFolderSelect }: FolderSidebarProps) {
  const { getToken } = useKindeAuth()
  const navigate = useNavigate()
  const { active } = useDndContext()

  const activeDragType = (active?.data.current?.type as string) ?? null
  const activeDragFolderId = activeDragType === 'folder'
    ? (active?.data.current?.folderId as string)
    : null

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
  })

  const { createMutation, renameMutation, deleteMutation } = useFolderMutations()

  const [folderFilter, setFolderFilter] = useState('')
  const [nameDialog, setNameDialog] = useState<NameDialogState>(null)
  const [deleteTarget, setDeleteTarget] = useState<Folder | null>(null)

  function handleDelete() {
    if (!deleteTarget) return
    deleteMutation.mutate(deleteTarget.id)
    setDeleteTarget(null)
  }

  function handleFolderSelect(folder: FolderNode) {
    const isActive = view.type === 'folder' && view.id === folder.id
    if (!isActive) onFolderSelect?.()
    navigate(`/folders/${folder.id}`)
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
    <aside className="fixed inset-y-0 left-0 w-[240px] flex flex-col border-r bg-background">
      <div className="p-4 pb-3">
        <span className="text-sm font-semibold tracking-tight">Bookleaf</span>
      </div>

      <nav className="flex-1 overflow-y-auto px-2 py-1 space-y-0.5">
        <div
          className={`px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
            view.type === 'all'
              ? 'bg-accent text-accent-foreground font-medium'
              : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
          }`}
          onClick={() => navigate('/')}
        >
          All
        </div>
        <UnsortedEntry
          active={view.type === 'unsorted'}
          activeDragType={activeDragType}
          onClick={() => navigate('/unsorted')}
        />
        <TrashEntry active={view.type === 'trash'} onClick={() => navigate('/trash')} />

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
            onSelect={handleFolderSelect}
            onRename={(target) => setNameDialog({ mode: 'rename', target })}
            onDelete={setDeleteTarget}
            onNewSubfolder={(parent) => setNameDialog({ mode: 'create-sub', parent })}
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
