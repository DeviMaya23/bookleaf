import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route, Link } from 'react-router-dom'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import AppLayout from './AppLayout'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/folders', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/folders')>()),
  getFolders: vi.fn().mockResolvedValue([]),
}))

vi.mock('@/hooks/useVisionSuggestion', () => ({
  useVisionSuggestion: () => ({ checkVision: vi.fn() }),
}))

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
  }
})

vi.mock('@/components/ui/scroll-area', async () => {
  const React = await import('react')
  return {
    ScrollArea: ({ children, className }: { children: React.ReactNode; className?: string }) =>
      React.createElement('div', { className }, children),
  }
})

vi.mock('./FolderSidebar', () => ({
  default: () => <div data-testid="folder-sidebar" />,
}))

vi.mock('./ImageViewer', () => ({
  default: () => null,
}))

vi.mock('./UploadModal', () => ({
  default: () => null,
}))

vi.mock('./BatchUploadModal', () => ({
  default: () => null,
}))

vi.mock('./RightPanel', () => ({
  default: () => null,
}))

vi.mock('./ImageGrid', () => ({
  default: ({ sortBy, sortDir }: { sortBy: string; sortDir: string | undefined }) => (
    <div data-testid="image-grid" data-sort-by={sortBy} data-sort-dir={sortDir ?? ''} />
  ),
}))

function renderApp(initialPath: string) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[initialPath]}>
        <nav>
          <Link to="/">All</Link>
          <Link to="/folders/folder-1">Folder 1</Link>
          <Link to="/folders/folder-2">Folder 2</Link>
        </nav>
        <Routes>
          <Route path="/" element={<AppLayout />} />
          <Route path="/unsorted" element={<AppLayout />} />
          <Route path="/trash" element={<AppLayout />} />
          <Route path="/folders/:folderId" element={<AppLayout />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function imageGrid() {
  return screen.getByTestId('image-grid')
}

describe('AppLayout sort control', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('switching from a folder with an explicit sort to All resets to Date added / newest first', async () => {
    renderApp('/folders/folder-1')

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))

    await userEvent.click(screen.getByText('All'))

    await waitFor(() => {
      expect(imageGrid()).toHaveAttribute('data-sort-by', 'created_at')
      expect(imageGrid()).toHaveAttribute('data-sort-dir', 'desc')
    })

    expect(screen.getByRole('menuitemradio', { name: 'Date added' })).toHaveAttribute('aria-checked', 'true')
    expect(screen.getByText('Newest first')).toBeInTheDocument()
  })

  it('switching between two folders resets sort to Manual', async () => {
    renderApp('/folders/folder-1')

    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))

    await userEvent.click(screen.getByText('Folder 2'))

    await waitFor(() => {
      expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual')
      expect(imageGrid()).toHaveAttribute('data-sort-dir', '')
    })

    expect(screen.getByRole('menuitemradio', { name: 'Manual' })).toHaveAttribute('aria-checked', 'true')
  })

  it('hides the direction toggle for Manual and shows it with field-appropriate labels otherwise', async () => {
    renderApp('/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    expect(screen.queryByText(/oldest first|newest first|a → z|z → a/i)).not.toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Date added' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'created_at'))
    expect(screen.getByText('Newest first')).toBeInTheDocument()

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))
    expect(screen.getByText('A → Z')).toBeInTheDocument()
  })

  it('shows the sort trigger as active only when the selection differs from the view default', async () => {
    renderApp('/folders/folder-1')
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'manual'))

    const trigger = screen.getByRole('button', { name: /sort/i })
    expect(trigger.className).not.toMatch(/bg-primary/)

    await userEvent.click(screen.getByRole('menuitemradio', { name: 'Name' }))
    await waitFor(() => expect(imageGrid()).toHaveAttribute('data-sort-by', 'title'))

    expect(trigger.className).toMatch(/bg-primary/)
  })
})
