import { Menu } from 'lucide-react'

interface MobileTopBarProps {
  onMenuClick: () => void
}

export default function MobileTopBar({ onMenuClick }: MobileTopBarProps) {
  return (
    <div className="sm:hidden fixed inset-x-0 top-0 z-20 flex items-center h-12 px-4 border-b bg-background">
      <button
        type="button"
        onClick={onMenuClick}
        aria-label="Open menu"
        className="flex items-center justify-center w-8 h-8 -ml-1 rounded-md hover:bg-accent"
      >
        <Menu className="w-5 h-5" />
      </button>
      <span className="flex-1 text-center font-serif text-base font-semibold pr-8">
        Bookleaf
      </span>
    </div>
  )
}
