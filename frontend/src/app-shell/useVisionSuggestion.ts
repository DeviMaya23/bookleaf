import { useCallback, useEffect, useRef } from 'react'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { getMe } from '@/features/auth/lib/me'
import { getImage, acceptSuggestion } from '@/lib/images'

export function useVisionSuggestion(): { checkVision: (imageId: string) => void } {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const timer1 = useRef<ReturnType<typeof setTimeout> | null>(null)
  const timer2 = useRef<ReturnType<typeof setTimeout> | null>(null)

  useEffect(() => {
    return () => {
      if (timer1.current) clearTimeout(timer1.current)
      if (timer2.current) clearTimeout(timer2.current)
    }
  }, [])

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  const showSuggestionToast = useCallback((imageId: string, folderName: string) => {
    toast(`Suggested folder: "${folderName}"`, {
      action: {
        label: 'Accept',
        onClick: async () => {
          try {
            await acceptSuggestion(getToken, imageId, folderName)
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
  }, [getToken, queryClient])

  const checkVision = useCallback((imageId: string) => {
    if (!me?.vision_enabled) return

    timer1.current = setTimeout(async () => {
      try {
        const image = await getImage(getToken, imageId)
        if (image.suggested_folder_name) {
          showSuggestionToast(imageId, image.suggested_folder_name)
          return
        }
      } catch {
        toast.error("Couldn't get folder suggestion")
        return
      }

      timer2.current = setTimeout(async () => {
        try {
          const image = await getImage(getToken, imageId)
          if (image.suggested_folder_name) {
            showSuggestionToast(imageId, image.suggested_folder_name)
          } else {
            toast.error("Couldn't get folder suggestion")
          }
        } catch {
          toast.error("Couldn't get folder suggestion")
        }
      }, 2000)
    }, 1000)
  }, [me, getToken, showSuggestionToast])

  return { checkVision }
}
