import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { updateTag, deleteTag } from '@/lib/tags'
import type { Tag } from '@/lib/tags'

export function useRenameMutation(onSuccess: () => void) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [duplicateError, setDuplicateError] = useState(false)

  const mutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => updateTag(getToken, id, name),
    onSuccess: (result) => {
      if (result === null) {
        setDuplicateError(true)
        return
      }
      queryClient.invalidateQueries({ queryKey: ['tags'] })
      setDuplicateError(false)
      onSuccess()
    },
  })

  return {
    mutate: (id: string, name: string) => mutation.mutate({ id, name }),
    isPending: mutation.isPending,
    duplicateError,
    clearDuplicateError: () => setDuplicateError(false),
  }
}

export function useDeleteMutation(onSuccess: () => void) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const mutation = useMutation({
    mutationFn: (id: string) => deleteTag(getToken, id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] })
      onSuccess()
    },
  })

  return {
    mutate: (id: string) => mutation.mutate(id),
    isPending: mutation.isPending,
  }
}
