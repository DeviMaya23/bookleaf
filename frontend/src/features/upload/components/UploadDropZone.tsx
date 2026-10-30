import { UploadCloud } from 'lucide-react'

interface UploadDropZoneProps {
  isDragOver: boolean
  helpText: string
  hint: string
  error?: string | null
  onClick: () => void
  onDragOver: React.DragEventHandler<HTMLDivElement>
  onDragLeave: () => void
  onDrop: React.DragEventHandler<HTMLDivElement>
}

export default function UploadDropZone({
  isDragOver,
  helpText,
  hint,
  error,
  onClick,
  onDragOver,
  onDragLeave,
  onDrop,
}: UploadDropZoneProps) {
  return (
    <div
      role="button"
      tabIndex={0}
      className={`flex flex-col items-center justify-center gap-2 rounded-lg border-2 border-dashed p-8 cursor-pointer transition-colors ${
        isDragOver
          ? 'border-primary bg-primary/5'
          : 'border-muted-foreground/30 hover:border-muted-foreground/60'
      }`}
      onClick={onClick}
      onKeyDown={(e) => { if (e.key === 'Enter' || e.key === ' ') onClick() }}
      onDragOver={onDragOver}
      onDragLeave={onDragLeave}
      onDrop={onDrop}
    >
      <UploadCloud className="w-8 h-8 text-muted-foreground" />
      <p className="text-sm text-muted-foreground">{helpText}</p>
      <p className="text-xs text-muted-foreground">{hint}</p>
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  )
}
