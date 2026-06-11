import { render, screen } from '@testing-library/react'
import { describe, it, expect, vi } from 'vitest'
import FolderInput from './FolderInput'

describe('FolderInput', () => {
  it('shows the "Add to folder…" placeholder when empty', () => {
    render(<FolderInput folders={[]} onChange={vi.fn()} />)

    expect(screen.getByPlaceholderText('Add to folder…')).toBeInTheDocument()
  })

  it('renders folder items as chips', () => {
    const nature = { id: 'folder-nature', name: 'Nature', description: null, parent_id: null, created_at: '', updated_at: '' }
    render(<FolderInput folders={[nature]} onChange={vi.fn()} />)

    expect(screen.getByText('Nature')).toBeInTheDocument()
  })
})
