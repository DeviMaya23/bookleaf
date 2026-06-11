import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import DownloadButton from './DownloadButton'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', () => ({
  downloadImage: vi.fn(),
}))

import { downloadImage } from '@/lib/images'

describe('DownloadButton — success scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('renders the default state, then a loading state while downloading', async () => {
    let resolveDownload: (url: string) => void = () => {}
    vi.mocked(downloadImage).mockReturnValue(
      new Promise((resolve) => {
        resolveDownload = resolve
      }),
    )

    render(<DownloadButton imageId="img-1" />)

    expect(screen.getByRole('button', { name: /download image/i })).toBeInTheDocument()

    await userEvent.click(screen.getByRole('button', { name: /download image/i }))

    expect(screen.getByRole('button', { name: /downloading/i })).toBeDisabled()

    resolveDownload('https://r2.example.com/download?sig=abc')

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download image/i })).not.toBeDisabled()
    })
    expect(downloadImage).toHaveBeenCalledWith(expect.any(Function), 'img-1')
  })
})

describe('DownloadButton — failure scenario', () => {
  beforeEach(() => {
    vi.clearAllMocks()
  })

  it('re-enables the button and does not crash when downloadImage rejects', async () => {
    vi.mocked(downloadImage).mockRejectedValue(new Error('Network error'))

    render(<DownloadButton imageId="img-1" />)

    const button = screen.getByRole('button', { name: /download image/i })
    await userEvent.click(button)

    await waitFor(() => {
      expect(downloadImage).toHaveBeenCalledWith(expect.any(Function), 'img-1')
    })

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /download image/i })).not.toBeDisabled()
    })
  })
})
