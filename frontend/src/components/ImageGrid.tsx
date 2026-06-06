import { useEffect, useRef, useState } from 'react'
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Loader2, ImageIcon } from 'lucide-react'
import { SortableContext, arrayMove, useSortable } from '@dnd-kit/sortable'
import { CSS } from '@dnd-kit/utilities'
import { useDndMonitor } from '@dnd-kit/core'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { getImages, getAllImages, getTrashedImages, deleteImage, restoreImage, updateImagePosition, computeNewPosition } from '@/lib/images'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'
import MasonryLayout, { MasonryCardContent } from '@/components/MasonryLayout'


export type LayoutMode = 'masonry'

export interface SortEndTrigger {
  activeId: string
  overId: string
  ts: number
}

interface ImageCardProps {
  image: Image
  imgHeight: number
  isTrash: boolean
  isDropTarget: boolean
  currentFolderId: string | null
  onAction: (image: Image) => void
  onSelect: (image: Image) => void
}

function ImageCard({ image, imgHeight, isTrash, isDropTarget, currentFolderId, onAction, onSelect }: ImageCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: `image-${image.id}`,
    disabled: isTrash,
    data: { type: 'image', imageId: image.id, currentFolderId, thumbnailUrl: image.thumbnail_url },
  })

  const style: React.CSSProperties = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.4 : 1,
  }

  return (
    <ContextMenu>
      <ContextMenuTrigger>
        <div
          ref={setNodeRef}
          style={style}
          {...listeners}
          {...attributes}
          className={`cursor-pointer rounded-lg overflow-hidden bg-card${isDropTarget ? ' ring-2 ring-primary' : ''}`}
          onClick={() => onSelect(image)}
        >
          <MasonryCardContent image={image} imgHeight={imgHeight} />
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        {isTrash ? (
          <ContextMenuItem onClick={() => onAction(image)}>Restore</ContextMenuItem>
        ) : (
          <ContextMenuItem
            onClick={() => onAction(image)}
            className="text-destructive focus:text-destructive"
          >
            Delete
          </ContextMenuItem>
        )}
      </ContextMenuContent>
    </ContextMenu>
  )
}

function queryKeyFor(view: AppView): unknown[] {
  switch (view.type) {
    case 'all': return ['images', 'all']
    case 'unsorted': return ['images', 'unsorted']
    case 'trash': return ['images', 'trash']
    case 'folder': return ['images', 'folder', view.id]
  }
}

function fetcherFor(view: AppView, getToken: () => Promise<string | undefined>) {
  return ({ pageParam }: { pageParam: string | undefined }) => {
    switch (view.type) {
      case 'all': return getAllImages(getToken, pageParam)
      case 'unsorted': return getImages(getToken, null, pageParam)
      case 'trash': return getTrashedImages(getToken, pageParam)
      case 'folder': return getImages(getToken, view.id, pageParam)
    }
  }
}

interface ImageGridProps {
  view: AppView
  layoutMode?: LayoutMode
  onImageSelect: (image: Image) => void
  onImageDeleted?: (id: string) => void
  sortEndTrigger?: SortEndTrigger | null
}

