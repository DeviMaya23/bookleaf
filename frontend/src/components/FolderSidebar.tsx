import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { toast } from 'sonner'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useDraggable, useDroppable, useDndContext } from '@dnd-kit/core'
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

export function getFolderSubtreeIds(folders: Folder[], folderId: string): Set<string> {
  const ids = new Set<string>([folderId])
  let changed = true
  while (changed) {
    changed = false
    for (const f of folders) {
      if (f.parent_id && ids.has(f.parent_id) && !ids.has(f.id)) {
        ids.add(f.id)
        changed = true
      }
    }
  }
  return ids
}

interface FolderItemProps {
  folder: FolderNode
  depth: number
  view: AppView
  folders: Folder[]
  activeDragType: string | null
  activeDragFolderId: string | null
  onSelect: (folder: FolderNode) => void
  onRename: (folder: Folder) => void
  onDelete: (folder: Folder) => void
  onNewSubfolder: (folder: Folder) => void
}

function FolderItem({
  folder, depth, view, folders,
  activeDragType, activeDragFolderId,
  onSelect, onRename, onDelete, onNewSubfolder,
}: FolderItemProps) {
  const [open, setOpen] = useState(depth === 0)
  const hasChildren = folder.children.length > 0
  const isActive = view.type === 'folder' && view.id === folder.id

  const subtreeIds = activeDragFolderId
    ? getFolderSubtreeIds(folders, activeDragFolderId)
    : new Set<string>()
  const isInvalidFolderTarget = activeDragType === 'folder' && subtreeIds.has(folder.id)

  const { setNodeRef: setDropRef, isOver } = useDroppable({
    id: `folder-drop-${folder.id}`,
    disabled: isInvalidFolderTarget,
    data: { type: 'folder', folderId: folder.id },
  })

  const { attributes, listeners, setNodeRef: setDragRef, isDragging } = useDraggable({
    id: `folder-drag-${folder.id}`,
    data: { type: 'folder', folderId: folder.id, name: folder.name, parentId: folder.parent_id },
  })

  const isImageDropTarget = activeDragType === 'image' && isOver
  const isFolderDropTarget = activeDragType === 'folder' && isOver && !isInvalidFolderTarget

  const setRef = (el: HTMLDivElement | null) => {
    setDropRef(el)
    setDragRef(el)
  }

  return (
    <div style={{ opacity: isDragging ? 0.4 : 1 }}>
      <ContextMenu>
        <ContextMenuTrigger>
          <div
            ref={setRef}
            {...listeners}
            {...attributes}
            style={{ paddingLeft: 8 + depth * 14 }}
            className={`flex items-center gap-1 pr-2 py-1 rounded-md cursor-pointer mb-0.5 text-sm select-none ${
              isActive
                ? 'bg-accent text-accent-foreground font-medium'
                : isImageDropTarget || isFolderDropTarget
                ? 'bg-accent text-accent-foreground ring-1 ring-primary/40'
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
              folders={folders}
              activeDragType={activeDragType}
              activeDragFolderId={activeDragFolderId}
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

interface UnsortedEntryProps {
  active: boolean
  activeDragType: string | null
  onClick: () => void
}

function UnsortedEntry({ active, activeDragType, onClick }: UnsortedEntryProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'unsorted-drop',
    data: { type: 'unsorted' },
  })

  const isDropTarget = activeDragType === 'image' && isOver

  return (
    <div
      ref={setNodeRef}
      className={`px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
        active
          ? 'bg-accent text-accent-foreground font-medium'
          : isDropTarget
          ? 'bg-accent text-accent-foreground ring-1 ring-primary/40'
          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
      }`}
      onClick={onClick}
    >
      Unsorted
    </div>
  )
}

function RootDropZone() {
  const { active } = useDndContext()
  const isDraggingFolder = active?.data.current?.type === 'folder'

  const { setNodeRef, isOver } = useDroppable({
    id: 'root-drop',
    data: { type: 'root' },
  })

  if (!isDraggingFolder) return null

  return (
    <div
      ref={setNodeRef}
      className={`mx-2 mt-1 px-3 py-2 rounded-md border-2 border-dashed text-xs text-center select-none transition-colors ${
        isOver
          ? 'border-primary bg-primary/5 text-primary'
          : 'border-muted-foreground/30 text-muted-foreground/50'
      }`}
    >
      Move to root
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
  const { active } = useDndContext()

  const activeDragType = (active?.data.current?.type as string) ?? null
  const activeDragFolderId = activeDragType === 'folder'
    ? (active?.data.current?.folderId as string)
    : null

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
        <div className="mt-[8px]">
          <div
            className={`px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
              view.type === 'trash'
                ? 'bg-accent text-accent-foreground font-medium'
                : 'text-muted-foreground/60 hover:bg-accent hover:text-accent-foreground'
            }`}
            onClick={() => navigate('/trash')}
          >
            Trash
          </div>
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
            folders={folders}
            activeDragType={activeDragType}
            activeDragFolderId={activeDragFolderId}
            onSelect={handleFolderSelect}
            onRename={setRenameTarget}
            onDelete={setDeleteTarget}
            onNewSubfolder={setSubfolderParent}
          />
        ))}

        <RootDropZone />
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
