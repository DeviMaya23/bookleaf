import { describe, it, expect, vi, beforeEach } from 'vitest'
import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { emptyTrash } from '@/lib/images'
import TrashEntry from './TrashEntry'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getToken: vi.fn().mockResolvedValue('token'),
  }),
}))

vi.mock('@/lib/images', async (importOriginal) => ({
  ...(await importOriginal<typeof import('@/lib/images')>()),
  emptyTrash: vi.fn(),
}))

vi.mock('@/components/ui/context-menu', async () => {
  const React = await import('react')
  return {
    ContextMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuTrigger: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    ContextMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
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
  }
})

function renderTrashEntry(active: boolean, onClick = vi.fn()) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return {
    queryClient,
    ...render(
      <QueryClientProvider client={queryClient}>
        <TrashEntry active={active} onClick={onClick} />
      </QueryClientProvider>,
    ),
  }
}

describe('TrashEntry', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('applies active styling when active is true', () => {
    renderTrashEntry(true)

    expect(screen.getByText('Trash')).toHaveClass('bg-accent', 'text-accent-foreground', 'font-medium')
  })

  it('does not apply active styling when active is false', () => {
    renderTrashEntry(false)

    expect(screen.getByText('Trash')).not.toHaveClass('bg-accent', 'text-accent-foreground', 'font-medium')
  })

  it('shows Empty trash item in the context menu', () => {
    renderTrashEntry(false)

    expect(screen.getByRole('menuitem', { name: /empty trash/i })).toBeInTheDocument()
  })

  it('calls emptyTrash and invalidates the trash query after confirming the dialog', async () => {
    vi.mocked(emptyTrash).mockResolvedValue(undefined)

    const { queryClient } = renderTrashEntry(false)
    const invalidateSpy = vi.spyOn(queryClient, 'invalidateQueries')

    await userEvent.click(screen.getByRole('menuitem', { name: /empty trash/i }))

    const confirmButton = await screen.findByRole('button', { name: /empty trash/i })
    await userEvent.click(confirmButton)

    await waitFor(() => {
      expect(emptyTrash).toHaveBeenCalledWith(expect.any(Function))
    })
    await waitFor(() => {
      expect(invalidateSpy).toHaveBeenCalledWith({ queryKey: ['images', 'trash'] })
    })
  })
})
