import { useState } from 'react'
import { useInfiniteQuery, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Loader2, ImageIcon } from 'lucide-react'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog'
import { Button } from '@/components/ui/button'
import { toast } from 'sonner'
import { getImages, deleteImage, getImage } from '@/lib/images'
import type { Image } from '@/lib/images'

interface ImageCardProps {
  image: Image
  onDelete: (image: Image) => void
  onOpen: (image: Image) => void
}

function ImageCard({ image, onDelete, onOpen }: ImageCardProps) {
  return (
    <ContextMenu>
      <ContextMenuTrigger>
        <div className="cursor-pointer rounded-lg overflow-hidden border bg-card" onClick={() => onOpen(image)}>
          <div className="aspect-square bg-muted">
            {image.thumbnail_url ? (
              <img
                src={image.thumbnail_url}
                alt={image.title}
                className="w-full h-full object-cover"
              />
            ) : (
              <div className="w-full h-full flex items-center justify-center">
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
        <ContextMenuItem
          onClick={() => onDelete(image)}
          className="text-destructive focus:text-destructive"
        >
          Delete
        </ContextMenuItem>
      </ContextMenuContent>
    </ContextMenu>
  )
}

interface ImageGridProps {
  folderId: string | null
}

export default function ImageGrid({ folderId }: ImageGridProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [deleteTarget, setDeleteTarget] = useState<Image | null>(null)
  const [lightboxTarget, setLightboxTarget] = useState<Image | null>(null)

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: ['images', folderId],
    queryFn: ({ pageParam }) => getImages(getToken, folderId, pageParam as string | undefined),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
  })

  const { data: imageDetail, isLoading: isLoadingDetail } = useQuery({
    queryKey: ['image', lightboxTarget?.id],
    queryFn: () => getImage(getToken, lightboxTarget!.id),
    enabled: !!lightboxTarget,
    staleTime: 0,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteImage(getToken, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      setDeleteTarget(null)
      toast.success('Image deleted')
    },
    onError: () => {
      toast.error('Failed to delete image')
    },
  })

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
        <p className="text-sm">No images here yet</p>
      </div>
    )
  }

  return (
    <>
      <div className="grid grid-cols-2 lg:grid-cols-6 gap-4">
        {allImages.map((image) => (
          <ImageCard key={image.id} image={image} onDelete={setDeleteTarget} onOpen={setLightboxTarget} />
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

      <Dialog open={!!lightboxTarget} onOpenChange={(open) => { if (!open) setLightboxTarget(null) }}>
        <DialogContent className="sm:max-w-fit p-0 overflow-hidden">
          <DialogTitle className="sr-only">{lightboxTarget?.title}</DialogTitle>
          {isLoadingDetail ? (
            <div className="flex items-center justify-center w-64 h-64">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
            </div>
          ) : imageDetail ? (
            <img
              src={imageDetail.image_url}
              alt={lightboxTarget?.title}
              className="max-h-[90vh] max-w-[90vw] object-contain"
            />
          ) : null}
        </DialogContent>
      </Dialog>

      <Dialog open={!!deleteTarget} onOpenChange={(open) => { if (!open) setDeleteTarget(null) }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>Delete image</DialogTitle>
          </DialogHeader>
          <p className="text-sm text-muted-foreground">
            Are you sure you want to delete{' '}
            <span className="font-medium text-foreground">"{deleteTarget?.title}"</span>? This
            cannot be undone.
          </p>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteTarget(null)}>
              Cancel
            </Button>
            <Button
              variant="destructive"
              onClick={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
              disabled={deleteMutation.isPending}
            >
              Delete
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </>
  )
}
