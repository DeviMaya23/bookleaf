import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { describe, it, expect, vi, beforeEach } from 'vitest'
import AdvancedSection from './AdvancedSection'

const mockGetMe = vi.fn()

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getToken: vi.fn(),
  }),
}))

vi.mock('@/features/auth/lib/me', () => ({
  getMe: () => mockGetMe(),
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

  it('renders the AI features toggle as on and disabled when vision_enabled is true', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })
    expect(screen.getByRole('switch')).toHaveAttribute('aria-disabled', 'true')
    expect(screen.getByText('Active — folder suggestions on upload')).toBeInTheDocument()
  })

  it('renders the AI features toggle as off and disabled when vision_enabled is false', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_2', vision_enabled: false })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.getByRole('switch')).toHaveAttribute('aria-disabled', 'true')
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
})
