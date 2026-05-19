import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import FolderNameDialog from './FolderNameDialog'
import ProfileMenu from './ProfileMenu'
import { getFolders, createFolder, renameFolder, deleteFolder } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import type { AppView } from '@/lib/view'

interface FolderNode extends Folder {
  children: FolderNode[]
}

function buildFolderTree(folders: Folder[]): FolderNode[] {
  const nodeMap = new Map<string, FolderNode>()
  const visited = new Set<string>()
  const roots: FolderNode[] = []

  for (const f of folders) {
    nodeMap.set(f.id, { ...f, children: [] })
  }

  for (const f of folders) {
    const node = nodeMap.get(f.id)!
    if (f.parent_id && nodeMap.has(f.parent_id) && !visited.has(f.id)) {
      visited.add(f.id)
      nodeMap.get(f.parent_id)!.children.push(node)
    } else if (!f.parent_id) {
      roots.push(node)
    }
  }

  return roots
}

interface FolderItemProps {
  folder: FolderNode
  depth: number
  view: AppView
  onSelect: (folder: FolderNode) => void
  onRename: (folder: Folder) => void
  onDelete: (folder: Folder) => void
  onNewSubfolder: (folder: Folder) => void
}

function FolderItem({ folder, depth, view, onSelect, onRename, onDelete, onNewSubfolder }: FolderItemProps) {
  const [open, setOpen] = useState(depth === 0)
  const hasChildren = folder.children.length > 0
  const isActive = view.type === 'folder' && view.id === folder.id

  return (
    <div>
      <ContextMenu>
        <ContextMenuTrigger>
          <div
            style={{ paddingLeft: 8 + depth * 14 }}
            className={`flex items-center gap-1 pr-2 py-1 rounded-md cursor-pointer mb-0.5 text-sm select-none ${
              isActive
                ? 'bg-accent text-accent-foreground font-medium'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
            }`}
            onClick={() => onSelect(folder)}
          >
            <span
              onClick={(e) => { e.stopPropagation(); setOpen((o) => !o) }}
              className={`w-3.5 h-3.5 shrink-0 flex items-center justify-center text-muted-foreground/50 text-[9px] transition-transform ${
                hasChildren ? '' : 'invisible'
              } ${open ? 'rotate-90' : ''}`}
            >
              ▶
            </span>
            <span className="flex-1 truncate">{folder.name}</span>
          </div>
        </ContextMenuTrigger>
        <ContextMenuContent>
          <ContextMenuItem onClick={() => onNewSubfolder(folder)}>
            New subfolder
          </ContextMenuItem>
          <ContextMenuItem onClick={() => onRename(folder)}>
            Rename
          </ContextMenuItem>
          <ContextMenuItem
            onClick={() => onDelete(folder)}
            className="text-destructive focus:text-destructive"
          >
            Delete
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>

      {hasChildren && open && (
        <div>
          {folder.children.map((child) => (
            <FolderItem
              key={child.id}
              folder={child}
              depth={depth + 1}
              view={view}
              onSelect={onSelect}
              onRename={onRename}
              onDelete={onDelete}
              onNewSubfolder={onNewSubfolder}
            />
          ))}
        </div>
      )}
    </div>
  )
}

interface SystemEntryProps {
  label: string
  active: boolean
  muted?: boolean
  onClick: () => void
}

function SystemEntry({ label, active, muted, onClick }: SystemEntryProps) {
  return (
    <div
      className={`px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
        active
          ? 'bg-accent text-accent-foreground font-medium'
          : muted
          ? 'text-muted-foreground/60 hover:bg-accent hover:text-accent-foreground'
          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
      }`}
      onClick={onClick}
    >
      {label}
    </div>
  )
}

interface FolderSidebarProps {
  view: AppView
}

export default function FolderSidebar({ view }: FolderSidebarProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const navigate = useNavigate()

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
  })

  const [newFolderOpen, setNewFolderOpen] = useState(false)
  const [subfolderParent, setSubfolderParent] = useState<Folder | null>(null)
  const [renameTarget, setRenameTarget] = useState<Folder | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Folder | null>(null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['folders'] })

  const createMutation = useMutation({
    mutationFn: ({ name, parentId }: { name: string; parentId?: string }) =>
      createFolder(getToken, name, parentId),
    onSuccess: invalidate,
  })

  const renameMutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => renameFolder(getToken, id, name),
    onSuccess: invalidate,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteFolder(getToken, id),
    onSuccess: () => {
      invalidate()
      toast.success('Folder deleted')
    },
    onError: () => {
      toast.error('Failed to delete folder')
    },
  })

  function handleDelete() {
    if (!deleteTarget) return
    deleteMutation.mutate(deleteTarget.id)
    setDeleteTarget(null)
  }

  function handleFolderSelect(folder: FolderNode) {
    navigate(`/folders/${folder.id}`)
  }

  const tree = buildFolderTree(folders)

  return (
    <aside className="fixed inset-y-0 left-0 w-[240px] flex flex-col border-r bg-background">
      <div className="p-4 pb-3">
        <span className="text-sm font-semibold tracking-tight">Bookleaf</span>
      </div>

      <nav className="flex-1 overflow-y-auto px-2 py-1 space-y-0.5">
        <SystemEntry
          label="All"
          active={view.type === 'all'}
          onClick={() => navigate('/')}
        />
        <SystemEntry
          label="Unsorted"
          active={view.type === 'unsorted'}
          onClick={() => navigate('/unsorted')}
        />
        <div className="mt-[8px]">
          <SystemEntry
            label="Trash"
            active={view.type === 'trash'}
            muted
            onClick={() => navigate('/trash')}
          />
        </div>

        <div className="pt-2 pb-1">
          <div className="border-t mb-2" />
          <p className="px-3 text-[10px] font-semibold tracking-widest uppercase text-muted-foreground/50">
            Folders
          </p>
        </div>

        {tree.map((folder) => (
          <FolderItem
            key={folder.id}
            folder={folder}
            depth={0}
            view={view}
            onSelect={handleFolderSelect}
            onRename={setRenameTarget}
            onDelete={setDeleteTarget}
            onNewSubfolder={setSubfolderParent}
          />
        ))}
      </nav>

      <div className="p-2 border-t space-y-1">
        <button
          className="w-full rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground text-left"
          onClick={() => setNewFolderOpen(true)}
        >
          + New folder
        </button>
        <ProfileMenu />
      </div>

      {/* New root folder */}
      <FolderNameDialog
        open={newFolderOpen}
        onOpenChange={setNewFolderOpen}
        title="New folder"
        onSubmit={(name) => createMutation.mutate({ name })}
      />

      {/* New subfolder */}
      <FolderNameDialog
        open={!!subfolderParent}
        onOpenChange={(open) => { if (!open) setSubfolderParent(null) }}
        title="New subfolder"
        onSubmit={(name) =>
          subfolderParent && createMutation.mutate({ name, parentId: subfolderParent.id })
        }
      />

      {/* Rename */}
      <FolderNameDialog
        open={!!renameTarget}
        onOpenChange={(open) => { if (!open) setRenameTarget(null) }}
        title="Rename folder"
        initialValue={renameTarget?.name ?? ''}
        onSubmit={(name) => renameTarget && renameMutation.mutate({ id: renameTarget.id, name })}
      />

      {/* Delete confirmation */}
      <Dialog open={!!deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete folder</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Are you sure you want to delete{' '}
            <span className="font-medium text-foreground">"{deleteTarget?.name}"</span>? This cannot
            be undone.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleDelete}>
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </aside>
  )
}
