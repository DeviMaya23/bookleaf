import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Loader2, Download, X, Copy } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { Switch } from '@/components/ui/switch'
import { getFolder, exportFolder, updateFolder } from '@/lib/folders'
import { getFolderShare, createFolderShare, deleteFolderShare } from '@/lib/share'
import { useFieldAutosave } from '../hooks/useFieldAutosave'
import DisableShareDialog from './DisableShareDialog'

interface FolderPanelContentProps {
  folder: { id: string; name: string; description: string | null }
  onClose: () => void
}

function formatImageCount(count: number): string {
  return count === 1 ? '1 image' : `${count} images`
}

export default function FolderPanelContent({ folder, onClose }: FolderPanelContentProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [isExporting, setIsExporting] = useState(false)
  const [isDisableDialogOpen, setIsDisableDialogOpen] = useState(false)

  const { data: folderDetail, isLoading: isFolderDetailLoading } = useQuery({
    queryKey: ['folder', folder.id],
    queryFn: () => getFolder(getToken, folder.id),
  })

  const { data: shareData, isLoading: isShareLoading } = useQuery({
    queryKey: ['folder-share', folder.id],
    queryFn: () => getFolderShare(getToken, folder.id),
  })

  const saveMutation = useMutation({
    mutationFn: (params: Parameters<typeof updateFolder>[2]) =>
      updateFolder(getToken, folder.id, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['folders'] })
      toast.success('Saved')
    },
    onError: () => {
      toast.error('Failed to save')
    },
  })

  const createShareMutation = useMutation({
    mutationFn: () => createFolderShare(getToken, folder.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['folder-share', folder.id] })
    },
    onError: () => {
      toast.error('Failed to enable sharing')
    },
  })

  const deleteShareMutation = useMutation({
    mutationFn: () => deleteFolderShare(getToken, folder.id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['folder-share', folder.id] })
      setIsDisableDialogOpen(false)
    },
    onError: () => {
      toast.error('Failed to disable sharing')
    },
  })

  const handleShareToggle = (checked: boolean) => {
    if (checked) {
      createShareMutation.mutate()
    } else {
      setIsDisableDialogOpen(true)
    }
  }

  const handleCopyShareLink = async () => {
    if (!shareData) return
    const url = `${window.location.origin}/share/${shareData.token}`
    try {
      await navigator.clipboard.writeText(url)
      toast.success('Link copied')
    } catch {
      toast.error('Failed to copy link')
    }
  }

  const handleExport = async () => {
    setIsExporting(true)
    try {
      const blob = await exportFolder(getToken, folder.id)
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `${folder.name}.zip`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch {
      toast.error('Failed to export folder')
    } finally {
      setIsExporting(false)
    }
  }

  const nameField = useFieldAutosave(
    folder.name,
    (value) => saveMutation.mutate({ name: value }),
    { isEmpty: (value) => value.trim() === '' },
  )
  const descriptionField = useFieldAutosave(
    folder.description ?? '',
    (value) => saveMutation.mutate({ description: value || null }),
  )

  return (
    <>
      <div className="relative flex-shrink-0 border-b px-4 pt-4 pb-3">
        <input
          value={nameField.value}
          onChange={(e) => nameField.onChange(e.target.value)}
          onBlur={nameField.onBlur}
          className="w-full text-base font-semibold bg-transparent border-b border-transparent focus:border-border focus:bg-muted/40 outline-none px-0 py-0.5 transition-colors"
        />
        {folderDetail && (
          <p className="text-xs text-muted-foreground mt-0.5">{formatImageCount(folderDetail.image_count)}</p>
        )}
        <button
          onClick={onClose}
          className="absolute top-2 right-2 w-7 h-7 rounded-full bg-black/40 text-white flex items-center justify-center hover:bg-black/60 transition-colors"
          aria-label="Close panel"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Notes</p>
          <textarea
            value={descriptionField.value}
            onChange={(e) => descriptionField.onChange(e.target.value)}
            onBlur={descriptionField.onBlur}
            placeholder="Add a note…"
            rows={3}
            className="w-full resize-none text-sm bg-muted/30 border border-border/50 rounded-lg px-3 py-2 outline-none focus:border-border transition-colors"
          />
        </div>

        <div className="px-4 py-3 border-b">
          <div className="flex items-center justify-between">
            <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">Share folder</p>
            <Switch
              checked={!!shareData}
              onCheckedChange={handleShareToggle}
              disabled={isShareLoading || createShareMutation.isPending || deleteShareMutation.isPending}
            />
          </div>
          {shareData && (
            <div className="mt-2 flex items-center gap-2">
              <input
                readOnly
                value={`${window.location.origin}/share/${shareData.token}`}
                className="flex-1 truncate text-sm bg-muted/30 border border-border/50 rounded-lg px-3 py-2 outline-none"
              />
              <Button
                variant="outline"
                size="icon-sm"
                onClick={handleCopyShareLink}
                aria-label="Copy share link"
              >
                <Copy className="w-3.5 h-3.5" />
              </Button>
            </div>
          )}
        </div>
      </div>

      <DisableShareDialog
        open={isDisableDialogOpen}
        onCancel={() => setIsDisableDialogOpen(false)}
        onConfirm={() => deleteShareMutation.mutate()}
      />

      <div className="flex-shrink-0 border-t px-4 py-3 bg-background">
        <Button
          className="w-full"
          onClick={handleExport}
          disabled={isExporting || isFolderDetailLoading || folderDetail?.image_count === 0}
        >
          {isExporting ? (
            <>
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              Preparing export...
            </>
          ) : (
            <>
              <Download className="w-4 h-4 mr-2" />
              Export folder
            </>
          )}
        </Button>
      </div>
    </>
  )
}
