import type { Image } from '@/lib/images'

function formatFileSize(bytes: number | null): string {
  if (bytes === null) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

interface DetailsGridProps {
  image: Image
}

export default function DetailsGrid({ image }: DetailsGridProps) {
  return (
    <div className="px-4 py-3">
      <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-3">Details</p>
      <div className="grid grid-cols-2 gap-x-4 gap-y-3">
        {[
          ['Size', formatFileSize(image.file_size)],
          ['Dimensions', image.width && image.height ? `${image.width} × ${image.height}` : '—'],
          ['Added', formatDate(image.created_at)],
        ].map(([label, value]) => (
          <div key={label}>
            <p className="text-[10px] text-muted-foreground mb-0.5">{label}</p>
            <p className="text-xs font-medium truncate">{value}</p>
          </div>
        ))}
      </div>
    </div>
  )
}
