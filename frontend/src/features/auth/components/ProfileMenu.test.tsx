import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import ProfileMenu from './ProfileMenu'

const mockLogout = vi.fn()
const mockGetUserProfile = vi.fn()

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({
    getUserProfile: mockGetUserProfile,
    logout: mockLogout,
  }),
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
    DropdownMenuTrigger: ({ children }: { children: React.ReactNode; asChild?: boolean }) =>
      React.createElement(React.Fragment, null, children),
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

describe('ProfileMenu', () => {
  beforeEach(() => {
    vi.clearAllMocks()
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
})
