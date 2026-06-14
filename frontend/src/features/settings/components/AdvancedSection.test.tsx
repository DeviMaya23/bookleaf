import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AdvancedSection from './AdvancedSection'

const mockGetMe = vi.fn()
const mockUpdateMe = vi.fn()
const mockToastError = vi.fn()

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getToken: vi.fn(),
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

function renderAdvancedSection() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <AdvancedSection />
    </QueryClientProvider>,
  )
}

describe('AdvancedSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the AI features toggle as on when vision_enabled is true', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })
    expect(screen.getByText('Active — folder suggestions on upload')).toBeInTheDocument()
  })

  it('renders the AI features toggle as off when vision_enabled is false', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_2', vision_enabled: false })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.getByText('Disabled — all AI features are off')).toBeInTheDocument()
  })

  it('shows the AI features explanation on hover over the help tooltip', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })

    renderAdvancedSection()

    await waitFor(() => screen.getByRole('switch'))

    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()

    await userEvent.hover(screen.getByRole('button', { name: 'What does this do?' }))
    expect(screen.getByRole('tooltip')).toBeInTheDocument()

    await userEvent.unhover(screen.getByRole('button', { name: 'What does this do?' }))
    expect(screen.queryByRole('tooltip')).not.toBeInTheDocument()
  })

  it('enables vision via PATCH /me and updates the displayed state', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })
    mockUpdateMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByRole('switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { vision_enabled: true })
    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })
    expect(screen.getByText('Active — folder suggestions on upload')).toBeInTheDocument()
  })

  it('disables vision via PATCH /me and updates the displayed state', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })
    mockUpdateMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { vision_enabled: false })
    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.getByText('Disabled — all AI features are off')).toBeInTheDocument()
  })

  it('disables the switch while the update is pending', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })
    let resolveUpdate: (value: { id: string; vision_enabled: boolean }) => void = () => {}
    mockUpdateMe.mockImplementation(
      () => new Promise((resolve) => { resolveUpdate = resolve }),
    )

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByRole('switch'))

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('aria-disabled', 'true')
    })

    resolveUpdate({ id: 'kp_1', vision_enabled: true })

    await waitFor(() => {
      expect(screen.getByRole('switch')).not.toHaveAttribute('aria-disabled')
    })
  })

  it('shows an error toast and leaves the toggle unchanged when the update fails', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })
    mockUpdateMe.mockRejectedValue(new Error('boom'))

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByRole('switch'))

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Failed to update settings')
    })
    expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    expect(screen.getByRole('switch')).not.toHaveAttribute('aria-disabled')
  })
})
