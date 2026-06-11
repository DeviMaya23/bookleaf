import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import type { Image } from '@/lib/images'

interface DeleteImageDialogProps {
  image: Image | null
  onCancel: () => void
  onConfirm: () => void
}

export default function DeleteImageDialog({ image, onCancel, onConfirm }: DeleteImageDialogProps) {
  return (
    <Dialog open={!!image} onOpenChange={(open) => { if (!open) onCancel() }}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>Delete permanently?</DialogTitle>
        </DialogHeader>
        <p className="text-sm text-muted-foreground">This image will be permanently deleted. This action cannot be undone.</p>
        <DialogFooter>
          <Button variant="outline" onClick={onCancel}>Cancel</Button>
          <Button variant="destructive" onClick={onConfirm}>Delete permanently</Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
