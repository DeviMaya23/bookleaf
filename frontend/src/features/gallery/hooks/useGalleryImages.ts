import { useMemo } from 'react'
import { useInfiniteQuery, keepPreviousData } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { getImages, getFolderImages, getAllImages, getTrashedImages } from '@/lib/images'
import type { AppView } from '@/lib/view'
import type { SortBy, SortDir } from './useGalleryControls'

export const EMPTY_FILTER: string[] = []

export function sortParamsFor(sortBy: SortBy, sortDir: SortDir | undefined): { sort?: 'created_at' | 'title' | 'deleted_at'; direction?: SortDir } {
  if (sortBy === 'manual') return {}
  return { sort: sortBy, direction: sortDir }
}

export interface GalleryQueryParams {
  view: AppView
  debouncedSearch: string
  sortBy: SortBy
  sortDir: SortDir | undefined
  filterTagIds: string[]
  filterMimeTypes: string[]
  filterFolderIds: string[]
  searchLabels: boolean
}

export function queryKeyFor({ view, debouncedSearch, sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds, searchLabels }: GalleryQueryParams): unknown[] {
  const sortKey = sortBy === 'manual' ? undefined : [sortBy, sortDir]
  switch (view.type) {
    case 'all': return ['images', 'all', debouncedSearch, sortKey, filterTagIds, filterMimeTypes, filterFolderIds, searchLabels]
    case 'unsorted': return ['images', 'unsorted', debouncedSearch, sortKey, filterTagIds, filterMimeTypes, searchLabels]
    case 'trash': return ['images', 'trash', debouncedSearch, sortKey]
    case 'folder': return ['images', 'folder', view.id, sortKey]
  }
}

export function fetcherFor({ view, getToken, debouncedSearch, sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds, searchLabels }: GalleryQueryParams & { getToken: () => Promise<string | undefined> }) {
  const { sort, direction } = sortParamsFor(sortBy, sortDir)
  return ({ pageParam }: { pageParam: string | undefined }) => {
    switch (view.type) {
      // sortBy never resolves to 'deleted_at' for these views (sortFieldOptions excludes it).
      case 'all': return getAllImages(getToken, pageParam, debouncedSearch, sort as 'created_at' | 'title' | undefined, direction, filterTagIds, filterMimeTypes, filterFolderIds, searchLabels)
      case 'unsorted': return getImages(getToken, pageParam, debouncedSearch, sort as 'created_at' | 'title' | undefined, direction, filterTagIds, filterMimeTypes, undefined, searchLabels)
      // sortBy never resolves to 'created_at' for Trash (sortFieldOptions excludes it).
      case 'trash': return getTrashedImages(getToken, pageParam, debouncedSearch, sort as 'deleted_at' | 'title' | undefined, direction)
      case 'folder': return getFolderImages(getToken, view.id, sort as 'created_at' | 'title' | undefined, direction)
    }
  }
}

export interface UseGalleryImagesParams {
  view: AppView
  searchTerm: string
  debouncedSearchTerm: string
  sortBy: SortBy
  sortDir: SortDir | undefined
  filterTagIds: string[]
  filterMimeTypes: string[]
  filterFolderIds: string[]
  searchLabels: boolean
}

export function useGalleryImages({
  view,
  searchTerm,
  debouncedSearchTerm,
  sortBy,
  sortDir,
  filterTagIds,
  filterMimeTypes,
  filterFolderIds,
  searchLabels,
}: UseGalleryImagesParams) {
  const { getToken } = useKindeAuth()
  const isFolderView = view.type === 'folder'

  const queryParams: GalleryQueryParams = { view, debouncedSearch: debouncedSearchTerm, sortBy, sortDir, filterTagIds, filterMimeTypes, filterFolderIds, searchLabels }

  const { data, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage } = useInfiniteQuery({
    queryKey: queryKeyFor(queryParams),
    queryFn: fetcherFor({ ...queryParams, getToken }),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.next_cursor ?? undefined,
    placeholderData: keepPreviousData,
  })

  const fetchedImages = useMemo(() => data?.pages.flatMap((p) => p.images) ?? [], [data])
  const trimmedSearch = searchTerm.trim().toLowerCase()
  const images = useMemo(() => (
    isFolderView
      ? fetchedImages.filter((img) =>
          (!trimmedSearch || img.title.toLowerCase().includes(trimmedSearch))
          && (filterTagIds.length === 0 || img.tags.some((t) => filterTagIds.includes(t.id)))
          && (filterMimeTypes.length === 0 || filterMimeTypes.includes(img.mime_type))
        )
      : fetchedImages
  ), [fetchedImages, isFolderView, trimmedSearch, filterTagIds, filterMimeTypes])

  return { images, isLoading, fetchNextPage, hasNextPage, isFetchingNextPage }
}
