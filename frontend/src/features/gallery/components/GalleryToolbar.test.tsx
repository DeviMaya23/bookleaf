import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect } from 'vitest'
import type { Folder } from '@/lib/folders'
import type { Tag } from '@/lib/tags'
import type { AppView } from '@/lib/view'
import GalleryToolbar from './GalleryToolbar'
import { useGalleryControls } from '../hooks/useGalleryControls'

vi.mock('@/components/ui/dropdown-menu', async () => {
  const React = await import('react')
  const RadioGroupContext = React.createContext<{ value?: string; onValueChange?: (value: string) => void }>({})

  return {
    DropdownMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DropdownMenuTrigger: ({ children, className, ...props }: { children: React.ReactNode; className?: string }) =>
      React.createElement('button', { className, ...props }, children),
    DropdownMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { 'data-testid': 'dropdown-content' }, children),
    DropdownMenuGroup: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DropdownMenuItem: ({ children, onClick }: { children: React.ReactNode; onClick?: () => void }) =>
      React.createElement('button', { role: 'menuitem', onClick }, children),
    DropdownMenuSeparator: () => React.createElement('hr'),
    DropdownMenuRadioGroup: ({ children, value, onValueChange }: { children: React.ReactNode; value?: string; onValueChange?: (value: string) => void }) =>
      React.createElement(RadioGroupContext.Provider, { value: { value, onValueChange } }, children),
    DropdownMenuRadioItem: ({ children, value }: { children: React.ReactNode; value: string }) => {
      const ctx = React.useContext(RadioGroupContext)
      return React.createElement('button', {
        role: 'menuitemradio',
        'aria-checked': ctx.value === value,
        onClick: () => ctx.onValueChange?.(value),
      }, children)
    },
    DropdownMenuLabel: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', null, children),
    DropdownMenuCheckboxItem: ({ children, checked, onCheckedChange }: { children: React.ReactNode; checked?: boolean; onCheckedChange?: (checked: boolean) => void }) =>
      React.createElement('button', {
        role: 'menuitemcheckbox',
        'aria-checked': !!checked,
        onClick: () => onCheckedChange?.(!checked),
      }, children),
  }
})

function makeFolder(id: string, name: string): Folder {
  return { id, name, description: null, icon: null, parent_id: null, created_at: '', updated_at: '' }
}

function Harness({ view, tags, folders }: { view: AppView; tags: Tag[]; folders: Folder[] }) {
  const controls = useGalleryControls(view, tags, folders)
  return (
    <>
      <GalleryToolbar
        view={view}
        controls={controls}
        focusToggle={<button aria-label="Focus mode">Focus</button>}
        uploadActions={<button>Image</button>}
      />
      <div
        data-testid="grid"
        data-sort-by={controls.sortBy}
        data-sort-dir={controls.sortDir ?? ''}
        data-filter-tag-ids={controls.filterTagIds.join(',')}
      />
    </>
  )
}

function renderToolbar(view: AppView, { tags = [], folders = [] }: { tags?: Tag[]; folders?: Folder[] } = {}) {
  return render(<Harness view={view} tags={tags} folders={folders} />)
}

function grid() {
  return screen.getByTestId('grid')
}

describe('GalleryToolbar sort control', () => {
  it('hides the direction toggle for Manual and shows field-appropriate labels otherwise', async () => {
    renderToolbar({ type: 'folder', id: 'folder-1' })
    expect(grid()).toHaveAttribute('data-sort-by', 'manual')
    expect(screen.queryByText(/oldest first|newest first|a → z|z → a/i)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Date added' }))
    expect(grid()).toHaveAttribute('data-sort-by', 'created_at')
    expect(screen.getByText('Newest first')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    expect(grid()).toHaveAttribute('data-sort-by', 'title')
    expect(screen.getByText('A → Z')).toBeInTheDocument()
  })

  it('shows the sort trigger as active only when the selection differs from the view default', async () => {
    renderToolbar({ type: 'folder', id: 'folder-1' })

    const trigger = screen.getByRole('button', { name: /sort/i })
    expect(trigger.className).not.toMatch(/bg-secondary/)

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))

    expect(trigger.className).toMatch(/bg-secondary/)
  })
})

describe('GalleryToolbar mobile visibility', () => {
  it('hides the sort control and upload actions container below sm', () => {
    renderToolbar({ type: 'folder', id: 'folder-1' })

    const sortTrigger = screen.getByRole('button', { name: /sort/i })
    expect(sortTrigger.parentElement?.className).toMatch(/hidden sm:flex/)

    const uploadActionsContainer = screen.getByText('Image').closest('div')
    expect(uploadActionsContainer?.className).toMatch(/hidden sm:flex/)
  })
})

