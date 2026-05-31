import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect } from 'vitest'
import FolderInput from './FolderInput'

const nature = { id: 'folder-nature', name: 'Nature', description: null, parent_id: null, created_at: '', updated_at: '' }
const travel = { id: 'folder-travel', name: 'Travel', description: null, parent_id: null, created_at: '', updated_at: '' }

describe('FolderInput — success scenario', () => {
  it('selecting a suggestion calls onChange with folder appended', async () => {
    const onChange = vi.fn()
    render(
      <FolderInput
        folders={[]}
        onChange={onChange}
        suggestions={[nature, travel]}
      />,
    )

    const input = screen.getByPlaceholderText('Add to folder…')
    await userEvent.type(input, 'nat')

    const option = await screen.findByText('Nature')
    await userEvent.click(option)

    expect(onChange).toHaveBeenCalledWith([nature])
  })
})

describe('FolderInput — failure scenario', () => {
  it('blurring without selecting a suggestion does not call onChange', async () => {
    const onChange = vi.fn()
    render(
      <FolderInput
        folders={[]}
        onChange={onChange}
        suggestions={[nature, travel]}
      />,
    )

    const input = screen.getByPlaceholderText('Add to folder…')
    await userEvent.type(input, 'nat')
    await userEvent.tab()

    await waitFor(() => {
      expect(onChange).not.toHaveBeenCalled()
    })
  })
})
