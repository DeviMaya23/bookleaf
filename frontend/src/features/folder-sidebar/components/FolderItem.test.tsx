import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import FolderItem from './FolderItem'
import type { FolderNode } from '../lib/folderTree'

vi.mock('@/components/ui/context-menu', async () => {
  const React = await import('react')
  return {
    ContextMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuTrigger: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuSeparator: () => React.createElement('hr'),
    ContextMenuItem: ({
      children,
      onClick,
      className,
    }: {
      children: React.ReactNode
      onClick?: (e: React.MouseEvent) => void
      className?: string
    }) =>
      React.createElement('button', { role: 'menuitem', className, onClick }, children),
    ContextMenuSub: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuSubTrigger: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuSubContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
  }
})

function mockPointer(isCoarse: boolean) {
  window.matchMedia = (query: string) =>
    ({
      matches: isCoarse,
      media: query,
      onchange: null,
      addEventListener: () => {},
      removeEventListener: () => {},
      addListener: () => {},
      removeListener: () => {},
      dispatchEvent: () => false,
    }) as MediaQueryList
}

function makeFolder(overrides?: Partial<FolderNode>): FolderNode {
  return {
    id: '1',
    name: 'Nature',
    description: null,
    icon: null,
    parent_id: null,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    children: [],
    ...overrides,
  }
}

function renderFolderItem(props: Partial<React.ComponentProps<typeof FolderItem>> = {}) {
  const folder = props.folder ?? makeFolder()
  return render(
    <FolderItem
      folder={folder}
      depth={0}
      view={{ type: 'all' }}
      folders={[folder]}
      activeDragType={null}
      activeDragFolderId={null}
      iconsEnabled={false}
      onSelect={props.onSelect ?? vi.fn()}
      onRename={props.onRename ?? vi.fn()}
      onDelete={props.onDelete ?? vi.fn()}
      onNewSubfolder={props.onNewSubfolder ?? vi.fn()}
      onChangeIcon={props.onChangeIcon ?? vi.fn()}
      onViewDetails={props.onViewDetails}
    />,
  )
}

describe('FolderItem — View details context menu item', () => {
  const originalMatchMedia = window.matchMedia

  beforeEach(() => {
    vi.clearAllMocks()
  })

  afterEach(() => {
    window.matchMedia = originalMatchMedia
  })

  it('shows "View details" above "New subfolder" on a coarse-pointer device', () => {
    mockPointer(true)

    renderFolderItem()

    const items = screen.getAllByRole('menuitem').map((el) => el.textContent)
    expect(items.indexOf('View details')).toBeGreaterThanOrEqual(0)
    expect(items.indexOf('View details')).toBeLessThan(items.indexOf('New subfolder'))
  })

  it('does not show "View details" on a fine-pointer device', () => {
    mockPointer(false)

    renderFolderItem()

    expect(screen.queryByRole('menuitem', { name: 'View details' })).not.toBeInTheDocument()
  })

  it('calls onViewDetails with the correct folder when selected', async () => {
    mockPointer(true)
    const folder = makeFolder({ id: 'folder-2', name: 'Travel' })
    const onViewDetails = vi.fn()

    renderFolderItem({ folder, onViewDetails })

    await userEvent.click(screen.getByRole('menuitem', { name: 'View details' }))

    expect(onViewDetails).toHaveBeenCalledWith(folder)
  })
})
