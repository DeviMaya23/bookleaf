import { useRef, useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { UploadCloud, Loader2 } from 'lucide-react'
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
  acceptSuggestion,
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
}

interface SuggestionState {
  imageId: string
  folderName: string
}

export default function UploadModal({ open, onOpenChange, folderId }: UploadModalProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [file, setFile] = useState<File | null>(null)
  const [title, setTitle] = useState('')
  const [typeError, setTypeError] = useState<string | null>(null)
  const [isDragOver, setIsDragOver] = useState(false)
  const [suggestion, setSuggestion] = useState<SuggestionState | null>(null)

  const inputRef = useRef<HTMLInputElement>(null)

  function handleFile(selected: File) {
    if (!isValidType(selected)) {
      setTypeError('Only JPEG, PNG, GIF, and WEBP files are supported.')
      setFile(null)
      return
    }
    setTypeError(null)
    setFile(selected)
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
    setTitle('')
    setTypeError(null)
    setSuggestion(null)
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
      })
      await putToR2(initiated.upload_url, file)
      const result = await completeUpload(getToken, initiated.id)
      return result
    },
    onSuccess: (result) => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      if (result.suggested_folder_name) {
        setSuggestion({ imageId: result.image_id, folderName: result.suggested_folder_name })
      } else {
        toast.success('Image uploaded successfully')
        handleClose()
      }
    },
    onError: () => {
      toast.error('Upload failed. Please try again.')
    },
  })

  const acceptMutation = useMutation({
    mutationFn: async () => {
      if (!suggestion) return
      await acceptSuggestion(getToken, suggestion.imageId, suggestion.folderName)
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images', 'folders'] })
      toast.success('Image uploaded and added to folder')
      handleClose()
    },
    onError: () => {
      toast.error('Failed to accept folder suggestion')
    },
  })

  function handleIgnore() {
    toast.success('Image uploaded successfully')
    handleClose()
  }

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose() }}>
      <DialogContent className="sm:max-w-md">
        <DialogHeader>
          <DialogTitle>Upload image</DialogTitle>
        </DialogHeader>

        {suggestion ? (
          <div className="flex flex-col gap-4 py-2">
            <p className="text-sm text-muted-foreground">
              AI suggested a folder for this image:
            </p>
            <p className="text-sm font-medium rounded-md border px-3 py-2 bg-muted">
              {suggestion.folderName}
            </p>
            <p className="text-sm text-muted-foreground">
              Add to this folder?
            </p>
            <DialogFooter className="flex-col gap-2 sm:flex-row">
              <Button
                variant="outline"
                onClick={handleIgnore}
                disabled={acceptMutation.isPending}
              >
                Ignore
              </Button>
              <Button
                onClick={() => acceptMutation.mutate()}
                disabled={acceptMutation.isPending}
              >
                {acceptMutation.isPending ? (
                  <><Loader2 className="w-4 h-4 animate-spin mr-2" />Accepting…</>
                ) : (
                  'Accept'
                )}
              </Button>
            </DialogFooter>
          </div>
        ) : (
          <div className="flex flex-col gap-4 py-2">
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
              {file ? (
                <p className="text-sm font-medium">{file.name}</p>
              ) : (
                <>
                  <p className="text-sm text-muted-foreground">Drop an image here or click to browse</p>
                  <p className="text-xs text-muted-foreground">JPEG, PNG, GIF, WEBP</p>
                </>
              )}
              {typeError && (
                <p className="text-xs text-destructive">{typeError}</p>
              )}
            </div>

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
              placeholder={file ? fileBaseName(file.name) : 'Title (optional)'}
              value={title}
              onChange={(e) => setTitle(e.target.value)}
            />

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
        )}
      </DialogContent>
    </Dialog>
  )
}
