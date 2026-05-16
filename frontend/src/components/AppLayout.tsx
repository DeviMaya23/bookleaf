import { useState } from 'react'
import { useParams } from 'react-router-dom'
import { Plus } from 'lucide-react'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Button } from '@/components/ui/button'
import FolderSidebar from './FolderSidebar'
import ImageGrid from './ImageGrid'
import UploadModal from './UploadModal'

export default function AppLayout() {
  const { folderId } = useParams<{ folderId: string }>()
  const [uploadOpen, setUploadOpen] = useState(false)

  return (
    <div className="flex h-screen">
      <FolderSidebar />
      <main className="ml-[240px] flex-1 h-screen">
        <ScrollArea className="h-full">
          <div className="p-6">
            <div className="flex justify-end mb-4">
              <Button onClick={() => setUploadOpen(true)}>
                <Plus className="w-4 h-4 mr-1" />
                Image
              </Button>
            </div>
            <ImageGrid folderId={folderId ?? null} />
          </div>
        </ScrollArea>
      </main>
      <UploadModal
        open={uploadOpen}
        onOpenChange={setUploadOpen}
        folderId={folderId ?? null}
      />
    </div>
  )
}
