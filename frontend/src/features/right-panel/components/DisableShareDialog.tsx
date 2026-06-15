import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface DisableShareDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
}

export default function DisableShareDialog({ open, onCancel, onConfirm }: DisableShareDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(open) => { if (!open) onCancel() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Disable sharing</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          The existing share link will stop working. If you re-enable sharing later, a new link
          will be generated.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>
            Cancel
          </Button>
          <Button variant="destructive" onClick={onConfirm}>
            Disable
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
