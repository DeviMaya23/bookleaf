import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { getFolders, getFolderSubtreeIds } from '@/lib/folders'
import { buildFolderTree, filterFolderTree } from '../lib/folderTree'

export function useFolderTree(activeDragFolderId: string | null) {
  const { getToken } = useKindeAuth()
  const [folderFilter, setFolderFilter] = useState('')

  const { data: folders = [], isPending: isFoldersPending } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
  })

  const subtreeIds = activeDragFolderId
    ? getFolderSubtreeIds(folders, activeDragFolderId)
    : new Set<string>()

  const trimmedFilter = folderFilter.trim()
  const tree = buildFolderTree(folders)
  const visibleTree = trimmedFilter ? filterFolderTree(tree, trimmedFilter) : tree

  return { folders, isFoldersPending, folderFilter, setFolderFilter, subtreeIds, visibleTree }
}
