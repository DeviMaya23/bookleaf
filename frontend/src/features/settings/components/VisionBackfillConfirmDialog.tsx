import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'

interface VisionBackfillConfirmDialogProps {
  open: boolean
  onCancel: () => void
  onConfirm: () => void
  isPending: boolean
}

export default function VisionBackfillConfirmDialog({
  open,
  onCancel,
  onConfirm,
  isPending,
}: VisionBackfillConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(isOpen) => { if (!isOpen) onCancel() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Enable Smart Features</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">
          Enabling this feature will send a thumbnail version of all images to Google's Vision API for processing. The backfill process might take awhile to complete; it's recommended to wait a few minutes before using any smart features for the best experience.
        </p>
        <p className="text-sm text-muted-foreground">
          For a detailed explanation, see our <a href="/ainotes" className="underline">explanation of smart features</a> and our <a href="/privacy" className="underline">privacy policy</a>.
        </p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={isPending}>
            Cancel
          </Button>
          <Button onClick={onConfirm} disabled={isPending}>
            Enable
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
