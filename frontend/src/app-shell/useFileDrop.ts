import { useState, useCallback } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { handleFileAutoUpload } from './lib/dragHandlers'
import type { Image } from '@/lib/images'

interface Params {
  folderId: string | null
  onAutoUploaded: (img: Image) => void
  onBatchFilesDropped: (files: File[]) => void
}

export function useFileDrop({ folderId, onAutoUploaded, onBatchFilesDropped }: Params) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [isFileDragOver, setIsFileDragOver] = useState(false)

  const autoUploadMutation = useMutation({
    mutationFn: (file: File) => handleFileAutoUpload(getToken, file, folderId),
    onSuccess: ({ image, duplicates }) => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      if (duplicates.length > 0) {
        toast.warning(`Possible duplicate of "${duplicates[0].title}"`)
      }
      onAutoUploaded(image)
    },
    onError: (err) => {
      if ((err as Error).message === 'heic_safari_only') {
        toast.error('HEIC uploads are only supported in Safari.')
      } else if ((err as Error).message === 'unsupported_type') {
        toast.error('Unsupported file type. Use JPEG, PNG, GIF, WEBP, or AVIF.')
      } else if ((err as Error).message === 'oversized') {
        toast.error('File is too large. Maximum size is 50 MB.')
      } else {
        toast.error('Upload failed. Please try again.')
      }
    },
  })

  const { mutate: autoUpload } = autoUploadMutation

  const dragHandlers = {
    onDragOver: useCallback((e: React.DragEvent<HTMLElement>) => {
      if (e.dataTransfer.types.includes('Files')) {
        e.preventDefault()
        setIsFileDragOver(true)
      }
    }, []),

    onDragLeave: useCallback((e: React.DragEvent<HTMLElement>) => {
      if (e.currentTarget.contains(e.relatedTarget as Node)) return
      setIsFileDragOver(false)
    }, []),

    onDrop: useCallback((e: React.DragEvent<HTMLElement>) => {
      e.preventDefault()
      setIsFileDragOver(false)

      const files = Array.from(e.dataTransfer.files)
      if (files.length === 0) return

      if (files.length > 1) {
        onBatchFilesDropped(files)
        return
      }

      autoUpload(files[0])
    }, [autoUpload, onBatchFilesDropped]),
  }

  return { isFileDragOver, isAutoUploading: autoUploadMutation.isPending, dragHandlers }
}
