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
}))

import { updateFolder } from '@/lib/folders'

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
