import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import { vi, describe, it, expect } from 'vitest'
import SharePage from './SharePage'
import type { SharedFolderResponse } from '@/lib/share'

vi.mock('@/lib/share', () => ({
  getSharedFolder: vi.fn(),
  exportSharedFolder: vi.fn(),
}))

import { getSharedFolder } from '@/lib/share'

function renderSharePage(token = 'tok_abc123') {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  })
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={[`/share/${token}`]}>
        <Routes>
          <Route path="/share/:token" element={<SharePage />} />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  )
}

function makeFolderResponse(overrides?: Partial<SharedFolderResponse>): SharedFolderResponse {
  return {
    folder: { name: 'Travel', notes: 'A trip board' },
    images: [
      { title: 'First', thumbnail_url: 'https://example.com/thumb1.jpg', full_res_url: 'https://example.com/full1.jpg', download_url: 'https://example.com/full1.jpg?download', width: 200, height: 100 },
      { title: 'Second', thumbnail_url: 'https://example.com/thumb2.jpg', full_res_url: 'https://example.com/full2.jpg', download_url: 'https://example.com/full2.jpg?download', width: 100, height: 200 },
    ],
    ...overrides,
  }
}

describe('SharePage', () => {
  it('shows a loading spinner only while the request is in flight', () => {
    vi.mocked(getSharedFolder).mockReturnValue(new Promise(() => {}))

    const { container } = renderSharePage()

    expect(container.querySelector('.animate-spin')).toBeInTheDocument()
    expect(screen.queryByText('Bookleaf')).not.toBeInTheDocument()
  })

  it('renders the invalid-link error state on a failed request', async () => {
    vi.mocked(getSharedFolder).mockRejectedValue(new Error('Failed to fetch shared folder'))

    renderSharePage()

    expect(await screen.findByText('This link is invalid or has expired')).toBeInTheDocument()
    expect(screen.getByText('Shared via Bookleaf')).toBeInTheDocument()
    expect(screen.queryByRole('button', { name: /export folder/i })).not.toBeInTheDocument()
  })

  it('renders the topbar, grid, and panel for a folder with images', async () => {
    vi.mocked(getSharedFolder).mockResolvedValue(makeFolderResponse())

    renderSharePage()

    expect(await screen.findByText('Bookleaf')).toBeInTheDocument()
    expect(screen.getAllByText('Travel').length).toBeGreaterThan(0)
    expect(screen.getByText('First')).toBeInTheDocument()
    expect(screen.getByText('Second')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /export folder/i })).toBeEnabled()
  })

  it('renders the empty-state message and disabled export for a folder with zero images', async () => {
    vi.mocked(getSharedFolder).mockResolvedValue(makeFolderResponse({ images: [] }))

    renderSharePage()

    expect(await screen.findByText('No images here yet')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /export folder/i })).toBeDisabled()
  })

  it('renders a download link per image with the download URL', async () => {
    vi.mocked(getSharedFolder).mockResolvedValue(makeFolderResponse())

    renderSharePage()
    await screen.findByText('First')

    const downloadLink = screen.getByLabelText('Download First')
    expect(downloadLink).toHaveAttribute('href', 'https://example.com/full1.jpg?download')
    expect(downloadLink).toHaveAttribute('download')
  })

  it('opens the lightbox on double-click', async () => {
    vi.mocked(getSharedFolder).mockResolvedValue(makeFolderResponse())

    renderSharePage()
    const card = await screen.findByText('First')

    await userEvent.dblClick(card)

    expect(await screen.findByText('1 / 2')).toBeInTheDocument()
  })

  it('toggles a selection outline on the card when clicked', async () => {
    vi.mocked(getSharedFolder).mockResolvedValue(makeFolderResponse())

    renderSharePage()
    const title = await screen.findByText('First')
    const card = title.closest('.cursor-pointer') as HTMLElement

    await userEvent.click(card)
    expect(card).toHaveClass('ring-1')

    await userEvent.click(card)
    expect(card).not.toHaveClass('ring-1')
  })

  it('does not toggle the selection outline when the download button is clicked', async () => {
    vi.mocked(getSharedFolder).mockResolvedValue(makeFolderResponse())

    renderSharePage()
    const title = await screen.findByText('First')
    const card = title.closest('.cursor-pointer') as HTMLElement
    const downloadLink = screen.getByLabelText('Download First')

    await userEvent.click(downloadLink)

    expect(card).not.toHaveClass('ring-1')
  })
})
