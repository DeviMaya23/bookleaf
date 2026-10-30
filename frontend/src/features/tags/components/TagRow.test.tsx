import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import TagRow from './TagRow'
import { useRenameMutation, useDeleteMutation } from '../hooks/useTagMutations'

vi.mock('@/components/ui/dropdown-menu', async () => {
  const React = await import('react')
  return {
    DropdownMenu: ({ children }: { children: React.ReactNode }) =>
      React.createElement(React.Fragment, null, children),
    DropdownMenuTrigger: ({ children, ...props }: React.ButtonHTMLAttributes<HTMLButtonElement> & { children: React.ReactNode }) =>
      React.createElement('button', props, children),
    DropdownMenuContent: ({ children }: { children: React.ReactNode }) =>
      React.createElement('div', null, children),
    DropdownMenuItem: ({ children, onClick, className }: { children: React.ReactNode; onClick?: () => void; className?: string }) =>
      React.createElement('div', { role: 'menuitem', onClick, className }, children),
  }
})

vi.mock('../hooks/useTagMutations')

const mockRenameMutate = vi.fn()
const mockDeleteMutate = vi.fn()
const mockClearDuplicateError = vi.fn()

function setupMocks({ duplicateError = false } = {}) {
  vi.mocked(useRenameMutation).mockReturnValue({
    mutate: mockRenameMutate,
    isPending: false,
    duplicateError,
    clearDuplicateError: mockClearDuplicateError,
  })
  vi.mocked(useDeleteMutation).mockReturnValue({
    mutate: mockDeleteMutate,
    isPending: false,
  })
}

const tag = { id: 'tag-1', name: 'Nature', image_count: 5 }

beforeEach(() => {
  mockRenameMutate.mockClear()
  mockDeleteMutate.mockClear()
  mockClearDuplicateError.mockClear()
})

describe('TagRow — rename mode', () => {
  it('resets the input to the tag name when entering rename mode', async () => {
    setupMocks()
    render(<TagRow tag={tag} onClick={vi.fn()} />)

    await userEvent.click(screen.getByRole('menuitem', { name: /rename/i }))

    expect(screen.getByRole('textbox')).toHaveValue('Nature')
  })

  it('does not call mutate when submitting with the same name', async () => {
    setupMocks()
    render(<TagRow tag={tag} onClick={vi.fn()} />)

    await userEvent.click(screen.getByRole('menuitem', { name: /rename/i }))
    await userEvent.click(screen.getByRole('button', { name: /save/i }))

    expect(mockRenameMutate).not.toHaveBeenCalled()
  })

  it('does not call mutate when submitting an empty name', async () => {
    setupMocks()
    render(<TagRow tag={tag} onClick={vi.fn()} />)

    await userEvent.click(screen.getByRole('menuitem', { name: /rename/i }))
    await userEvent.clear(screen.getByRole('textbox'))
    await userEvent.click(screen.getByRole('button', { name: /save/i }))

    expect(mockRenameMutate).not.toHaveBeenCalled()
  })

  it('calls mutate with the tag id and new name when submitting a changed name', async () => {
    setupMocks()
    render(<TagRow tag={tag} onClick={vi.fn()} />)

    await userEvent.click(screen.getByRole('menuitem', { name: /rename/i }))
    const input = screen.getByRole('textbox')
    await userEvent.clear(input)
    await userEvent.type(input, 'Wildlife')
    await userEvent.click(screen.getByRole('button', { name: /save/i }))

    expect(mockRenameMutate).toHaveBeenCalledWith('tag-1', 'Wildlife')
  })

  it('shows the duplicate error message and clears it on input change', async () => {
    setupMocks({ duplicateError: true })
    render(<TagRow tag={tag} onClick={vi.fn()} />)

    await userEvent.click(screen.getByRole('menuitem', { name: /rename/i }))

    expect(screen.getByText(/that tag already exists/i)).toBeInTheDocument()

    await userEvent.type(screen.getByRole('textbox'), 'x')

    expect(mockClearDuplicateError).toHaveBeenCalled()
  })
})

describe('TagRow — delete confirm mode', () => {
  it('calls del.mutate with the tag id when the Delete button is confirmed', async () => {
    setupMocks()
    render(<TagRow tag={tag} onClick={vi.fn()} />)

    await userEvent.click(screen.getByRole('menuitem', { name: /delete/i }))
    await userEvent.click(screen.getByRole('button', { name: /^delete$/i }))

    expect(mockDeleteMutate).toHaveBeenCalledWith('tag-1')
  })
})
