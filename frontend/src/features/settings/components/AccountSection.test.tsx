import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import AccountSection from './AccountSection'

const mockLogout = vi.fn()
const mockGetUserProfile = vi.fn()
const mockDeleteMe = vi.fn()
const mockToastError = vi.fn()

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getUserProfile: mockGetUserProfile,
    logout: mockLogout,
    getToken: vi.fn(),
  }),
}))

vi.mock('@/features/auth/lib/me', () => ({
  deleteMe: () => mockDeleteMe(),
}))

vi.mock('sonner', () => ({
  toast: {
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

vi.mock('@/components/ui/avatar', async () => {
  const React = await import('react')
  return {
    Avatar: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { 'data-testid': 'avatar' }, children),
    AvatarImage: ({ src }: { src?: string }) =>
      src ? React.createElement('img', { 'data-testid': 'avatar-image', src }) : null,
    AvatarFallback: ({ children }: { children: React.ReactNode }) =>
      React.createElement('span', { 'data-testid': 'avatar-fallback' }, children),
  }
})

const PROFILE = {
  id: 'kp_1',
  givenName: 'Jane',
  familyName: 'Doe',
  email: 'jane@example.com',
  picture: null as string | null,
}

function renderAccountSection() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <AccountSection />
    </QueryClientProvider>,
  )
}

describe('AccountSection', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders profile fields read-only', async () => {
    mockGetUserProfile.mockResolvedValue({ ...PROFILE, picture: 'https://example.com/avatar.jpg' })

    renderAccountSection()

    await waitFor(() => {
      expect(screen.getByText('Jane Doe')).toBeInTheDocument()
    })
    expect(screen.getByText('jane@example.com')).toBeInTheDocument()
    expect(screen.getByTestId('avatar-image')).toHaveAttribute('src', 'https://example.com/avatar.jpg')
    expect(screen.queryByRole('textbox')).not.toBeInTheDocument()
  })

  it('shows initials fallback when no picture is available', async () => {
    mockGetUserProfile.mockResolvedValue({ ...PROFILE, picture: null })

    renderAccountSection()

    await waitFor(() => {
      expect(screen.getByText('JD')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('avatar-image')).not.toBeInTheDocument()
  })

  it('calls logout when Log out is clicked', async () => {
    mockGetUserProfile.mockResolvedValue(PROFILE)

    renderAccountSection()

    await waitFor(() => screen.getByText('Jane Doe'))

    await userEvent.click(screen.getByRole('button', { name: 'Log out' }))

    expect(mockLogout).toHaveBeenCalledOnce()
  })

  describe('delete account', () => {
    async function openConfirmView() {
      mockGetUserProfile.mockResolvedValue(PROFILE)
      renderAccountSection()
      await waitFor(() => screen.getByText('Jane Doe'))
      await userEvent.click(screen.getByRole('button', { name: 'Delete account…' }))
    }

    it('keeps the confirm action disabled until the typed email matches', async () => {
      await openConfirmView()

      const confirmButton = screen.getByRole('button', { name: 'Delete account' })
      expect(confirmButton).toBeDisabled()

      await userEvent.type(screen.getByRole('textbox'), 'wrong@example.com')
      expect(confirmButton).toBeDisabled()

      await userEvent.clear(screen.getByRole('textbox'))
      await userEvent.type(screen.getByRole('textbox'), '  JANE@EXAMPLE.COM  ')
      expect(confirmButton).toBeEnabled()
    })

    it('logs the user out on successful deletion', async () => {
      mockDeleteMe.mockResolvedValue(undefined)
      await openConfirmView()

      await userEvent.type(screen.getByRole('textbox'), 'jane@example.com')
      await userEvent.click(screen.getByRole('button', { name: 'Delete account' }))

      await waitFor(() => {
        expect(mockLogout).toHaveBeenCalledOnce()
      })
    })

    it('shows a toast and preserves the confirmation view on failure', async () => {
      mockDeleteMe.mockRejectedValue(new Error('boom'))
      await openConfirmView()

      await userEvent.type(screen.getByRole('textbox'), 'jane@example.com')
      await userEvent.click(screen.getByRole('button', { name: 'Delete account' }))

      await waitFor(() => {
        expect(mockToastError).toHaveBeenCalled()
      })
      expect(mockLogout).not.toHaveBeenCalled()
      expect(screen.getByRole('textbox')).toHaveValue('jane@example.com')
      expect(screen.getByRole('button', { name: 'Delete account' })).toBeInTheDocument()
    })
  })
})
