import { useState } from 'react'
import { Link } from 'react-router-dom'
import { Loader2, Download } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { exportSharedFolder } from '@/lib/share'

interface SharedFolderPanelProps {
  token: string
  folder: { name: string; notes: string | null }
  imageCount: number
  className?: string
}

function formatImageCount(count: number): string {
  return count === 1 ? '1 image' : `${count} images`
}

export default function SharedFolderPanel({ token, folder, imageCount, className }: SharedFolderPanelProps) {
  const [isExporting, setIsExporting] = useState(false)

  const handleExport = async () => {
    setIsExporting(true)
    try {
      const blob = await exportSharedFolder(token)
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

  return (
    <div className={`flex flex-col border-b sm:h-full sm:w-[280px] sm:border-b-0 sm:border-l sm:flex-shrink-0${className ? ` ${className}` : ''}`}>
      <div className="flex-shrink-0 border-b px-4 pt-4 pb-3">
        <h1 className="text-base font-semibold">{folder.name}</h1>
        <p className="text-xs text-muted-foreground mt-0.5">{formatImageCount(imageCount)}</p>
      </div>

      <div className="sm:flex-1 sm:overflow-y-auto">
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Notes</p>
          <p className="text-sm text-muted-foreground whitespace-pre-wrap">
            {folder.notes || 'No notes added'}
          </p>
        </div>
      </div>

      <div className="flex-shrink-0 border-t px-4 py-3 bg-background flex flex-col gap-3">
        <Button
          className="w-full"
          onClick={handleExport}
          disabled={isExporting || imageCount === 0}
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
        <Link to="/" className="text-xs text-center text-muted-foreground hover:text-foreground transition-colors">
          Shared via Bookleaf
        </Link>
      </div>
    </div>
  )
}
