import { useState } from 'react'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useQuery } from '@tanstack/react-query'
import { Trash2, Download, Loader2, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { getFolders } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import { getTags } from '@/lib/tags'
import type { Tag } from '@/lib/tags'
import FolderInput from './FolderInput'
import TagInput from './TagInput'

interface SelectionPanelBodyProps {
  selectedCount: number
  onApply: (params: { folderIds: string[]; tags: Tag[] }) => void
  onMoveToTrash: () => void
  onExitSelectMode: () => void
  onDownloadZip: () => Promise<void>
}

export default function SelectionPanelBody({
  selectedCount,
  onApply,
  onMoveToTrash,
  onExitSelectMode,
  onDownloadZip,
}: SelectionPanelBodyProps) {
  const { getToken } = useKindeAuth()
  const [pendingFolders, setPendingFolders] = useState<Folder[]>([])
  const [pendingTags, setPendingTags] = useState<Tag[]>([])
  const [isDownloading, setIsDownloading] = useState(false)
  const [trashConfirmOpen, setTrashConfirmOpen] = useState(false)

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  const { data: allTags = [] } = useQuery({
    queryKey: ['tags'],
    queryFn: () => getTags(getToken),
    staleTime: 60_000,
  })

  const handleApply = () => {
    onApply({
      folderIds: pendingFolders.map((f) => f.id),
      tags: pendingTags,
    })
    setPendingFolders([])
    setPendingTags([])
  }

  const handleDownload = async () => {
    setIsDownloading(true)
    try {
      await onDownloadZip()
    } finally {
      setIsDownloading(false)
    }
  }

  const handleTrashConfirm = () => {
    setTrashConfirmOpen(false)
    onMoveToTrash()
  }

  const folderSuggestions = folders.filter((f) => !pendingFolders.some((p) => p.id === f.id))
  const tagSuggestions = allTags.filter((t) => !pendingTags.some((p) => p.id === t.id))
  const applyDisabled = pendingFolders.length === 0 && pendingTags.length === 0

  return (
    <>
      <div className="flex-shrink-0 border-b px-4 pt-4 pb-3 pr-10">
        <div className="flex items-center gap-2">
          <p className="text-base font-semibold">{selectedCount} selected</p>
          <button
            type="button"
            onClick={onExitSelectMode}
            className="text-muted-foreground hover:text-foreground transition-colors"
            aria-label="Exit select mode"
          >
            <X className="w-4 h-4" />
          </button>
        </div>
      </div>

      {selectedCount > 0 && <div className="flex-1 overflow-y-auto px-4 py-3 space-y-3">
        <div className="space-y-3">
          <div>
            <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">
              Add to folder
            </p>
            <FolderInput
              folders={pendingFolders}
              onChange={setPendingFolders}
              suggestions={folderSuggestions}
            />
          </div>

          <div>
            <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">
              Add tags
            </p>
            <TagInput
              tags={pendingTags}
              onChange={setPendingTags}
              suggestions={tagSuggestions}
            />
          </div>

          <button
            type="button"
            onClick={handleApply}
            disabled={applyDisabled}
            className="w-full rounded-lg border px-3 py-1.5 text-sm font-medium transition-colors disabled:opacity-40 disabled:pointer-events-none hover:bg-muted/60"
          >
            Apply
          </button>
        </div>

        <div className="space-y-3 pt-1 border-t">
          <Button className="w-full" onClick={handleDownload} disabled={isDownloading}>
            {isDownloading ? (
              <>
                <Loader2 className="w-4 h-4 mr-2 animate-spin" />
                Exporting…
              </>
            ) : (
              <>
                <Download className="w-4 h-4 mr-2" />
                Export images
              </>
            )}
          </Button>

          <div className="flex justify-center">
            <button
              type="button"
              onClick={() => setTrashConfirmOpen(true)}
              className="inline-flex items-center gap-1 text-sm text-destructive hover:text-destructive/80 transition-colors"
            >
              <Trash2 className="w-3.5 h-3.5" />
              Move to trash
            </button>
          </div>
        </div>
      </div>}

      <Dialog open={trashConfirmOpen} onOpenChange={(open) => { if (!open) setTrashConfirmOpen(false) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Move {selectedCount} image{selectedCount === 1 ? '' : 's'} to trash?</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            You can restore them from the trash later.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setTrashConfirmOpen(false)}>
              Cancel
            </Button>
            <Button variant="destructive" onClick={handleTrashConfirm}>
              Move to trash
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
