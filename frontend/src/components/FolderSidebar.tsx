import { useState } from 'react'
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
import LogoutButton from './LogoutButton'
import { getFolders, createFolder, renameFolder, deleteFolder } from '@/lib/folders'
import type { Folder } from '@/lib/folders'

export default function FolderSidebar() {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
  })

  const [newFolderOpen, setNewFolderOpen] = useState(false)
  const [renameTarget, setRenameTarget] = useState<Folder | null>(null)
  const [deleteTarget, setDeleteTarget] = useState<Folder | null>(null)

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['folders'] })

  const createMutation = useMutation({
    mutationFn: (name: string) => createFolder(getToken, name),
    onSuccess: invalidate,
  })

  const renameMutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => renameFolder(getToken, id, name),
    onSuccess: invalidate,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteFolder(getToken, id),
    onSuccess: invalidate,
  })

  function handleDelete() {
    if (!deleteTarget) return
    deleteMutation.mutate(deleteTarget.id)
    setDeleteTarget(null)
  }

  return (
    <aside className="fixed inset-y-0 left-0 w-[240px] flex flex-col border-r bg-background">
      <div className="p-4 border-b">
        <span className="text-sm font-semibold tracking-tight">Bookleaf</span>
      </div>

      <nav className="flex-1 overflow-y-auto p-2">
        <ul className="space-y-1">
          <li className="rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground cursor-pointer">
            Unsorted
          </li>

          {folders.length > 0 && <li className="my-1 border-t" />}

          {folders.map((folder) => (
            <ContextMenu key={folder.id}>
              <ContextMenuTrigger asChild>
                <li className="rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground cursor-pointer">
                  {folder.name}
                </li>
              </ContextMenuTrigger>
              <ContextMenuContent>
                <ContextMenuItem onSelect={() => setRenameTarget(folder)}>
                  Rename
                </ContextMenuItem>
                <ContextMenuItem
                  onSelect={() => setDeleteTarget(folder)}
                  className="text-destructive focus:text-destructive"
                >
                  Delete
                </ContextMenuItem>
              </ContextMenuContent>
            </ContextMenu>
          ))}
        </ul>
      </nav>

      <div className="p-2 border-t space-y-1">
        <button
          className="w-full rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground text-left"
          onClick={() => setNewFolderOpen(true)}
        >
          + New folder
        </button>
        <div className="px-1">
          <LogoutButton />
        </div>
      </div>

      <FolderNameDialog
        open={newFolderOpen}
        onOpenChange={setNewFolderOpen}
        title="New folder"
        onSubmit={(name) => createMutation.mutate(name)}
      />

      <FolderNameDialog
        open={!!renameTarget}
        onOpenChange={(open) => { if (!open) setRenameTarget(null) }}
        title="Rename folder"
        initialValue={renameTarget?.name ?? ''}
        onSubmit={(name) => renameTarget && renameMutation.mutate({ id: renameTarget.id, name })}
      />

      <Dialog open={!!deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete folder</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Are you sure you want to delete <span className="font-medium text-foreground">"{deleteTarget?.name}"</span>? This cannot be undone.
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
