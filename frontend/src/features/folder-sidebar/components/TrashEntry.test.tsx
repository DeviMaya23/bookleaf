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

vi.mock('@/components/ui/context-menu', () => import('@/test/mocks/contextMenu').then((m) => m.contextMenuMock))

function renderTrashEntry(active: boolean, onClick = vi.fn(), iconsEnabled = true) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return {
    queryClient,
    ...render(
      <QueryClientProvider client={queryClient}>
        <TrashEntry active={active} iconsEnabled={iconsEnabled} onClick={onClick} />
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

  it('renders the trash-2 icon when iconsEnabled is true', () => {
    renderTrashEntry(false, vi.fn(), true)

    expect(screen.getByText('Trash').closest('div')?.querySelector('svg')).toBeTruthy()
  })

  it('renders no icon when iconsEnabled is false', () => {
    renderTrashEntry(false, vi.fn(), false)

    expect(screen.getByText('Trash').closest('div')?.querySelector('svg')).toBeFalsy()
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
