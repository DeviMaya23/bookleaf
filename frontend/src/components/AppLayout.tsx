import { useParams } from 'react-router-dom'
import { ScrollArea } from '@/components/ui/scroll-area'
import FolderSidebar from './FolderSidebar'
import ImageGrid from './ImageGrid'

export default function AppLayout() {
  const { folderId } = useParams<{ folderId: string }>()

  return (
    <div className="flex h-screen">
      <FolderSidebar />
      <main className="ml-[240px] flex-1 h-screen">
        <ScrollArea className="h-full">
          <div className="p-6">
            <ImageGrid folderId={folderId ?? null} />
          </div>
        </ScrollArea>
      </main>
    </div>
  )
}
