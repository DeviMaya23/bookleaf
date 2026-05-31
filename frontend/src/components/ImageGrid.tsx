
import { useInfiniteQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Loader2, ImageIcon } from 'lucide-react'
import { useDraggable } from '@dnd-kit/core'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { getImages, getAllImages, getTrashedImages, deleteImage, restoreImage } from '@/lib/images'
import type { Image } from '@/lib/images'
import type { AppView } from '@/lib/view'

interface ImageCardProps {
  image: Image
  isTrash: boolean
  currentFolderId: string | null
  onAction: (image: Image) => void
  onSelect: (image: Image) => void
}

function ImageCard({ image, isTrash, currentFolderId, onAction, onSelect }: ImageCardProps) {
  const { attributes, listeners, setNodeRef, isDragging } = useDraggable({
    id: `image-${image.id}`,
    disabled: isTrash,
    data: { type: 'image', imageId: image.id, currentFolderId, thumbnailUrl: image.thumbnail_url },
  })

  return (
    <ContextMenu>
      <ContextMenuTrigger>
        <div
          ref={setNodeRef}
          {...listeners}
          {...attributes}
          className="cursor-pointer rounded-lg overflow-hidden bg-card break-inside-avoid mb-3"
          style={{ opacity: isDragging ? 0.4 : 1 }}
          onClick={() => onSelect(image)}
        >
          <div className="bg-muted">
            {image.thumbnail_url ? (
              <img
                src={image.thumbnail_url}
                alt={image.title}
                className="w-full h-auto object-cover"
              />
            ) : (
              <div className="w-full aspect-square flex items-center justify-center">
                <ImageIcon className="w-8 h-8 text-muted-foreground" />
              </div>
            )}
          </div>
          <div className="p-2">
            <p className="text-sm truncate">{image.title}</p>
          </div>
        </div>
      </ContextMenuTrigger>
      <ContextMenuContent>
        {isTrash ? (
          <ContextMenuItem onClick={() => onAction(image)}>
            Restore
          </ContextMenuItem>
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
  onImageSelect: (image: Image) => void
}

export default function ImageGrid({ view, onImageSelect }: ImageGridProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const isTrash = view.type === 'trash'

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: queryKeyFor(view),
    queryFn: fetcherFor(view, getToken),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteImage(getToken, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      toast.success('Image moved to trash')
    },
    onError: () => {
      toast.error('Failed to delete image')
    },
  })

  const restoreMutation = useMutation({
    mutationFn: (id: string) => restoreImage(getToken, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images', 'trash'] })
      toast.success('Image restored')
    },
    onError: () => {
      toast.error('Failed to restore image')
    },
  })

  function handleAction(image: Image) {
    if (isTrash) {
      restoreMutation.mutate(image.id)
    } else {
      deleteMutation.mutate(image.id)
    }
  }

  const allImages = data?.pages.flatMap((p) => p.images) ?? []

  if (isLoading) {
    return (
      <div className="flex items-center justify-center py-20">
        <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
      </div>
    )
  }

  if (allImages.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center py-20 text-muted-foreground gap-3">
        <ImageIcon className="w-10 h-10" />
        <p className="text-sm">{isTrash ? 'Trash is empty' : 'No images here yet'}</p>
      </div>
    )
  }

  return (
    <>
      <div className="columns-2 md:columns-3 lg:columns-4 gap-3">
        {allImages.map((image) => (
          <ImageCard
            key={image.id}
            image={image}
            isTrash={isTrash}
            currentFolderId={view.type === 'folder' ? view.id : null}
            onAction={handleAction}
            onSelect={onImageSelect}
          />
        ))}
      </div>

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

    </>
  )
}
