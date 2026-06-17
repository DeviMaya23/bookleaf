import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect, vi } from 'vitest'
import FloatingUploadButton from './FloatingUploadButton'

describe('FloatingUploadButton', () => {
  it('renders the upload button', () => {
    render(<FloatingUploadButton onClick={vi.fn()} />)

    expect(screen.getByRole('button', { name: 'Upload image' })).toBeInTheDocument()
  })

  it('calls onClick when tapped', async () => {
    const onClick = vi.fn()
    render(<FloatingUploadButton onClick={onClick} />)

    await userEvent.click(screen.getByRole('button', { name: 'Upload image' }))

    expect(onClick).toHaveBeenCalledOnce()
  })
})
