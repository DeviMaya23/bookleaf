import { describe, it, expect } from 'vitest'
import { act, renderHook } from '@testing-library/react'
import type { Folder } from '@/lib/folders'
import type { Tag } from '@/lib/tags'
import type { AppView } from '@/lib/view'
import { useGalleryControls } from './useGalleryControls'

const tags: Tag[] = [
  { id: 'tag-1', name: 'Cats' },
  { id: 'tag-2', name: 'Dogs' },
]
const folders: Folder[] = [
  { id: 'folder-1', name: 'Vacation', description: null, parent_id: null, created_at: '', updated_at: '' },
  { id: 'folder-2', name: 'Work', description: null, parent_id: null, created_at: '', updated_at: '' },
]

function render(view: AppView) {
  return renderHook(({ v }) => useGalleryControls(v, tags, folders), { initialProps: { v: view } })
}

describe('useGalleryControls defaults', () => {
  it('defaults a non-folder view to Date added / newest first with no manual option', () => {
    const { result } = render({ type: 'all' })

    expect(result.current.sortBy).toBe('created_at')
    expect(result.current.sortDir).toBe('desc')
    expect(result.current.sortActive).toBe(false)
    expect(result.current.sortFieldOptions).toEqual(['created_at', 'title'])
    expect(result.current.filterSections).toEqual(['tags', 'mimeTypes', 'folders'])
  })

  it('defaults a folder view to Manual sort with a manual option and no folder filter section', () => {
    const { result } = render({ type: 'folder', id: 'folder-1' })

    expect(result.current.sortBy).toBe('manual')
    expect(result.current.sortDir).toBeUndefined()
    expect(result.current.sortFieldOptions).toEqual(['manual', 'created_at', 'title'])
    expect(result.current.filterSections).toEqual(['tags', 'mimeTypes'])
  })
})

describe('useGalleryControls sort handlers', () => {
  it('changing the sort field applies its default direction and marks sort active', () => {
    const { result } = render({ type: 'all' })

    act(() => result.current.handleSortFieldChange('title'))

    expect(result.current.sortBy).toBe('title')
    expect(result.current.sortDir).toBe('asc')
    expect(result.current.sortActive).toBe(true)
  })

  it('selecting manual clears the direction', () => {
    const { result } = render({ type: 'folder', id: 'folder-1' })

    act(() => result.current.handleSortFieldChange('created_at'))
    expect(result.current.sortDir).toBe('desc')

    act(() => result.current.handleSortFieldChange('manual'))
    expect(result.current.sortDir).toBeUndefined()
  })

  it('toggling the direction flips asc/desc', () => {
    const { result } = render({ type: 'all' })

    expect(result.current.sortDir).toBe('desc')
    act(() => result.current.handleSortDirToggle())
    expect(result.current.sortDir).toBe('asc')
  })
})

describe('useGalleryControls filters', () => {
  it('builds removable chips labelled from tags, folders, and mime types', () => {
    const { result } = render({ type: 'all' })

    act(() => {
      result.current.setFilterTagIds(['tag-1'])
      result.current.setFilterFolderIds(['folder-2'])
      result.current.setFilterMimeTypes(['image/jpeg'])
    })

    expect(result.current.filterCount).toBe(3)
    const labels = result.current.activeFilterChips.map((c) => c.label)
    expect(labels).toEqual(expect.arrayContaining(['Cats', 'Work', 'JPEG']))

    act(() => result.current.activeFilterChips.find((c) => c.label === 'Cats')!.onRemove())
    expect(result.current.filterTagIds).toEqual([])
  })

  it('clearAllFilters clears tags, folders, and mime types', () => {
    const { result } = render({ type: 'all' })

    act(() => {
      result.current.setFilterTagIds(['tag-1'])
      result.current.setFilterMimeTypes(['image/jpeg'])
    })
    act(() => result.current.clearAllFilters())

    expect(result.current.filterCount).toBe(0)
  })

  it('filteredTags and filteredFolders narrow by their search terms independently', () => {
    const { result } = render({ type: 'all' })

    act(() => result.current.setFilterTagSearch('Ca'))

    expect(result.current.filteredTags.map((t) => t.name)).toEqual(['Cats'])
    expect(result.current.filteredFolders).toHaveLength(2)
  })
})

describe('useGalleryControls view changes', () => {
  it('resets search, filters, and sort when the view changes', () => {
    const { result, rerender } = render({ type: 'folder', id: 'folder-1' })

    act(() => {
      result.current.setSearchTerm('beach')
      result.current.setFilterTagIds(['tag-1'])
      result.current.handleSortFieldChange('title')
    })
    expect(result.current.sortBy).toBe('title')

    rerender({ v: { type: 'all' } })

    expect(result.current.searchTerm).toBe('')
    expect(result.current.filterTagIds).toEqual([])
    expect(result.current.sortBy).toBe('created_at')
    expect(result.current.sortDir).toBe('desc')
  })
})