describe('GalleryToolbar filter control', () => {
  it('hides the Filters button in Trash', () => {
    renderToolbar({ type: 'trash' })
    expect(screen.queryByRole('button', { name: /filters/i })).not.toBeInTheDocument()
  })

  it('shows File type, Tags, and Folder sections, in that order, in the All view', async () => {
    renderToolbar({ type: 'all' }, { tags: [{ id: 'tag-1', name: 'Cats' }], folders: [makeFolder('folder-1', 'Vacation')] })

    expect(screen.getByText('File type')).toBeInTheDocument()
    expect(screen.getByText('Tags')).toBeInTheDocument()
    expect(screen.getByText('Folder')).toBeInTheDocument()
    const filterPanel = screen.getAllByTestId('dropdown-content').find((el) => el.textContent?.includes('File type'))
    const panelText = filterPanel?.textContent ?? ''
    expect(panelText.indexOf('File type')).toBeLessThan(panelText.indexOf('Tags'))
    expect(panelText.indexOf('Tags')).toBeLessThan(panelText.indexOf('Folder'))
    expect(screen.getByRole('button', { name: 'JPEG' })).toBeInTheDocument()
    expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument()
  })

  it('shows only Tags and File type sections in a folder view', () => {
    renderToolbar({ type: 'folder', id: 'folder-1' }, { tags: [{ id: 'tag-1', name: 'Cats' }], folders: [makeFolder('folder-1', 'Vacation')] })

    expect(screen.getByText('Tags')).toBeInTheDocument()
    expect(screen.getByText('File type')).toBeInTheDocument()
    expect(screen.queryByText('Folder')).not.toBeInTheDocument()
  })

  it('shows a badge with the total filter count and switches to the active variant', async () => {
    renderToolbar({ type: 'all' }, { tags: [{ id: 'tag-1', name: 'Cats' }, { id: 'tag-2', name: 'Dogs' }] })

    const filtersButton = screen.getByRole('button', { name: /filters/i })
    expect(filtersButton.className).not.toMatch(/bg-secondary/)
    expect(screen.queryByText('3')).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Dogs' }))
    await userEvent.click(screen.getByRole('button', { name: 'JPEG' }))

    await waitFor(() => expect(screen.getByText('3')).toBeInTheDocument())
    expect(filtersButton.className).toMatch(/bg-secondary/)
  })

  it('searching tags filters the tag list without affecting the folder list', async () => {
    renderToolbar({ type: 'all' }, {
      tags: [{ id: 'tag-1', name: 'Cats' }, { id: 'tag-2', name: 'Dogs' }],
      folders: [makeFolder('folder-1', 'Vacation')],
    })

    await userEvent.type(screen.getByPlaceholderText('Search tags…'), 'Ca')

    expect(screen.getByRole('menuitemcheckbox', { name: 'Cats' })).toBeInTheDocument()
    expect(screen.queryByRole('menuitemcheckbox', { name: 'Dogs' })).not.toBeInTheDocument()
    expect(screen.getByRole('menuitemcheckbox', { name: 'Vacation' })).toBeInTheDocument()
  })

  it('hides all tags when the search matches none, even a checked tag, while keeping it active', async () => {
    renderToolbar({ type: 'all' }, { tags: [{ id: 'tag-1', name: 'Cats' }, { id: 'tag-2', name: 'Dogs' }] })

    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))
    await waitFor(() => expect(grid()).toHaveAttribute('data-filter-tag-ids', 'tag-1'))

    await userEvent.type(screen.getByPlaceholderText('Search tags…'), 'zzz')

    expect(screen.queryByRole('menuitemcheckbox', { name: 'Cats' })).not.toBeInTheDocument()
    expect(screen.queryByRole('menuitemcheckbox', { name: 'Dogs' })).not.toBeInTheDocument()
    expect(grid()).toHaveAttribute('data-filter-tag-ids', 'tag-1')
    expect(screen.getByRole('button', { name: 'Remove filter Cats' })).toBeInTheDocument()
  })

  it('removes a single filter via its chip and clears all via Clear all', async () => {
    renderToolbar({ type: 'all' }, { tags: [{ id: 'tag-1', name: 'Cats' }, { id: 'tag-2', name: 'Dogs' }] })

    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Cats' }))
    await userEvent.click(screen.getByRole('menuitemcheckbox', { name: 'Dogs' }))
    await waitFor(() => expect(grid()).toHaveAttribute('data-filter-tag-ids', 'tag-1,tag-2'))

    await userEvent.click(screen.getByRole('button', { name: 'Remove filter Cats' }))
    await waitFor(() => expect(grid()).toHaveAttribute('data-filter-tag-ids', 'tag-2'))

    await userEvent.click(screen.getByRole('button', { name: 'Clear all' }))
    await waitFor(() => expect(grid()).toHaveAttribute('data-filter-tag-ids', ''))
    expect(screen.queryByText('Clear all')).not.toBeInTheDocument()
  })
})
