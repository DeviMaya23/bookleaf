import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, beforeEach, vi } from 'vitest'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { ThemeProvider } from '@/hooks/useTheme'
import AppSection from './AppSection'

const mockGetMe = vi.fn()
const mockUpdateMe = vi.fn()
const mockToastError = vi.fn()

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getToken: vi.fn().mockResolvedValue('token'),
  }),
}))

vi.mock('@/features/auth/lib/me', () => ({
  getMe: () => mockGetMe(),
  updateMe: (...args: unknown[]) => mockUpdateMe(...args),
}))

vi.mock('sonner', () => ({
  toast: {
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

function renderWithProviders(ui: React.ReactElement) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  return render(
    <QueryClientProvider client={queryClient}>
      <ThemeProvider>{ui}</ThemeProvider>
    </QueryClientProvider>,
  )
}

describe('AppSection', () => {
  beforeEach(() => {
    localStorage.clear()
    document.documentElement.removeAttribute('data-theme')
    vi.clearAllMocks()
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })
  })

  it('shows the Parchment theme as selected, reflecting useTheme()', () => {
    renderWithProviders(<AppSection />)

    expect(screen.getByText('Parchment')).toBeInTheDocument()
    expect(screen.getByText('Warm parchment')).toBeInTheDocument()
    expect(screen.getByRole('radio', { name: 'Parchment theme' })).toBeChecked()
  })

  it('renders all three theme rows with the correct selected state', () => {
    renderWithProviders(<AppSection />)

    expect(screen.getByRole('radio', { name: 'Parchment theme' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Lumen theme' })).not.toBeChecked()
    expect(screen.getByRole('radio', { name: 'Sunless theme' })).not.toBeChecked()
  })

  it('calls setTheme when an unselected row is clicked', async () => {
    const user = userEvent.setup()
    renderWithProviders(<AppSection />)

    await user.click(screen.getByRole('radio', { name: 'Lumen theme' }))

    expect(screen.getByRole('radio', { name: 'Lumen theme' })).toBeChecked()
    expect(screen.getByRole('radio', { name: 'Parchment theme' })).not.toBeChecked()
  })

  it('renders the folder icons switch as on when folder_icons_enabled is true', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })

    renderWithProviders(<AppSection />)

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })
  })

  it('renders the folder icons switch as off when folder_icons_enabled is false', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: false })

    renderWithProviders(<AppSection />)

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
  })

  it('disables folder icons via PATCH /me and updates the displayed state', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })
    mockUpdateMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: false })

    renderWithProviders(<AppSection />)

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { folder_icons_enabled: false })
    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
  })

  it('enables folder icons via PATCH /me and updates the displayed state', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: false })
    mockUpdateMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })

    renderWithProviders(<AppSection />)

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByRole('switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { folder_icons_enabled: true })
    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })
  })

  it('disables the folder icons switch while the update is pending', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })
    let resolveUpdate: (value: { id: string; vision_enabled: boolean; folder_icons_enabled: boolean }) => void = () => {}
    mockUpdateMe.mockImplementation(
      () => new Promise((resolve) => { resolveUpdate = resolve }),
    )

    renderWithProviders(<AppSection />)

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('switch'))

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('aria-disabled', 'true')
    })

    resolveUpdate({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: false })

    await waitFor(() => {
      expect(screen.getByRole('switch')).not.toHaveAttribute('aria-disabled')
    })
  })

  it('shows an error toast and leaves the folder icons toggle unchanged when the update fails', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_abc123', vision_enabled: false, folder_icons_enabled: true })
    mockUpdateMe.mockRejectedValue(new Error('boom'))

    renderWithProviders(<AppSection />)

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('switch'))

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Failed to update settings')
    })
    expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    expect(screen.getByRole('switch')).not.toHaveAttribute('aria-disabled')
  })
})
