import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import FolderSidebar from './FolderSidebar'
import ImageGrid from './ImageGrid'
import UploadModal from './UploadModal'
import RightPanel from './RightPanel'
import type { Image } from '@/lib/images'

export default function AppLayout() {
  const { folderId } = useParams<{ folderId: string }>()
  const [uploadOpen, setUploadOpen] = useState(false)
  const [selectedImage, setSelectedImage] = useState<Image | null>(null)

  return (
    <div className="flex h-screen">
      <FolderSidebar />
      <main className="ml-[240px] flex-1 h-screen min-w-0">
        <ScrollArea className="h-full">
          <div className="p-6">
            <div className="flex justify-end mb-4">
              <Button onClick={() => setUploadOpen(true)}>
                <Plus className="w-4 h-4 mr-1" />
                Image
              </Button>
            </div>
            <ImageGrid
              folderId={folderId ?? null}
              onImageSelect={setSelectedImage}
            />
          </div>
        </ScrollArea>
      </main>
      {selectedImage && (
        <RightPanel
          image={selectedImage}
          onClose={() => setSelectedImage(null)}
        />
      )}
      <UploadModal
        open={uploadOpen}
        onOpenChange={setUploadOpen}
        folderId={folderId ?? null}
      />
    </div>
  )
}
