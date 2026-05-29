import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect } from 'vitest'
import TagInput from './TagInput'

function makeTags(names: string[]) {
  return names.map((name, i) => ({ id: `id-${i}`, name }))
}

describe('TagInput — success scenario', () => {
  it('calls onChange with the new tag when Enter is pressed', async () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.type(input, 'nature{Enter}')

    expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'nature' }])
  })

  it('calls onChange removing the tag when the remove button is clicked', async () => {
    const tags = makeTags(['nature', 'landscape'])
    const onChange = vi.fn()
    render(<TagInput tags={tags} onChange={onChange} />)

    await userEvent.click(screen.getByRole('button', { name: /remove tag nature/i }))

    expect(onChange).toHaveBeenCalledWith([tags[1]])
  })
})

describe('TagInput — failure scenario', () => {
  it('does not call onChange when a duplicate name is committed', async () => {
    const tags = makeTags(['nature'])
    const onChange = vi.fn()
    render(<TagInput tags={tags} onChange={onChange} />)

    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'nature{Enter}')

    expect(onChange).not.toHaveBeenCalled()
  })
})
