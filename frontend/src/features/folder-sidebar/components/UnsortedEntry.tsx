import { useDroppable } from '@dnd-kit/core'

interface UnsortedEntryProps {
  active: boolean
  activeDragType: string | null
  onClick: () => void
}

export default function UnsortedEntry({ active, activeDragType, onClick }: UnsortedEntryProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: 'unsorted-drop',
    data: { type: 'unsorted' },
  })

  const isDropTarget = activeDragType === 'image' && isOver

  return (
    <div
      ref={setNodeRef}
      className={`px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
        active
          ? 'bg-accent text-accent-foreground font-medium'
          : isDropTarget
          ? 'bg-accent text-accent-foreground ring-1 ring-primary/40'
          : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
      }`}
      onClick={onClick}
    >
      Unsorted
    </div>
  )
}
