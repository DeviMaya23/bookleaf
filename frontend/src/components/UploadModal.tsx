import { useRef, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { UploadCloud, Loader2, X, ChevronDown, ChevronRight } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import {
  initiateUpload,
  putToR2,
  completeUpload,
} from '@/lib/images'

const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

function fileBaseName(name: string): string {
  return name.replace(/\.[^.]+$/, '')
}

function isValidType(file: File): boolean {
  return ACCEPTED_TYPES.includes(file.type)
}

interface UploadModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  folderId: string | null
  onUploadSuccess?: (imageId: string) => void
}

export default function UploadModal({ open, onOpenChange, folderId, onUploadSuccess }: UploadModalProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [file, setFile] = useState<File | null>(null)
  const [previewUrl, setPreviewUrl] = useState<string | null>(null)
  const [title, setTitle] = useState('')
  const [notes, setNotes] = useState('')
  const [sourceUrl, setSourceUrl] = useState('')
  const [detailsOpen, setDetailsOpen] = useState(false)
  const [typeError, setTypeError] = useState<string | null>(null)
  const [isDragOver, setIsDragOver] = useState(false)

  const inputRef = useRef<HTMLInputElement>(null)

  function revokePreview() {
    if (previewUrl) {
      URL.revokeObjectURL(previewUrl)
      setPreviewUrl(null)
    }
  }

  function handleFile(selected: File) {
    if (!isValidType(selected)) {
      setTypeError('Only JPEG, PNG, GIF, and WEBP files are supported.')
      setFile(null)
      revokePreview()
      return
    }
    setTypeError(null)
    revokePreview()
    const url = URL.createObjectURL(selected)
    setPreviewUrl(url)
    setFile(selected)
    setTitle(fileBaseName(selected.name))
  }

  function handleRemoveFile() {
    revokePreview()
    setFile(null)
    setTitle('')
    setTypeError(null)
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    setIsDragOver(false)
    const dropped = e.dataTransfer.files[0]
    if (dropped) handleFile(dropped)
  }

  function handlePickerChange(e: React.ChangeEvent<HTMLInputElement>) {
    const picked = e.target.files?.[0]
    if (picked) handleFile(picked)
    e.target.value = ''
  }

  function handleClose() {
    setFile(null)
    revokePreview()
    setTitle('')
    setNotes('')
    setSourceUrl('')
    setDetailsOpen(false)
    setTypeError(null)
    onOpenChange(false)
  }

  const uploadMutation = useMutation({
    mutationFn: async () => {
      if (!file) throw new Error('No file selected')
      const resolvedTitle = title.trim() !== '' ? title.trim() : fileBaseName(file.name)
      const initiated = await initiateUpload(getToken, {
        title: resolvedTitle,
        mimeType: file.type,
        folderId: folderId ?? undefined,
        description: notes.trim() || undefined,
        sourceUrl: sourceUrl.trim() || undefined,
      })
      await putToR2(initiated.upload_url, file)
      const result = await completeUpload(getToken, initiated.id)
      return result
    },
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      toast.success('Image uploaded successfully')
      onUploadSuccess?.(result.image_id)
      handleClose()
    },
    onError: () => {
      toast.error('Upload failed. Please try again.')
    },
  })

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose() }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Upload image</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-2">
          {previewUrl && file ? (
            <div className="flex items-center gap-3 rounded-lg border px-3 py-2 bg-muted/30">
              <img
                src={previewUrl}
                alt={file.name}
                className="h-12 w-12 rounded object-cover flex-shrink-0"
              />
              <span className="text-sm truncate flex-1 text-muted-foreground">{file.name}</span>
              <button
                type="button"
                onClick={handleRemoveFile}
                className="flex-shrink-0 p-1 rounded hover:bg-muted transition-colors text-muted-foreground"
                aria-label="Remove file"
              >
                <X className="w-4 h-4" />
              </button>
            </div>
          ) : (
            <div
              role="button"
              tabIndex={0}
              className={`flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-8 cursor-pointer transition-colors ${
                isDragOver
                  ? 'border-primary bg-primary/5'
                  : 'border-muted-foreground/30 hover:border-muted-foreground/60'
              }`}
              onClick={() => inputRef.current?.click()}
              onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') inputRef.current?.click() }}
              onDragOver={(e) => { e.preventDefault(); setIsDragOver(true) }}
              onDragLeave={() => setIsDragOver(false)}
              onDrop={handleDrop}
            >
              <UploadCloud className="w-8 h-8 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Drop an image here or click to browse</p>
              <p className="text-xs text-muted-foreground">JPEG, PNG, GIF, WEBP</p>
              {typeError && (
                <p className="text-xs text-destructive">{typeError}</p>
              )}
            </div>
          )}

          <input
            ref={inputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            className="hidden"
            onChange={handlePickerChange}
          />

          <input
            type="text"
            className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
            placeholder="Title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
          />

          <button
            type="button"
            className="flex items-center gap-1 text-xs text-muted-foreground hover:text-foreground transition-colors w-fit"
            onClick={() => setDetailsOpen((o) => !o)}
          >
            {detailsOpen
              ? <ChevronDown className="w-3.5 h-3.5" />
              : <ChevronRight className="w-3.5 h-3.5" />
            }
            Add details
          </button>

          {detailsOpen && (
            <div className="flex flex-col gap-3">
              <textarea
                rows={3}
                placeholder="Notes"
                value={notes}
                onChange={(e) => setNotes(e.target.value)}
                className="w-full resize-none rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
              <input
                type="text"
                placeholder="https://…"
                value={sourceUrl}
                onChange={(e) => setSourceUrl(e.target.value)}
                className="flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-sm transition-colors placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
              />
            </div>
          )}

          <DialogFooter>
            <Button
              onClick={() => uploadMutation.mutate()}
              disabled={!file || uploadMutation.isPending}
              className="w-full"
            >
              {uploadMutation.isPending ? (
                <><Loader2 className="w-4 h-4 animate-spin mr-2" />Uploading…</>
              ) : (
                'Upload'
              )}
            </Button>
          </DialogFooter>
        </div>
      </DialogContent>
    </Dialog>
  )
}
