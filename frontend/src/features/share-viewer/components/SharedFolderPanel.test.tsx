import { render, screen } from '@testing-library/react'
import { MemoryRouter } from 'react-router-dom'
import { vi, describe, it, expect } from 'vitest'
import SharedFolderPanel from './SharedFolderPanel'

vi.mock('@/lib/share', () => ({
  exportSharedFolder: vi.fn(),
}))

function renderPanel(folder: { name: string; notes: string | null }, imageCount: number) {
  return render(
    <MemoryRouter>
      <SharedFolderPanel token="tok_abc123" folder={folder} imageCount={imageCount} />
    </MemoryRouter>,
  )
}

describe('SharedFolderPanel', () => {
  it('renders folder name, image count, and notes', () => {
    renderPanel({ name: 'Travel', notes: 'A trip board' }, 3)

    expect(screen.getByText('Travel')).toBeInTheDocument()
    expect(screen.getByText('3 images')).toBeInTheDocument()
    expect(screen.getByText('A trip board')).toBeInTheDocument()
  })

  it('shows "No notes added" when notes is null', () => {
    renderPanel({ name: 'Travel', notes: null }, 1)

    expect(screen.getByText('1 image')).toBeInTheDocument()
    expect(screen.getByText('No notes added')).toBeInTheDocument()
  })

  it('disables the export button when the folder has zero images', () => {
    renderPanel({ name: 'Travel', notes: null }, 0)

    expect(screen.getByRole('button', { name: /export folder/i })).toBeDisabled()
  })

  it('enables the export button when the folder has images', () => {
    renderPanel({ name: 'Travel', notes: null }, 2)

    expect(screen.getByRole('button', { name: /export folder/i })).toBeEnabled()
  })
})
