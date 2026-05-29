import { useRef, useState, useEffect, useCallback } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { UploadCloud, Loader2, CheckCircle2, AlertCircle, Clock } from 'lucide-react'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { initiateUpload, putToR2, completeUpload } from '@/lib/images'

const MAX_CONCURRENT = 3
const MAX_FILES = 20
const MAX_SIZE_BYTES = 50 * 1024 * 1024
const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp']

type FileStatus = 'PENDING' | 'UPLOADING' | 'SUCCESS' | 'FAILED_FINAL' | 'OVERSIZED' | 'UNSUPPORTED'

interface BatchFile {
  id: string
  file: File
  status: FileStatus
  retryCount: number
}

interface BatchUploadModalProps {
  open: boolean
  onOpenChange: (open: boolean) => void
  folderId: string | null
  initialFiles?: File[]
}

function fileBaseName(name: string): string {
  return name.replace(/\.[^.]+$/, '')
}

function makeId(file: File): string {
  return `${file.name}-${file.size}-${file.lastModified}-${Math.random().toString(36).slice(2)}`
}

function makeBatchFile(file: File): BatchFile {
  const status: FileStatus = !ACCEPTED_TYPES.includes(file.type)
    ? 'UNSUPPORTED'
    : file.size > MAX_SIZE_BYTES
      ? 'OVERSIZED'
      : 'PENDING'
  return { id: makeId(file), file, status, retryCount: 0 }
}

function StatusCell({
  batchFile,
  onRetry,
}: {
  batchFile: BatchFile
  onRetry: (id: string) => void
}) {
  switch (batchFile.status) {
    case 'PENDING':
      return <Clock className="w-4 h-4 text-muted-foreground flex-shrink-0" aria-label="Waiting" />
    case 'UPLOADING':
      return <Loader2 className="w-4 h-4 animate-spin text-primary flex-shrink-0" aria-label="Uploading" />
    case 'SUCCESS':
      return <CheckCircle2 className="w-4 h-4 text-green-500 flex-shrink-0" aria-label="Uploaded" />
    case 'FAILED_FINAL':
      return (
        <div className="flex items-center gap-2 flex-shrink-0">
          <AlertCircle className="w-4 h-4 text-destructive" aria-label="Failed" />
          <Button
            variant="outline"
            size="sm"
            className="h-6 px-2 text-xs"
            onClick={() => onRetry(batchFile.id)}
          >
            Retry
          </Button>
        </div>
      )
    case 'OVERSIZED':
      return (
        <span className="text-xs font-medium text-destructive bg-destructive/10 px-2 py-0.5 rounded flex-shrink-0">
          Too large
        </span>
      )
    case 'UNSUPPORTED':
      return (
        <span className="text-xs font-medium text-destructive bg-destructive/10 px-2 py-0.5 rounded flex-shrink-0">
          Unsupported type
        </span>
      )
    default:
      return null
  }
}

