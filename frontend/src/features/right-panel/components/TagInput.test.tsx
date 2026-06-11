import { render, screen, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect } from 'vitest'
import TagInput from './TagInput'

function makeTags(names: string[]) {
  return names.map((name, i) => ({ id: `id-${i}`, name }))
}

describe('TagInput', () => {
  it('shows the "Add tags…" placeholder when empty', () => {
    render(<TagInput tags={[]} onChange={vi.fn()} />)

    expect(screen.getByPlaceholderText('Add tags…')).toBeInTheDocument()
  })

  it('renders tag items as chips', () => {
    render(<TagInput tags={makeTags(['nature'])} onChange={vi.fn()} />)

    expect(screen.getByText('nature')).toBeInTheDocument()
  })
})

describe('TagInput — free-text creation', () => {
  it('Enter on raw text creates a lowercase tag with id ""', async () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.type(input, 'NATURE{Enter}')

    expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'nature' }])
  })

  it('strips embedded commas from pasted text', async () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.click(input)
    fireEvent.change(input, { target: { value: 'na,ture' } })
    await userEvent.keyboard('{Enter}')

    expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'nature' }])
  })

  it('does not call onChange when the new name duplicates an existing tag', async () => {
    const onChange = vi.fn()
    render(<TagInput tags={makeTags(['nature'])} onChange={onChange} />)

    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'nature{Enter}')

    expect(onChange).not.toHaveBeenCalled()
  })
})
