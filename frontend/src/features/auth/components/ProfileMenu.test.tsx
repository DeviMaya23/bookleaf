import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ProfileMenu from './ProfileMenu'

const mockLogout = vi.fn()
const mockGetUserProfile = vi.fn()
const mockGetMe = vi.fn()

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getUserProfile: mockGetUserProfile,
    getToken: vi.fn(),
    logout: mockLogout,
  }),
}))

vi.mock('@/features/auth/lib/me', () => ({
  getMe: () => mockGetMe(),
}))

vi.mock('@/components/ui/avatar', async () => {
  const React = await import('react')
  return {
    Avatar: ({ children, className }: { children: React.ReactNode; className?: string }) =>
      React.createElement('div', { 'data-testid': 'avatar', className }, children),
    AvatarImage: ({ src }: { src?: string }) =>
      src ? React.createElement('img', { 'data-testid': 'avatar-image', src }) : null,
    AvatarFallback: ({ children }: { children: React.ReactNode }) =>
      React.createElement('span', { 'data-testid': 'avatar-fallback' }, children),
  }
})

vi.mock('@/components/ui/dropdown-menu', async () => {
  const React = await import('react')
  return {
    DropdownMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DropdownMenuTrigger: ({
      children,
      onClick,
    }: {
      children: React.ReactNode
      asChild?: boolean
      onClick?: () => void
      className?: string
    }) => React.createElement('button', { onClick, 'data-testid': 'dropdown-trigger' }, children),
    DropdownMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { 'data-testid': 'dropdown-content' }, children),
    DropdownMenuItem: ({
      children,
      onClick,
      className,
    }: {
      children: React.ReactNode
      onClick?: () => void
      className?: string
    }) => React.createElement('button', { onClick, className }, children),
    DropdownMenuSeparator: () => React.createElement('hr'),
  }
})

vi.mock('@/features/settings/components/SettingsModal', async () => {
  const React = await import('react')
  return {
    default: ({ open }: { open: boolean; onOpenChange: (open: boolean) => void }) =>
      open ? React.createElement('div', { 'data-testid': 'settings-modal' }, 'Settings Modal') : null,
  }
})

function renderProfileMenu() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <ProfileMenu />
    </QueryClientProvider>,
  )
}

const defaultProfile = {
  id: 'kp_1',
  givenName: 'Jane',
  familyName: 'Doe',
  email: 'jane@example.com',
  picture: null,
}

const defaultMe = {
  id: 'kp_1',
  vision_enabled: false,
  folder_icons_enabled: false,
  ai_categorisation_enabled: false,
  ai_categorisation_count_this_month: 0,
}

describe('ProfileMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    localStorage.clear()
    mockGetMe.mockResolvedValue(defaultMe)
  })

  it('renders avatar image when picture is available', async () => {
    mockGetUserProfile.mockResolvedValue({
      id: 'kp_1',
      givenName: 'Jane',
      familyName: 'Doe',
      email: 'jane@example.com',
      picture: 'https://example.com/avatar.jpg',
    })

    renderProfileMenu()

    await waitFor(() => {
      expect(screen.getByTestId('avatar-image')).toHaveAttribute('src', 'https://example.com/avatar.jpg')
    })
    expect(screen.getByText('Jane Doe')).toBeInTheDocument()
  })

  it('renders initials fallback when picture is absent', async () => {
    mockGetUserProfile.mockResolvedValue({
      id: 'kp_2',
      givenName: 'Jane',
      familyName: 'Doe',
      email: 'jane@example.com',
      picture: null,
    })

    renderProfileMenu()

    await waitFor(() => {
      expect(screen.getByText('JD')).toBeInTheDocument()
    })
    expect(screen.queryByTestId('avatar-image')).not.toBeInTheDocument()
  })

  it('calls logout when Sign out is clicked', async () => {
    mockGetUserProfile.mockResolvedValue({
      id: 'kp_3',
      givenName: 'Jane',
      familyName: 'Doe',
      email: 'jane@example.com',
      picture: null,
    })

    renderProfileMenu()

    await waitFor(() => screen.getByText('JD'))

    await userEvent.click(screen.getByText('Sign out'))
    expect(mockLogout).toHaveBeenCalledOnce()
  })

  it('hides the Settings item below sm but keeps Sign out visible', async () => {
    mockGetUserProfile.mockResolvedValue({
      id: 'kp_5',
      givenName: 'Jane',
      familyName: 'Doe',
      email: 'jane@example.com',
      picture: null,
    })

    renderProfileMenu()

    await waitFor(() => screen.getByText('JD'))

    expect(screen.getByText('Settings').className).toMatch(/hidden sm:flex/)
    expect(screen.getByText('Sign out').className).not.toMatch(/hidden sm:flex/)
  })

  it('opens the SettingsModal when Settings is clicked', async () => {
    mockGetUserProfile.mockResolvedValue({
      id: 'kp_4',
      givenName: 'Jane',
      familyName: 'Doe',
      email: 'jane@example.com',
      picture: null,
    })

    renderProfileMenu()

    await waitFor(() => screen.getByText('JD'))

    expect(screen.queryByTestId('settings-modal')).not.toBeInTheDocument()

    await userEvent.click(screen.getByText('Settings'))

    expect(screen.getByTestId('settings-modal')).toBeInTheDocument()
  })

  it('renders the badge when limit is hit and localStorage key is not set', async () => {
    mockGetUserProfile.mockResolvedValue(defaultProfile)
    mockGetMe.mockResolvedValue({
      ...defaultMe,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 50,
    })

    renderProfileMenu()

    await waitFor(() => {
      expect(screen.getByTestId('categorisation-limit-badge')).toBeInTheDocument()
    })
  })

  it('does not render the badge when count is below 50', async () => {
    mockGetUserProfile.mockResolvedValue(defaultProfile)
    mockGetMe.mockResolvedValue({
      ...defaultMe,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 49,
    })

    renderProfileMenu()

    await waitFor(() => screen.getByTestId('avatar'))
    expect(screen.queryByTestId('categorisation-limit-badge')).not.toBeInTheDocument()
  })

  it('does not render the badge when ai_categorisation_enabled is false', async () => {
    mockGetUserProfile.mockResolvedValue(defaultProfile)
    mockGetMe.mockResolvedValue({
      ...defaultMe,
      ai_categorisation_enabled: false,
      ai_categorisation_count_this_month: 50,
    })

    renderProfileMenu()

    await waitFor(() => screen.getByTestId('avatar'))
    expect(screen.queryByTestId('categorisation-limit-badge')).not.toBeInTheDocument()
  })

  it('clicking the trigger sets the localStorage key and hides the badge', async () => {
    mockGetUserProfile.mockResolvedValue(defaultProfile)
    mockGetMe.mockResolvedValue({
      ...defaultMe,
      ai_categorisation_enabled: true,
      ai_categorisation_count_this_month: 50,
    })

    renderProfileMenu()

    await waitFor(() => {
      expect(screen.getByTestId('categorisation-limit-badge')).toBeInTheDocument()
    })

    await userEvent.click(screen.getByTestId('dropdown-trigger'))

    await waitFor(() => {
      expect(screen.queryByTestId('categorisation-limit-badge')).not.toBeInTheDocument()
    })

    const now = new Date()
    const key = `categorisation_limit_dismissed_${now.getUTCFullYear()}-${String(now.getUTCMonth() + 1).padStart(2, '0')}`
    expect(localStorage.getItem(key)).toBe('1')
  })
})