export default function BatchUploadModal({
  open,
  onOpenChange,
  folderId,
  initialFiles,
}: BatchUploadModalProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [files, setFiles] = useState<BatchFile[]>([])
  const [countError, setCountError] = useState<string | null>(null)
  const [isDragOver, setIsDragOver] = useState(false)
  const [isUploading, setIsUploading] = useState(false)

  const inputRef = useRef<HTMLInputElement>(null)
  const inFlightRef = useRef(0)
  const filesRef = useRef<BatchFile[]>([])
  const closedRef = useRef(false)

  // Updates filesRef synchronously so async upload logic always reads current state,
  // then schedules a React state update for rendering.
  function updateFile(id: string, updates: Partial<BatchFile>) {
    if (closedRef.current) return
    const next = filesRef.current.map(f => f.id === id ? { ...f, ...updates } : f)
    filesRef.current = next
    setFiles(next)
  }

  // scheduleNextRef breaks the circular dep: runUpload calls scheduleNext via ref,
  // so runUpload doesn't need scheduleNext in its deps array.
  const scheduleNextRef = useRef<() => void>(() => {})

  const runUpload = useCallback(async (batchFile: BatchFile) => {
    inFlightRef.current++
    updateFile(batchFile.id, { status: 'UPLOADING' })
    try {
      const initiated = await initiateUpload(getToken, {
        title: fileBaseName(batchFile.file.name),
        mimeType: batchFile.file.type,
        folderId: folderId ?? undefined,
      })
      await putToR2(initiated.upload_url, batchFile.file)
      await completeUpload(getToken, initiated.id)
      if (!closedRef.current) {
        updateFile(batchFile.id, { status: 'SUCCESS' })
        queryClient.invalidateQueries({ queryKey: ['images'] })
      }
    } catch {
      if (!closedRef.current) {
        const current = filesRef.current.find(f => f.id === batchFile.id)
        if (current && current.retryCount < 1) {
          updateFile(batchFile.id, { status: 'PENDING', retryCount: current.retryCount + 1 })
        } else {
          updateFile(batchFile.id, { status: 'FAILED_FINAL' })
        }
      }
    } finally {
      inFlightRef.current--
      if (!closedRef.current) scheduleNextRef.current()
    }
  }, [getToken, folderId, queryClient])

  const scheduleNext = useCallback(() => {
    const slots = MAX_CONCURRENT - inFlightRef.current
    if (slots <= 0) return
    const pending = filesRef.current.filter(f => f.status === 'PENDING')
    pending.slice(0, slots).forEach(f => runUpload(f))
  }, [runUpload])

  scheduleNextRef.current = scheduleNext

  useEffect(() => {
    if (open) {
      closedRef.current = false
      inFlightRef.current = 0
      filesRef.current = []
      setFiles([])
      setCountError(null)
      setIsDragOver(false)
      setIsUploading(false)

      if (initialFiles && initialFiles.length > 0) {
        if (initialFiles.length > MAX_FILES) {
          setCountError(`You can upload a maximum of ${MAX_FILES} files at a time`)
        } else {
          const batch = initialFiles.map(makeBatchFile)
          filesRef.current = batch
          setFiles(batch)
        }
      }
    }
  }, [open]) // eslint-disable-line react-hooks/exhaustive-deps

  function addFiles(newFiles: File[]) {
    const total = filesRef.current.length + newFiles.length
    if (total > MAX_FILES) {
      setCountError(`You can upload a maximum of ${MAX_FILES} files at a time`)
      return
    }
    setCountError(null)
    const batch = newFiles.map(makeBatchFile)
    const next = [...filesRef.current, ...batch]
    filesRef.current = next
    setFiles(next)
  }

  function handlePickerChange(e: React.ChangeEvent<HTMLInputElement>) {
    const picked = Array.from(e.target.files ?? [])
    if (picked.length > 0) addFiles(picked)
    e.target.value = ''
  }

  function handleDrop(e: React.DragEvent<HTMLDivElement>) {
    e.preventDefault()
    setIsDragOver(false)
    const dropped = Array.from(e.dataTransfer.files)
    if (dropped.length > 0) addFiles(dropped)
  }

  function handleRetry(id: string) {
    const next = filesRef.current.map(f => f.id === id ? { ...f, status: 'PENDING' as FileStatus } : f)
    filesRef.current = next
    setFiles(next)
    scheduleNext()
  }

  function startUploads() {
    setIsUploading(true)
    scheduleNext()
  }

  function handleClose() {
    closedRef.current = true
    onOpenChange(false)
  }

  const uploadableCount = files.filter(f => f.status === 'PENDING').length
  const hasUploadable = uploadableCount > 0

  return (
    <Dialog open={open} onOpenChange={(o) => { if (!o) handleClose() }}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>Upload images</DialogTitle>
        </DialogHeader>

        <div className="flex flex-col gap-4 py-2">
          {!isUploading && (
            <div
              role="button"
              tabIndex={0}
              className={`flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-6 cursor-pointer transition-colors ${
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
              <UploadCloud className="w-7 h-7 text-muted-foreground" />
              <p className="text-sm text-muted-foreground">Drop images here or click to browse</p>
              <p className="text-xs text-muted-foreground">JPEG, PNG, GIF, WEBP · max 50 MB · up to 20 files</p>
            </div>
          )}

          <input
            ref={inputRef}
            type="file"
            accept="image/jpeg,image/png,image/gif,image/webp"
            multiple
            className="hidden"
            onChange={handlePickerChange}
          />

          {countError && (
            <p className="text-sm text-destructive">{countError}</p>
          )}

          {files.length > 0 && (
            <div className="flex flex-col gap-1 max-h-72 overflow-y-auto pr-1">
              {files.map(f => (
                <div key={f.id} className="flex items-center gap-3 rounded-lg border px-3 py-2">
                  <span className="text-sm truncate flex-1 text-muted-foreground min-w-0">{f.file.name}</span>
                  <StatusCell batchFile={f} onRetry={handleRetry} />
                </div>
              ))}
            </div>
          )}

          {!isUploading && (
            <DialogFooter>
              <Button
                onClick={startUploads}
                disabled={!hasUploadable}
                className="w-full"
              >
                {uploadableCount > 0
                  ? `Upload ${uploadableCount} image${uploadableCount !== 1 ? 's' : ''}`
                  : 'Upload'}
              </Button>
            </DialogFooter>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}
