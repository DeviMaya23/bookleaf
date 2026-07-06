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
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-checked')
    })
    expect(screen.getByTestId('vision-description')).not.toBeEmptyDOMElement()
  })

  it('renders the AI features toggle as off when vision_enabled is false', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_2', vision_enabled: false })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.getByTestId('vision-description')).not.toBeEmptyDOMElement()
  })

  it('shows different description text depending on vision_enabled state', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })
    renderAdvancedSection()
    await waitFor(() => {
      expect(screen.getAllByTestId('vision-switch')[0]).toHaveAttribute('data-checked')
    })
    const enabledText = screen.getAllByTestId('vision-description')[0].textContent

    mockGetMe.mockResolvedValue({ id: 'kp_2', vision_enabled: false })
    renderAdvancedSection()
    await waitFor(() => {
      expect(screen.getAllByTestId('vision-switch')[1]).toHaveAttribute('data-unchecked')
    })
    const disabledText = screen.getAllByTestId('vision-description')[1].textContent

    expect(disabledText).not.toEqual(enabledText)
  })

  it('shows the AI features explanation on hover over the help tooltip', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })

    renderAdvancedSection()

    await waitFor(() => screen.getByTestId('vision-switch'))

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
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-unchecked')
    })
    const initialDescription = screen.getByTestId('vision-description').textContent

    await userEvent.click(screen.getByTestId('vision-switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { vision_enabled: true })
    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-checked')
    })
    expect(screen.getByTestId('vision-description').textContent).not.toEqual(initialDescription)
  })

  it('disables vision via PATCH /me and updates the displayed state', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: true })
    mockUpdateMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-checked')
    })
    const initialDescription = screen.getByTestId('vision-description').textContent

    await userEvent.click(screen.getByTestId('vision-switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { vision_enabled: false })
    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.getByTestId('vision-description').textContent).not.toEqual(initialDescription)
  })

  it('disables the switch while the update is pending', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })
    let resolveUpdate: (value: { id: string; vision_enabled: boolean }) => void = () => {}
    mockUpdateMe.mockImplementation(
      () => new Promise((resolve) => { resolveUpdate = resolve }),
    )

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByTestId('vision-switch'))

    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('aria-disabled', 'true')
    })

    resolveUpdate({ id: 'kp_1', vision_enabled: true })

    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).not.toHaveAttribute('aria-disabled')
    })
  })

  it('shows an error toast and leaves the toggle unchanged when the update fails', async () => {
    mockGetMe.mockResolvedValue({ id: 'kp_1', vision_enabled: false })
    mockUpdateMe.mockRejectedValue(new Error('boom'))

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByTestId('vision-switch'))

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Failed to update settings')
    })
    expect(screen.getByTestId('vision-switch')).toHaveAttribute('data-unchecked')
    expect(screen.getByTestId('vision-switch')).not.toHaveAttribute('aria-disabled')
  })

  it('renders the AI auto-categorisation toggle reflecting ai_categorisation_enabled from the me cache', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: false,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 10,
    })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-checked')
    })
  })

  it('toggling the AI auto-categorisation switch fires PATCH /me with the new value', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: true,
      ai_categorisation_enabled: false,
      ai_categorisation_count_this_month: 0,
    })
    mockUpdateMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: true,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 0,
    })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByTestId('ai-categorisation-switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { ai_categorisation_enabled: true })
    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-checked')
    })
  })

  it('disables the AI auto-categorisation switch when vision_enabled is false', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: false,
      ai_categorisation_enabled: false,
      ai_categorisation_count_this_month: 0,
    })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('aria-disabled', 'true')
    })
    await userEvent.click(screen.getByTestId('ai-categorisation-switch'))
    expect(mockUpdateMe).not.toHaveBeenCalled()
  })

  it('disables AI auto-categorisation via PATCH /me and updates the displayed state', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: true,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 0,
    })
    mockUpdateMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: true,
      ai_categorisation_enabled: false,
      ai_categorisation_count_this_month: 0,
    })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByTestId('ai-categorisation-switch'))

    expect(mockUpdateMe).toHaveBeenCalledWith(expect.anything(), { ai_categorisation_enabled: false })
    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-unchecked')
    })
  })

  it('shows an error toast and leaves the categorisation toggle unchanged when the update fails', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: true,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 0,
    })
    mockUpdateMe.mockRejectedValue(new Error('boom'))

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByTestId('ai-categorisation-switch'))

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Failed to update settings')
    })
    expect(screen.getByTestId('ai-categorisation-switch')).toHaveAttribute('data-checked')
    expect(screen.getByTestId('ai-categorisation-switch')).not.toHaveAttribute('aria-disabled')
  })

  it('renders the categorisation counter showing current month usage', async () => {
    mockGetMe.mockResolvedValue({
      id: 'kp_1',
      vision_enabled: true,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 10,
    })

    renderAdvancedSection()

    await waitFor(() => {
      expect(screen.getByTestId('categorisation-counter').textContent).toBe('10 / 50 this month')
    })
  })
})
