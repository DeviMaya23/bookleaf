import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { deleteImage, hardDeleteImage, restoreImage } from '@/lib/images'
import type { Image } from '@/lib/images'

export function useImageLifecycle(isTrash: boolean, removeImage: (id: string) => void, onImageDeleted?: (id: string) => void) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [confirmDeleteImage, setConfirmDeleteImage] = useState<Image | null>(null)

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteImage(getToken, id),
    onSuccess: (_, id) => {
      removeImage(id)
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

  const hardDeleteMutation = useMutation({
    mutationFn: (id: string) => hardDeleteImage(getToken, id),
    onSuccess: (_, id) => {
      removeImage(id)
      queryClient.invalidateQueries({ queryKey: ['images', 'trash'] })
      toast.success('Image permanently deleted')
      onImageDeleted?.(id)
    },
    onError: () => toast.error('Failed to permanently delete image'),
  })

  function handleAction(image: Image) {
    if (isTrash) {
      restoreMutation.mutate(image.id)
    } else {
      deleteMutation.mutate(image.id)
    }
  }

  function closeDeleteConfirm() {
    setConfirmDeleteImage(null)
  }

  function confirmDelete() {
    if (confirmDeleteImage) {
      hardDeleteMutation.mutate(confirmDeleteImage.id)
      setConfirmDeleteImage(null)
    }
  }

  return {
    handleAction,
    confirmDeleteImage,
    openDeleteConfirm: setConfirmDeleteImage,
    closeDeleteConfirm,
    confirmDelete,
  }
}
