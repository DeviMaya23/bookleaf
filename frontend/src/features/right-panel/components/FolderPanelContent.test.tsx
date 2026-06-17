import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import FolderPanelContent from './FolderPanelContent'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/folders', () => ({
  updateFolder: vi.fn(),
  getFolder: vi.fn(),
  exportFolder: vi.fn(),
}))

vi.mock('@/lib/share', () => ({
  getFolderShare: vi.fn(),
  createFolderShare: vi.fn(),
  deleteFolderShare: vi.fn(),
}))

vi.mock('@/components/ui/dialog', async () => {
  const React = await import('react')
  return {
    Dialog: ({ open, children }: { open: boolean; onOpenChange?: (v: boolean) => void; children: React.ReactNode }) =>
      open ? React.createElement(React.Fragment, null, children) : null,
    DialogContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', { role: 'dialog' }, children),
    DialogHeader: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DialogTitle: ({ children }: { children: React.ReactNode }) =>
      React.createElement('h2', null, children),
    DialogFooter: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
  }
})

const mockToastSuccess = vi.fn()
const mockToastError = vi.fn()

vi.mock('sonner', () => ({
  toast: {
    success: (...args: unknown[]) => mockToastSuccess(...args),
    error: (...args: unknown[]) => mockToastError(...args),
  },
}))

import { updateFolder, getFolder, exportFolder } from '@/lib/folders'
import { getFolderShare, createFolderShare, deleteFolderShare } from '@/lib/share'

const folder = { id: 'folder-1', name: 'Nature', description: 'Outdoor shots', icon: null }

beforeEach(() => {
  vi.mocked(getFolderShare).mockResolvedValue(null)
})

function renderPanel() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <FolderPanelContent folder={folder} onClose={vi.fn()} />
    </QueryClientProvider>,
  )
}

describe('FolderPanelContent — success scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '' })
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 1 })
    vi.mocked(exportFolder).mockResolvedValue(new Blob())
  })

  it('saves the name on blur when changed', async () => {
    renderPanel()

    const nameInput = screen.getByDisplayValue('Nature')
    await userEvent.clear(nameInput)
    await userEvent.type(nameInput, 'Wildlife')
    await userEvent.tab()

    await waitFor(() => {
      expect(updateFolder).toHaveBeenCalledWith(expect.any(Function), 'folder-1', { name: 'Wildlife' })
    })
  })

  it('saves the description on blur when changed', async () => {
    renderPanel()

    const descriptionInput = screen.getByDisplayValue('Outdoor shots')
    await userEvent.clear(descriptionInput)
    await userEvent.type(descriptionInput, 'Updated notes')
    await userEvent.tab()

    await waitFor(() => {
      expect(updateFolder).toHaveBeenCalledWith(expect.any(Function), 'folder-1', { description: 'Updated notes' })
    })
  })
})

describe('FolderPanelContent — empty name scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('reverts to the original name and does not save when blurred empty', async () => {
    renderPanel()

    const nameInput = screen.getByDisplayValue('Nature')
    await userEvent.clear(nameInput)
    await userEvent.tab()

    await waitFor(() => {
      expect(screen.getByDisplayValue('Nature')).toBeInTheDocument()
    })
    expect(updateFolder).not.toHaveBeenCalled()
  })
})

describe('FolderPanelContent — image count subtitle', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '' })
    vi.mocked(exportFolder).mockResolvedValue(new Blob())
  })

  it('shows the plural label for multiple images', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 12 })
    renderPanel()

    expect(await screen.findByText('12 images')).toBeInTheDocument()
  })

  it('shows the singular label for one image', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 1 })
    renderPanel()

    expect(await screen.findByText('1 image')).toBeInTheDocument()
  })

  it('shows the plural label for zero images', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 0 })
    renderPanel()

    expect(await screen.findByText('0 images')).toBeInTheDocument()
  })

  it('does not show a count while the folder detail is loading', async () => {
    vi.mocked(getFolder).mockReturnValue(new Promise(() => {}))
    renderPanel()

    expect(screen.queryByText(/image/i)).not.toBeInTheDocument()
  })
})

