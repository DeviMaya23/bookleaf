import { Dialog as DialogPrimitive } from '@base-ui/react/dialog'
import {
  Dialog,
  DialogClose,
  DialogFooter,
  DialogHeader,
  DialogOverlay,
  DialogPortal,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { XIcon } from 'lucide-react'

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
      <DialogPortal container={document.body}>
        <DialogOverlay forceRender onClick={onCancel} />
        <DialogPrimitive.Popup className="fixed top-1/2 left-1/2 z-50 grid w-full max-w-[calc(100%-2rem)] -translate-x-1/2 -translate-y-1/2 gap-4 rounded-xl bg-popover p-4 text-sm text-popover-foreground ring-1 ring-foreground/10 duration-100 outline-none sm:max-w-sm data-open:animate-in data-open:fade-in-0 data-open:zoom-in-95 data-closed:animate-out data-closed:fade-out-0 data-closed:zoom-out-95">
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
          <DialogClose
            data-slot="dialog-close"
            render={
              <Button
                variant="ghost"
                className="absolute top-2 right-2"
                size="icon-sm"
              />
            }
          >
            <XIcon />
            <span className="sr-only">Close</span>
          </DialogClose>
        </DialogPrimitive.Popup>
      </DialogPortal>
    </Dialog>
  )
}