export default function ImageGrid({ view, layoutMode = 'masonry', onImageSelect, onImageDeleted, sortEndTrigger }: ImageGridProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const isTrash = view.type === 'trash'
  const isFolderView = view.type === 'folder'

  const containerRef = useRef<HTMLDivElement>(null)
  const [containerWidth, setContainerWidth] = useState(660)
  const [orderedImages, setOrderedImages] = useState<Image[]>([])
  const [dragOverId, setDragOverId] = useState<string | null>(null)
  const activeDragIdRef = useRef<string | null>(null)

  useDndMonitor({
    onDragStart(event) {
      activeDragIdRef.current = String(event.active.id)
    },
    onDragOver(event) {
      if (!isFolderView || event.active.data.current?.type !== 'image') return
      const overId = event.over ? String(event.over.id) : null
      setDragOverId(overId !== activeDragIdRef.current ? overId : null)
    },
    onDragEnd() { setDragOverId(null); activeDragIdRef.current = null },
    onDragCancel() { setDragOverId(null); activeDragIdRef.current = null },
  })

  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver(([entry]) => {
      setContainerWidth(entry.contentRect.width)
    })
    observer.observe(el)
    // Fallback for initial measurement when ResizeObserver fires before layout completes
    const rafId = requestAnimationFrame(() => {
      const w = el.getBoundingClientRect().width
      if (w > 0) setContainerWidth(w)
    })
    return () => { observer.disconnect(); cancelAnimationFrame(rafId) }
  }, [])

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: queryKeyFor(view),
    queryFn: fetcherFor(view, getToken),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
  })

  const allImages = data?.pages.flatMap((p) => p.images) ?? []

  useEffect(() => {
    setOrderedImages(allImages)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [data])

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteImage(getToken, id),
    onSuccess: (_, id) => {
      setOrderedImages((prev) => prev.filter((img) => img.id !== id))
      queryClient.invalidateQueries({ queryKey: ['images', 'trash'] })
      toast.success('Image moved to trash')
      onImageDeleted?.(id)
    },
    onError: () => toast.error('Failed to delete image'),
  })

  const restoreMutation = useMutation({
    mutationFn: (id: string) => restoreImage(getToken, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images', 'trash'] })
      toast.success('Image restored')
    },
    onError: () => toast.error('Failed to restore image'),
  })

  const positionMutation = useMutation({
    mutationFn: ({ imageId, folderId, position }: { imageId: string; folderId: string; position: string }) =>
      updateImagePosition(getToken, imageId, folderId, position),
  })

  // Keep a ref so the sortEndTrigger effect always sees current orderedImages
  const orderedImagesRef = useRef(orderedImages)
  useEffect(() => { orderedImagesRef.current = orderedImages }, [orderedImages])

  const lastProcessedTriggerTs = useRef<number>(-1)

  useEffect(() => {
    if (!sortEndTrigger || view.type !== 'folder') return
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
  }, [sortEndTrigger, view])

  function handleAction(image: Image) {
    if (isTrash) {
      restoreMutation.mutate(image.id)
    } else {
      deleteMutation.mutate(image.id)
    }
  }

  const folderId = view.type === 'folder' ? view.id : null
  const sortableItems = orderedImages.map((i) => `image-${i.id}`)

  if (layoutMode !== 'masonry') {
    console.warn(`ImageGrid: unsupported layoutMode "${layoutMode}"`)
    return null
  }

  const grid = (
    <MasonryLayout
      images={orderedImages}
      containerWidth={containerWidth}
      dropIndicatorId={dragOverId}
      renderCard={(image, imgHeight, isDropTarget) => (
        <ImageCard
          key={image.id}
          image={image}
          imgHeight={imgHeight}
          isTrash={isTrash}
          isDropTarget={isDropTarget}
          currentFolderId={folderId}
          onAction={handleAction}
          onSelect={onImageSelect}
        />
      )}
    />
  )

  return (
    <div ref={containerRef} className="w-full">
      {isLoading ? (
        <div className="flex items-center justify-center py-20">
          <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
        </div>
      ) : orderedImages.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
          <ImageIcon className="w-10 h-10" />
          <p className="text-sm">{isTrash ? 'Trash is empty' : 'No images here yet'}</p>
        </div>
      ) : isFolderView ? (
        <SortableContext items={sortableItems} strategy={() => null}>
          {grid}
        </SortableContext>
      ) : (
        grid
      )}

      {hasNextPage && (
        <div className="flex justify-center mt-6">
          <Button
            variant="outline"
            onClick={() => fetchNextPage()}
            disabled={isFetchingNextPage}
          >
            {isFetchingNextPage ? (
              <>
                <Loader2 className="w-4 h-4 animate-spin mr-2" />
                Loading...
              </>
            ) : (
              'Load more'
            )}
          </Button>
        </div>
      )}
    </div>
  )
}
