import { render, screen, waitFor } from '@testing-library/react'
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

  it('calls onChange with the new tag when comma is pressed', async () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.type(input, 'nature,')

    expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'nature' }])
  })

  it('calls onChange removing last tag when Backspace is pressed on empty input', async () => {
    const tags = makeTags(['nature', 'landscape'])
    const onChange = vi.fn()
    render(<TagInput tags={tags} onChange={onChange} />)

    const input = screen.getByRole('textbox')
    await userEvent.click(input)
    await userEvent.keyboard('{Backspace}')

    expect(onChange).toHaveBeenCalledWith([tags[0]])
  })

  it('calls onChange when input loses focus with a non-empty value', async () => {
    const onChange = vi.fn()
    render(<TagInput tags={[]} onChange={onChange} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.click(input)
    await userEvent.type(input, 'nature')
    await userEvent.tab()

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'nature' }])
    })
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

describe('TagInput suggestions — success scenario', () => {
  it('clicking a suggestion calls onChange with the suggestion ID (not a blank placeholder)', async () => {
    const onChange = vi.fn()
    const suggestions = [{ id: 'tag-abc', name: 'nature' }]
    render(<TagInput tags={[]} onChange={onChange} suggestions={suggestions} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.type(input, 'nat')

    const suggestion = await screen.findByText('nature')
    await userEvent.click(suggestion)

    expect(onChange).toHaveBeenCalledWith([{ id: 'tag-abc', name: 'nature' }])
  })

  it('pressing ArrowDown then Enter commits the first suggestion', async () => {
    const onChange = vi.fn()
    const suggestions = [
      { id: 'tag-1', name: 'nature' },
      { id: 'tag-2', name: 'narrative' },
    ]
    render(<TagInput tags={[]} onChange={onChange} suggestions={suggestions} />)

    const input = screen.getByPlaceholderText('Add tags…')
    await userEvent.type(input, 'na')
    await userEvent.keyboard('{ArrowDown}{Enter}')

    expect(onChange).toHaveBeenCalledWith([{ id: 'tag-1', name: 'nature' }])
  })
})

describe('TagInput suggestions — failure scenario', () => {
  it('already-applied tags are not shown in the suggestion dropdown', async () => {
    const appliedTags = [{ id: 'tag-abc', name: 'nature' }]
    // nature is applied, so it should not appear even if passed in suggestions
    render(
      <TagInput
        tags={appliedTags}
        onChange={vi.fn()}
        suggestions={appliedTags}
      />,
    )

    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'nat')

    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })
})
