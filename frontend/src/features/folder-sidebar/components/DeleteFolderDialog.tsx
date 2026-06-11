import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import type { Folder } from '@/lib/folders'

interface DeleteFolderDialogProps {
  folder: Folder | null
  onCancel: () => void
  onConfirm: () => void
}

export default function DeleteFolderDialog({ folder, onCancel, onConfirm }: DeleteFolderDialogProps) {
  return (
    <Dialog open={!!folder} onOpenChange={(open) => { if (!open) onCancel() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete folder</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          Are you sure you want to delete{' '}
          <span className="font-medium text-foreground">"{folder?.name}"</span>? This cannot
          be undone.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm}>
            Delete
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
