import { useEffect } from 'react'
import { fetchEventSource } from '@microsoft/fetch-event-source'
import { useQuery, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { getMe } from '@/features/auth/lib/me'

const BASE_URL = import.meta.env.VITE_API_BASE_URL ?? ''

export function useSSEEvents() {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  useEffect(() => {
    if (!me?.ai_categorisation_enabled) return

    const controller = new AbortController()
    let cancelled = false

    const connect = () => {
      if (cancelled) return

      fetchEventSource(`${BASE_URL}/events`, {
        signal: controller.signal,
        fetch: async (input, init) => {
          const token = await getToken()
          return fetch(input, {
            ...init,
            headers: {
              ...init?.headers,
              ...(token ? { Authorization: `Bearer ${token}` } : {}),
            },
          })
        },
        onmessage(msg) {
          try {
            const data = JSON.parse(msg.data)
            if (data.type === 'categorisation_complete') {
              const imageId: string = data.payload?.image_id
              queryClient.invalidateQueries({ queryKey: ['folders'] })
              queryClient.invalidateQueries({ queryKey: ['images'] })
              queryClient.invalidateQueries({ queryKey: ['image', imageId] })
              queryClient.invalidateQueries({ queryKey: ['folder'], exact: false })
              queryClient.invalidateQueries({ queryKey: ['me'] })
            }
          } catch {
            // malformed event — ignore
          }
        },
        onclose() {
          if (!controller.signal.aborted) throw new Error('Server closed the connection')
        },
        onerror(err) {
          if (controller.signal.aborted) return
          if (err instanceof Error && err.message.includes('401')) throw err
        },
      })
    }

    connect()

    return () => {
      cancelled = true
      controller.abort()
    }
  }, [me?.ai_categorisation_enabled, getToken, queryClient])
}
