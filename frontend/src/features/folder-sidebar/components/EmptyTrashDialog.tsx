import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface EmptyTrashDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
}

export default function EmptyTrashDialog({ open, onCancel, onConfirm }: EmptyTrashDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(open) => { if (!open) onCancel() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Empty trash?</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          All images in trash will be permanently deleted. This action cannot be undone.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm}>
            Empty trash
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
