import { useEffect } from 'react'
import { useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { getImage, acceptSuggestion } from '@/lib/images'

const THUMBNAIL_POLL_INTERVAL_MS = 2000
const THUMBNAIL_POLL_MAX_ATTEMPTS = 15 // 30s total (15 attempts × 2s)
const VISION_CHECK_DELAY_MS = 2500

export function usePostUploadFeedback(imageId: string | null, visionEnabled: boolean) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  // Thumbnail polling: check every 2s until thumbnail_url appears (up to 30s).
  // On detection, invalidate the images list so the grid refetches with the real thumbnail.
  useEffect(() => {
    if (!imageId) return

    let attempts = 0
    const id = imageId

    const intervalId = setInterval(async () => {
      attempts++
      if (attempts >= THUMBNAIL_POLL_MAX_ATTEMPTS) {
        clearInterval(intervalId)
        return
      }

      try {
        const image = await getImage(getToken, id)
        if (image.thumbnail_url) {
          clearInterval(intervalId)
          queryClient.invalidateQueries({ queryKey: ['images'] })
        }
      } catch {
        // keep polling
      }
    }, THUMBNAIL_POLL_INTERVAL_MS)

    return () => clearInterval(intervalId)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [imageId])

  // Vision check: single call after 2.5s. Only runs when vision is enabled for the user.
  useEffect(() => {
    if (!imageId || !visionEnabled) return

    const id = imageId
    const timeoutId = setTimeout(async () => {
      try {
        const image = await getImage(getToken, id)
        if (image.suggested_folder_name) {
          const folderName = image.suggested_folder_name
          toast(`Suggested folder: "${folderName}"`, {
            action: {
              label: 'Accept',
              onClick: async () => {
                try {
                  await acceptSuggestion(getToken, id, folderName)
                  queryClient.invalidateQueries({ queryKey: ['images'] })
                  queryClient.invalidateQueries({ queryKey: ['folders'] })
                } catch {
                  toast.error('Failed to accept suggestion')
                }
              },
            },
            cancel: {
              label: 'Ignore',
              onClick: () => {},
            },
          })
        } else {
          toast.error("Couldn't get folder suggestion")
        }
      } catch {
        toast.error("Couldn't get folder suggestion")
      }
    }, VISION_CHECK_DELAY_MS)

    return () => clearTimeout(timeoutId)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [imageId, visionEnabled])
}
