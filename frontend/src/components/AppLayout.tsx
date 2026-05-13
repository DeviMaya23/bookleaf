import { ScrollArea } from '@/components/ui/scroll-area'
import ImageGrid from './ImageGrid'
import LogoutButton from './LogoutButton'

const PLACEHOLDER_FOLDERS = ['Unfiled', 'Nature', 'Travel']

export default function AppLayout() {
  return (
    <div className="flex h-screen">
      <aside className="fixed inset-y-0 left-0 w-[240px] flex flex-col border-r bg-background">
        <div className="p-4 border-b">
          <span className="text-sm font-semibold tracking-tight">Bookleaf</span>
        </div>
        <nav className="flex-1 overflow-y-auto p-2">
          <ul className="space-y-1">
            {PLACEHOLDER_FOLDERS.map((folder) => (
              <li
                key={folder}
                className="rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground cursor-pointer"
              >
                {folder}
              </li>
            ))}
          </ul>
        </nav>
        <div className="p-2 border-t space-y-1">
          <button className="w-full rounded-md px-3 py-1.5 text-sm text-muted-foreground hover:bg-accent hover:text-accent-foreground text-left">
            + New folder
          </button>
          <div className="px-1">
            <LogoutButton />
          </div>
        </div>
      </aside>

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
