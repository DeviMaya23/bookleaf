import { ScrollArea } from '@/components/ui/scroll-area'
import FolderSidebar from './FolderSidebar'
import ImageGrid from './ImageGrid'

export default function AppLayout() {
  return (
    <div className="flex h-screen">
      <FolderSidebar />
      <main className="ml-[240px] flex-1 h-screen">
        <ScrollArea className="h-full">
          <div className="p-6">
            <ImageGrid />
          </div>
        </ScrollArea>
      </main>
    </div>
  )
}
