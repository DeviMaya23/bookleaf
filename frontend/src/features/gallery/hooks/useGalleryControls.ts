import { useCallback, useEffect, useState } from 'react'
import { useDebouncedValue } from '@/hooks/useDebouncedValue'
import { MIME_TYPE_FILTER_OPTIONS } from '@/lib/images'
import type { Folder } from '@/lib/folders'
import type { Tag } from '@/lib/tags'
import type { AppView } from '@/lib/view'

export type SortBy = 'manual' | 'created_at' | 'title' | 'deleted_at'
export type SortDir = 'asc' | 'desc'

export const FIELD_DEFAULT_DIRECTION: Record<'created_at' | 'title' | 'deleted_at', SortDir> = {
  created_at: 'desc',
  title: 'asc',
  deleted_at: 'desc',
}

export function defaultSortForViewType(viewType: AppView['type']): { sortBy: SortBy; sortDir: SortDir | undefined } {
  if (viewType === 'folder') return { sortBy: 'manual', sortDir: undefined }
  if (viewType === 'trash') return { sortBy: 'deleted_at', sortDir: 'desc' }
  return { sortBy: 'created_at', sortDir: 'desc' }
}

export type FilterSection = 'tags' | 'mimeTypes' | 'folders'

export function filterSectionsForViewType(viewType: AppView['type']): FilterSection[] {
  switch (viewType) {
    case 'all': return ['tags', 'mimeTypes', 'folders']
    case 'unsorted': return ['tags', 'mimeTypes']
    case 'folder': return ['tags', 'mimeTypes']
    case 'trash': return []
  }
}

export interface FilterChip {
  key: string
  label: string
  onRemove: () => void
}

export type GalleryControls = ReturnType<typeof useGalleryControls>

export function useGalleryControls(view: AppView, tags: Tag[], folders: Folder[]) {
  const viewKey = view.type === 'folder' ? `folder:${view.id}` : view.type

  const [searchTerm, setSearchTerm] = useState('')
  const debouncedSearchTerm = useDebouncedValue(searchTerm, 300)
  const [sortBy, setSortBy] = useState<SortBy>(() => defaultSortForViewType(view.type).sortBy)
  const [sortDir, setSortDir] = useState<SortDir | undefined>(() => defaultSortForViewType(view.type).sortDir)
  const [filterTagIds, setFilterTagIds] = useState<string[]>([])
  const [filterMimeTypes, setFilterMimeTypes] = useState<string[]>([])
  const [filterFolderIds, setFilterFolderIds] = useState<string[]>([])
  const [filterTagSearch, setFilterTagSearch] = useState('')
  const [filterFolderSearch, setFilterFolderSearch] = useState('')

  const viewDefaultSort = defaultSortForViewType(view.type)
  const sortFieldOptions: SortBy[] = view.type === 'folder'
    ? ['manual', 'created_at', 'title']
    : view.type === 'trash'
      ? ['deleted_at', 'title']
      : ['created_at', 'title']
  const sortActive = sortBy !== viewDefaultSort.sortBy
    || (sortBy !== 'manual' && sortDir !== FIELD_DEFAULT_DIRECTION[sortBy])
  const filterSections = filterSectionsForViewType(view.type)
  const filterCount = filterTagIds.length + filterMimeTypes.length + filterFolderIds.length

  useEffect(() => {
    setSearchTerm('')
    setFilterTagIds([])
    setFilterMimeTypes([])
    setFilterFolderIds([])
    setFilterTagSearch('')
    setFilterFolderSearch('')
  }, [viewKey])

  useEffect(() => {
    const def = defaultSortForViewType(view.type)
    setSortBy(def.sortBy)
    setSortDir(def.sortDir)
  }, [viewKey, view.type])

  const handleSortFieldChange = useCallback((value: SortBy) => {
    setSortBy(value)
    setSortDir(value === 'manual' ? undefined : FIELD_DEFAULT_DIRECTION[value])
  }, [])

  const handleSortDirToggle = useCallback(() => {
    setSortDir((prev) => (prev === 'asc' ? 'desc' : 'asc'))
  }, [])

  const filteredTags = tags.filter((tag) => tag.name.toLowerCase().includes(filterTagSearch.toLowerCase()))
  const filteredFolders = folders.filter((folder) => folder.name.toLowerCase().includes(filterFolderSearch.toLowerCase()))

  const activeFilterChips: FilterChip[] = [
    ...filterTagIds.map((id) => ({
      key: `tag:${id}`,
      label: tags.find((t) => t.id === id)?.name ?? id,
      onRemove: () => setFilterTagIds((prev) => prev.filter((v) => v !== id)),
    })),
    ...filterMimeTypes.map((value) => ({
      key: `mime:${value}`,
      label: MIME_TYPE_FILTER_OPTIONS.find((o) => o.value === value)?.label ?? value,
      onRemove: () => setFilterMimeTypes((prev) => prev.filter((v) => v !== value)),
    })),
    ...filterFolderIds.map((id) => ({
      key: `folder:${id}`,
      label: folders.find((f) => f.id === id)?.name ?? id,
      onRemove: () => setFilterFolderIds((prev) => prev.filter((v) => v !== id)),
    })),
  ]

  const clearAllFilters = useCallback(() => {
    setFilterTagIds([])
    setFilterMimeTypes([])
    setFilterFolderIds([])
  }, [])

  return {
    searchTerm,
    setSearchTerm,
    debouncedSearchTerm,
    sortBy,
    sortDir,
    sortFieldOptions,
    sortActive,
    handleSortFieldChange,
    handleSortDirToggle,
    filterSections,
    filterCount,
    filterTagIds,
    setFilterTagIds,
    filterTagSearch,
    setFilterTagSearch,
    filteredTags,
    filterMimeTypes,
    setFilterMimeTypes,
    filterFolderIds,
    setFilterFolderIds,
    filterFolderSearch,
    setFilterFolderSearch,
    filteredFolders,
    activeFilterChips,
    clearAllFilters,
  }
}
