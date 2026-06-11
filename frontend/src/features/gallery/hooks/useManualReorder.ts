import { useCallback, useEffect, useRef, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useDndMonitor } from '@dnd-kit/core'
import { arrayMove } from '@dnd-kit/sortable'
import { toast } from 'sonner'
import { updateImagePosition, computeNewPosition } from '@/lib/images'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import type { SortBy } from './useGalleryControls'
import type { SortEndTrigger } from '../components/ImageGrid'

export function useManualReorder(view: AppView, sortBy: SortBy, isFolderView: boolean, images: Image[], sortEndTrigger?: SortEndTrigger | null) {
  const { getToken } = useKindeAuth()
  const [orderedImages, setOrderedImages] = useState<Image[]>([])
  const [dragOverId, setDragOverId] = useState<string | null>(null)
  const activeDragIdRef = useRef<string | null>(null)

  const removeImage = useCallback(
    (id: string) => setOrderedImages((prev) => prev.filter((img) => img.id !== id)),
    [],
  )

  useDndMonitor({
    onDragStart(event) {
      activeDragIdRef.current = String(event.active.id)
    },
    onDragOver(event) {
      if (!isFolderView || sortBy !== 'manual' || event.active.data.current?.type !== 'image') return
      const overId = event.over ? String(event.over.id) : null
      setDragOverId(overId !== activeDragIdRef.current ? overId : null)
    },
    onDragEnd() { setDragOverId(null); activeDragIdRef.current = null },
    onDragCancel() { setDragOverId(null); activeDragIdRef.current = null },
  })

  useEffect(() => {
    setOrderedImages(images)
  }, [images])

  const positionMutation = useMutation({
    mutationFn: ({ imageId, folderId, position }: { imageId: string; folderId: string; position: string }) =>
      updateImagePosition(getToken, imageId, folderId, position),
  })

  // Keep a ref so the sortEndTrigger effect always sees current orderedImages
  const orderedImagesRef = useRef(orderedImages)
  useEffect(() => { orderedImagesRef.current = orderedImages }, [orderedImages])

  const lastProcessedTriggerTs = useRef<number>(-1)

  useEffect(() => {
    if (!sortEndTrigger || view.type !== 'folder' || sortBy !== 'manual') return
    if (sortEndTrigger.ts <= lastProcessedTriggerTs.current) return
    lastProcessedTriggerTs.current = sortEndTrigger.ts
    const { activeId, overId } = sortEndTrigger
    const current = orderedImagesRef.current

    const oldIndex = current.findIndex((i) => `image-${i.id}` === activeId)
    const newIndex = current.findIndex((i) => `image-${i.id}` === overId)
    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) return

    const reordered = arrayMove(current, oldIndex, newIndex)

    let newKey: string
    try {
      newKey = computeNewPosition(reordered, newIndex)
    } catch {
      toast.error('Failed to compute position')
      return
    }

    const snapshot = current
    setOrderedImages(reordered.map((img, idx) =>
      idx === newIndex ? { ...img, position: newKey } : img
    ))

    positionMutation.mutate(
      { imageId: current[oldIndex].id, folderId: view.id, position: newKey },
      {
        onError: () => {
          setOrderedImages(snapshot)
          toast.error('Failed to save order')
        },
      }
    )
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [sortEndTrigger, view, sortBy])

  const sortableItems = orderedImages.map((i) => `image-${i.id}`)

  return { orderedImages, dragOverId, removeImage, sortableItems }
}
