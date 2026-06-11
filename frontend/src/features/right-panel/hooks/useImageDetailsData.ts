import { useEffect, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { getImage } from '@/lib/images'
import { getFolders } from '@/lib/folders'
import { getTags } from '@/lib/tags'
import type { Image } from '@/lib/images'
import type { Folder } from '@/lib/folders'

// Bundles the three queries the image panel needs (the live image detail, all
// folders, all tags) and keeps `selectedFolders` in sync with the image's
// folder_ids as either the image or the folder list changes.
export function useImageDetailsData(image: Image) {
  const { getToken } = useKindeAuth()
  const [selectedFolders, setSelectedFolders] = useState<Folder[]>([])

  const { data: imageDetail } = useQuery({
    queryKey: ['image', image.id],
    queryFn: () => getImage(getToken, image.id),
    staleTime: 0,
    refetchInterval: (query) => (!query.state.data?.thumbnail_url ? 1000 : false),
  })

  const { data: allFolders } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  const { data: allTags = [] } = useQuery({
    queryKey: ['tags'],
    queryFn: () => getTags(getToken),
    staleTime: 60_000,
  })

  useEffect(() => {
    if (!allFolders) return
    const resolved = image.folder_ids
      .map((id) => allFolders.find((f) => f.id === id))
      .filter((f): f is Folder => f !== undefined)
    setSelectedFolders(resolved)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [image.id, allFolders])

  return { imageDetail, allFolders, allTags, selectedFolders, setSelectedFolders }
}
