import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect } from 'vitest'
import TokenInput from './TokenInput'

interface Item {
  id: string
  name: string
}

const nature: Item = { id: 'item-nature', name: 'Nature' }
const travel: Item = { id: 'item-travel', name: 'Travel' }

describe('TokenInput — chip add/remove', () => {
  it('selecting a suggestion calls onChange with the item appended', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[]} onChange={onChange} suggestions={[nature, travel]} placeholder="Add…" />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, 'nat')

    const option = await screen.findByText('Nature')
    await userEvent.click(option)

    expect(onChange).toHaveBeenCalledWith([nature])
  })

  it('removing a chip via the remove button calls onChange without that item', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[nature, travel]} onChange={onChange} placeholder="Add…" />)

    await userEvent.click(screen.getByRole('button', { name: /remove nature/i }))

    expect(onChange).toHaveBeenCalledWith([travel])
  })

  it('Backspace on an empty input removes the last item', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[nature, travel]} onChange={onChange} placeholder="Add…" />)

    const input = screen.getByRole('textbox')
    await userEvent.click(input)
    await userEvent.keyboard('{Backspace}')

    expect(onChange).toHaveBeenCalledWith([nature])
  })
})

describe('TokenInput — dropdown keyboard navigation', () => {
  it('pressing ArrowDown then Enter selects the first suggestion', async () => {
    const onChange = vi.fn()
    const suggestions = [nature, { id: 'item-narrative', name: 'Narrative' }]
    render(<TokenInput items={[]} onChange={onChange} suggestions={suggestions} placeholder="Add…" />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, 'na')
    await userEvent.keyboard('{ArrowDown}{Enter}')

    expect(onChange).toHaveBeenCalledWith([nature])
  })

  it('ArrowDown wraps around to the first option after the last', async () => {
    const onChange = vi.fn()
    const suggestions = [nature, travel]
    render(<TokenInput items={[]} onChange={onChange} suggestions={suggestions} placeholder="Add…" />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, 'a')
    await userEvent.keyboard('{ArrowDown}{ArrowDown}{ArrowDown}{Enter}')

    expect(onChange).toHaveBeenCalledWith([nature])
  })

  it('already-applied items are not shown in the dropdown', async () => {
    render(<TokenInput items={[nature]} onChange={vi.fn()} suggestions={[nature]} placeholder="Add…" />)

    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'nat')

    expect(screen.queryByRole('listitem')).not.toBeInTheDocument()
  })
})

describe('TokenInput — blur-close behavior', () => {
  it('blurring without selecting a suggestion does not call onChange (no createFromText)', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[]} onChange={onChange} suggestions={[nature, travel]} placeholder="Add…" />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, 'nat')
    await userEvent.tab()

    await waitFor(() => {
      expect(onChange).not.toHaveBeenCalled()
    })
  })
})

describe('TokenInput — createFromText', () => {
  const createFromText = (raw: string): Item | null => {
    const name = raw.trim().toLowerCase().replace(/,/g, '')
    if (!name) return null
    return { id: '', name }
  }

  it('Enter without a selected suggestion commits raw text via createFromText', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[]} onChange={onChange} placeholder="Add…" createFromText={createFromText} />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, 'concept{Enter}')

    expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'concept' }])
  })

  it('comma key commits raw text via createFromText', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[]} onChange={onChange} placeholder="Add…" createFromText={createFromText} />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, 'concept,')

    expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'concept' }])
  })

  it('blur with a non-empty value commits raw text via createFromText', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[]} onChange={onChange} placeholder="Add…" createFromText={createFromText} />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.click(input)
    await userEvent.type(input, 'concept')
    await userEvent.tab()

    await waitFor(() => {
      expect(onChange).toHaveBeenCalledWith([{ id: '', name: 'concept' }])
    })
  })

  it('does not call onChange when createFromText returns null for a duplicate', async () => {
    const dedupCreateFromText = (raw: string): Item | null => {
      const name = raw.trim().toLowerCase().replace(/,/g, '')
      if (!name) return null
      if (name === 'nature') return null
      return { id: '', name }
    }
    const onChange = vi.fn()
    render(<TokenInput items={[nature]} onChange={onChange} placeholder="Add…" createFromText={dedupCreateFromText} />)

    const input = screen.getByRole('textbox')
    await userEvent.type(input, 'nature{Enter}')

    expect(onChange).not.toHaveBeenCalled()
  })

  it('does not call onChange when the trimmed value is empty', async () => {
    const onChange = vi.fn()
    render(<TokenInput items={[]} onChange={onChange} placeholder="Add…" createFromText={createFromText} />)

    const input = screen.getByPlaceholderText('Add…')
    await userEvent.type(input, '   {Enter}')

    expect(onChange).not.toHaveBeenCalled()
  })
})
