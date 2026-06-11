import { useState } from 'react'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Loader2, Download } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import { downloadImage } from '@/lib/images'

interface DownloadButtonProps {
  imageId: string
}

export default function DownloadButton({ imageId }: DownloadButtonProps) {
  const { getToken } = useKindeAuth()
  const [isDownloading, setIsDownloading] = useState(false)

  const handleDownload = async () => {
    setIsDownloading(true)
    try {
      const url = await downloadImage(getToken, imageId)
      const a = document.createElement('a')
      a.href = url
      a.download = ''
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    } catch {
      toast.error('Failed to download image')
    } finally {
      setIsDownloading(false)
    }
  }

  return (
    <div className="flex-shrink-0 border-t px-4 py-3 bg-background">
      <Button
        className="w-full"
        onClick={handleDownload}
        disabled={isDownloading}
      >
        {isDownloading ? (
          <>
            <Loader2 className="w-4 h-4 mr-2 animate-spin" />
            Downloading…
          </>
        ) : (
          <>
            <Download className="w-4 h-4 mr-2" />
            Download image
          </>
        )}
      </Button>
    </div>
  )
}
