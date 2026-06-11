import { useDroppable, useDndContext } from '@dnd-kit/core'

export default function RootDropZone() {
  const { active } = useDndContext()
  const isDraggingFolder = active?.data.current?.type === 'folder'

  const { setNodeRef, isOver } = useDroppable({
    id: 'root-drop',
    data: { type: 'root' },
  })

  if (!isDraggingFolder) return null

  return (
    <div
      ref={setNodeRef}
      className={`mx-2 mt-1 px-3 py-2 rounded-md border-2 border-dashed text-xs text-center select-none transition-colors ${
        isOver
          ? 'border-primary bg-primary/5 text-primary'
          : 'border-muted-foreground/30 text-muted-foreground/50'
      }`}
    >
      Move to root
    </div>
  )
}
