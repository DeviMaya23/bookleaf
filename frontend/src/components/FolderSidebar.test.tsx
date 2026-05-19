import { describe, it, expect } from 'vitest'
import { getFolderSubtreeIds } from './FolderSidebar'
import type { Folder } from '@/lib/folders'

function makeFolder(id: string, parentId: string | null = null): Folder {
  return {
    id,
    name: `Folder ${id}`,
    description: null,
    parent_id: parentId,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

describe('getFolderSubtreeIds', () => {
  it('returns subtree including self and all descendants', () => {
    const folders = [
      makeFolder('root'),
      makeFolder('child-a', 'root'),
      makeFolder('child-b', 'root'),
      makeFolder('grandchild', 'child-a'),
      makeFolder('other'),
    ]

    const result = getFolderSubtreeIds(folders, 'root')

    expect(result.has('root')).toBe(true)
    expect(result.has('child-a')).toBe(true)
    expect(result.has('child-b')).toBe(true)
    expect(result.has('grandchild')).toBe(true)
    expect(result.has('other')).toBe(false)
  })

  it('returns only self when folder has no descendants', () => {
    const folders = [
      makeFolder('parent'),
      makeFolder('leaf', 'parent'),
    ]

    const result = getFolderSubtreeIds(folders, 'leaf')

    expect(result.has('leaf')).toBe(true)
    expect(result.has('parent')).toBe(false)
    expect(result.size).toBe(1)
  })

  it('blocks self-drop: dragged folder ID is in its own subtree', () => {
    const folders = [makeFolder('folder-a')]
    const result = getFolderSubtreeIds(folders, 'folder-a')
    expect(result.has('folder-a')).toBe(true)
  })

  it('blocks descendant-drop: descendant ID is in subtree', () => {
    const folders = [
      makeFolder('a'),
      makeFolder('b', 'a'),
      makeFolder('c', 'b'),
    ]

    const result = getFolderSubtreeIds(folders, 'a')

    expect(result.has('b')).toBe(true)
    expect(result.has('c')).toBe(true)
  })
})
