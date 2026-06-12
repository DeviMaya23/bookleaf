import { useCallback, useRef, useState } from 'react'
import {
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from '@dnd-kit/core'
import { snapCenterToCursor } from '@dnd-kit/modifiers'
import { UploadCloud } from 'lucide-react'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useQueryClient } from '@tanstack/react-query'
import { toast } from 'sonner'
import type { Folder } from '@/lib/folders'
import type { SortEndTrigger } from '@/features/gallery/components/ImageGrid'
import { handleImageDrop, handleFolderDrop, isImageDragData, isFolderDragData, isDropData } from './lib/dragHandlers'

// Owns the cross-feature drag-and-drop orchestration: pointer sensors, the drag
// overlay, and the start/end handlers that reorder images (via sortEndTrigger)
// or move images/folders. File-drop-to-upload on the main area stays in the
// shell since it is upload, not dnd-kit dragging.
export function useAppDragAndDrop(folders: Folder[]) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [activeDragImage, setActiveDragImage] = useState<{ id: string; thumbnailUrl: string | null } | null>(null)
  const sortEndTriggerRef = useRef<number>(0)
  const [sortEndTrigger, setSortEndTrigger] = useState<SortEndTrigger | null>(null)

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } })
  )

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const data = event.active.data.current
    if (isImageDragData(data)) {
      setActiveDragImage({ id: data.imageId, thumbnailUrl: data.thumbnailUrl })
    }
  }, [])

  const handleDragEnd = useCallback(async (event: DragEndEvent) => {
    setActiveDragImage(null)
    const { active, over } = event
    if (!over) return

    const dragData = active.data.current
    const dropData = over.data.current
    if (!dragData) return

    // Image dropped on another image → reorder (handled inside ImageGrid via sortEndTrigger)
    if (isImageDragData(dragData) && isImageDragData(dropData)) {
      setSortEndTrigger({ activeId: String(active.id), overId: String(over.id), ts: ++sortEndTriggerRef.current })
      return
    }

    if (!isDropData(dropData)) return

    if (isImageDragData(dragData)) {
      try {
        const result = await handleImageDrop(getToken, dragData, dropData)
        if (result === 'moved') {
          queryClient.invalidateQueries({ queryKey: ['images'] })
          toast.success('Image moved')
        }
      } catch {
        toast.error('Failed to move image')
        queryClient.invalidateQueries({ queryKey: ['images'] })
      }
    } else if (isFolderDragData(dragData)) {
      try {
        const result = await handleFolderDrop(getToken, dragData, dropData, folders)
        if (result === 'moved') {
          queryClient.invalidateQueries({ queryKey: ['folders'] })
          toast.success('Folder moved')
        }
      } catch {
        toast.error('Failed to move folder')
      }
    }
  }, [getToken, queryClient, folders])

  const dragOverlay = (
    <DragOverlay dropAnimation={null} modifiers={[snapCenterToCursor]}>
      {activeDragImage && (
        <div className="w-20 h-20 rounded-lg overflow-hidden bg-card shadow-xl ring-1 ring-black/10 opacity-95">
          {activeDragImage.thumbnailUrl ? (
            <img src={activeDragImage.thumbnailUrl} className="w-full h-full object-cover" alt="" />
          ) : (
            <div className="w-full h-full bg-muted flex items-center justify-center">
              <UploadCloud className="w-6 h-6 text-muted-foreground" />
            </div>
          )}
        </div>
      )}
    </DragOverlay>
  )

  return { sensors, handleDragStart, handleDragEnd, sortEndTrigger, dragOverlay }
}