describe('FolderPanelContent — export folder button', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(updateFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '' })
  })

  it('disables the button while the folder has zero images', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 0 })
    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /export folder/i })).toBeDisabled()
    })
  })

  it('enables the button when the folder has images', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 2 })
    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /export folder/i })).not.toBeDisabled()
    })
  })

  it('shows a preparing state and triggers a download on success', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 2 })
    let resolveExport: (blob: Blob) => void = () => {}
    vi.mocked(exportFolder).mockReturnValue(
      new Promise((resolve) => {
        resolveExport = resolve
      }),
    )
    const createObjectURLSpy = vi.spyOn(URL, 'createObjectURL').mockReturnValue('blob:mock-url')
    const revokeObjectURLSpy = vi.spyOn(URL, 'revokeObjectURL').mockImplementation(() => {})

    renderPanel()

    const button = await screen.findByRole('button', { name: /export folder/i })
    await waitFor(() => expect(button).not.toBeDisabled())
    await userEvent.click(button)

    expect(screen.getByRole('button', { name: /preparing export/i })).toBeDisabled()

    resolveExport(new Blob())

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /export folder/i })).not.toBeDisabled()
    })
    expect(exportFolder).toHaveBeenCalledWith(expect.any(Function), 'folder-1')
    expect(createObjectURLSpy).toHaveBeenCalled()
    expect(revokeObjectURLSpy).toHaveBeenCalledWith('blob:mock-url')
  })

  it('shows an error toast and resets the button when export fails', async () => {
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 2 })
    vi.mocked(exportFolder).mockRejectedValue(new Error('Network error'))

    renderPanel()

    const button = await screen.findByRole('button', { name: /export folder/i })
    await waitFor(() => expect(button).not.toBeDisabled())
    await userEvent.click(button)

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /export folder/i })).not.toBeDisabled()
    })
    expect(exportFolder).toHaveBeenCalledWith(expect.any(Function), 'folder-1')
  })
})

describe('FolderPanelContent — share folder section', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(getFolder).mockResolvedValue({ ...folder, parent_id: null, created_at: '', updated_at: '', image_count: 1 })
    vi.mocked(exportFolder).mockResolvedValue(new Blob())
  })

  it('shows the switch off and no link field when sharing is off', async () => {
    vi.mocked(getFolderShare).mockResolvedValue(null)

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.queryByDisplayValue(/\/share\//)).not.toBeInTheDocument()
  })

  it('shows the switch on and the link field when sharing is on', async () => {
    vi.mocked(getFolderShare).mockResolvedValue({ token: 'abc123' })

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })
    expect(screen.getByDisplayValue(`${window.location.origin}/share/abc123`)).toBeInTheDocument()
  })

  it('enables sharing and shows the link field when the switch is turned on', async () => {
    vi.mocked(getFolderShare).mockResolvedValueOnce(null).mockResolvedValue({ token: 'new-token' })
    vi.mocked(createFolderShare).mockResolvedValue({ token: 'new-token' })

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })

    await userEvent.click(screen.getByRole('switch'))

    expect(createFolderShare).toHaveBeenCalledWith(expect.any(Function), 'folder-1')
    await waitFor(() => {
      expect(screen.getByDisplayValue(`${window.location.origin}/share/new-token`)).toBeInTheDocument()
    })
  })

  it('opens a confirm dialog when turning sharing off, and disables sharing on confirm', async () => {
    vi.mocked(getFolderShare).mockResolvedValue({ token: 'abc123' })
    vi.mocked(deleteFolderShare).mockResolvedValue(undefined)

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('switch'))

    expect(screen.getByRole('heading', { name: /disable sharing/i })).toBeInTheDocument()
    expect(deleteFolderShare).not.toHaveBeenCalled()

    vi.mocked(getFolderShare).mockResolvedValue(null)

    await userEvent.click(screen.getByRole('button', { name: 'Disable' }))

    expect(deleteFolderShare).toHaveBeenCalledWith(expect.any(Function), 'folder-1')
    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-unchecked')
    })
    expect(screen.queryByDisplayValue(/\/share\//)).not.toBeInTheDocument()
  })

  it('cancelling the confirm dialog leaves sharing on and makes no request', async () => {
    vi.mocked(getFolderShare).mockResolvedValue({ token: 'abc123' })

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('switch'))
    expect(screen.getByRole('heading', { name: /disable sharing/i })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: 'Cancel' }))

    expect(deleteFolderShare).not.toHaveBeenCalled()
    expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    expect(screen.getByDisplayValue(`${window.location.origin}/share/abc123`)).toBeInTheDocument()
  })

  it('copies the share link to the clipboard and shows a success toast', async () => {
    vi.mocked(getFolderShare).mockResolvedValue({ token: 'abc123' })
    const writeText = vi.fn().mockResolvedValue(undefined)
    Object.assign(navigator, { clipboard: { writeText } })

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('button', { name: /copy share link/i }))

    expect(writeText).toHaveBeenCalledWith(`${window.location.origin}/share/abc123`)
    await waitFor(() => {
      expect(mockToastSuccess).toHaveBeenCalledWith('Link copied')
    })
  })

  it('shows an error toast when copying the share link fails', async () => {
    vi.mocked(getFolderShare).mockResolvedValue({ token: 'abc123' })
    const writeText = vi.fn().mockRejectedValue(new Error('Clipboard unavailable'))
    Object.assign(navigator, { clipboard: { writeText } })

    renderPanel()

    await waitFor(() => {
      expect(screen.getByRole('switch')).toHaveAttribute('data-checked')
    })

    await userEvent.click(screen.getByRole('button', { name: /copy share link/i }))

    await waitFor(() => {
      expect(mockToastError).toHaveBeenCalledWith('Failed to copy link')
    })
  })
})
