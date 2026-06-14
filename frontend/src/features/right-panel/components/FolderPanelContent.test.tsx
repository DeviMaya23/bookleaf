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

import { updateFolder, getFolder, exportFolder } from '@/lib/folders'

const folder = { id: 'folder-1', name: 'Nature', description: 'Outdoor shots' }

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
