import { render, screen } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { describe, it, expect } from 'vitest'
import ExtensionsSection from './ExtensionsSection'

describe('ExtensionsSection', () => {
  it('hides the Firefox manual install steps by default', () => {
    render(<ExtensionsSection />)

    expect(screen.queryByTestId('ff-manual-steps')).not.toBeInTheDocument()
  })

  it('shows the Firefox manual install steps after clicking the toggle', async () => {
    const user = userEvent.setup()
    render(<ExtensionsSection />)

    await user.click(screen.getByTestId('ff-manual-toggle'))

    expect(screen.getByTestId('ff-manual-steps')).toBeInTheDocument()
  })
})
