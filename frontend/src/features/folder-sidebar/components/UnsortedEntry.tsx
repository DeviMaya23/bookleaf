import { useDroppable } from '@dnd-kit/core'
import { FOLDER_ICONS, SYSTEM_ICON_KEYS } from '../lib/folderIcons'

interface UnsortedEntryProps {
  active: boolean
  activeDragType: string | null
  iconsEnabled: boolean
  onClick: () => void
}

export default function UnsortedEntry({ active, activeDragType, iconsEnabled, onClick }: UnsortedEntryProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'unsorted-drop',
    data: { type: 'unsorted' },
  })

  const isDropTarget = activeDragType === 'image' && isOver
  const Icon = FOLDER_ICONS[SYSTEM_ICON_KEYS.unsorted]

  return (
    <div
      ref={setNodeRef}
      className={`flex items-center gap-1.5 px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
        active
          ? 'bg-accent text-accent-foreground font-medium'
          : isDropTarget
          ? 'bg-accent text-accent-foreground ring-1 ring-primary/40'
          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
      }`}
      onClick={onClick}
    >
      {iconsEnabled && <Icon className="w-3.5 h-3.5 shrink-0" />}
      Unsorted
    </div>
  )
}
