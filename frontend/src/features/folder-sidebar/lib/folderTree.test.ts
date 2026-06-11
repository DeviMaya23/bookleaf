import { describe, it, expect } from 'vitest'
import type { Folder } from '@/lib/folders'
import { buildFolderTree, filterFolderTree, type FolderNode } from './folderTree'

function makeFolder(id: string, parentId: string | null = null, name?: string): Folder {
  return {
    id,
    name: name ?? `Folder ${id}`,
    description: null,
    parent_id: parentId,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  }
}

describe('buildFolderTree', () => {
  it('nests children under their parents and keeps parentless folders as roots', () => {
    const tree = buildFolderTree([
      makeFolder('root'),
      makeFolder('child-a', 'root'),
      makeFolder('grandchild', 'child-a'),
      makeFolder('other-root'),
    ])

    expect(tree.map((n) => n.id)).toEqual(['root', 'other-root'])
    const root = tree[0]
    expect(root.children.map((n) => n.id)).toEqual(['child-a'])
    expect(root.children[0].children.map((n) => n.id)).toEqual(['grandchild'])
  })

  it('drops folders whose parent_id does not resolve to a known folder', () => {
    const tree = buildFolderTree([
      makeFolder('root'),
      makeFolder('orphan', 'missing-parent'),
    ])

    expect(tree.map((n) => n.id)).toEqual(['root'])
  })
})

describe('filterFolderTree', () => {
  const tree: FolderNode[] = buildFolderTree([
    makeFolder('animals', null, 'Animals'),
    makeFolder('cats', 'animals', 'Cats'),
    makeFolder('plants', null, 'Plants'),
  ]) as FolderNode[]

  it('keeps a matching node and its ancestors, case-insensitively', () => {
    const result = filterFolderTree(tree, 'cat')

    expect(result.map((n) => n.id)).toEqual(['animals'])
    expect(result[0].children.map((n) => n.id)).toEqual(['cats'])
  })

  it('keeps a matching parent but prunes children that do not match', () => {
    const result = filterFolderTree(tree, 'animals')

    expect(result.map((n) => n.id)).toEqual(['animals'])
    expect(result[0].children).toEqual([])
  })

  it('returns an empty tree when nothing matches', () => {
    expect(filterFolderTree(tree, 'zzz')).toEqual([])
  })
})
