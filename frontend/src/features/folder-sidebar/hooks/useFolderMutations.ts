import { toast } from 'sonner'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { createFolder, updateFolder, deleteFolder } from '@/lib/folders'

export function useFolderMutations() {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const invalidate = () => queryClient.invalidateQueries({ queryKey: ['folders'] })

  const createMutation = useMutation({
    mutationFn: ({ name, parentId }: { name: string; parentId?: string }) =>
      createFolder(getToken, name, parentId),
    onSuccess: invalidate,
  })

  const renameMutation = useMutation({
    mutationFn: ({ id, name }: { id: string; name: string }) => updateFolder(getToken, id, { name }),
    onSuccess: invalidate,
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteFolder(getToken, id),
    onSuccess: () => {
      invalidate()
      toast.success('Folder deleted')
    },
    onError: () => {
      toast.error('Failed to delete folder')
    },
  })

  return { createMutation, renameMutation, deleteMutation }
}
