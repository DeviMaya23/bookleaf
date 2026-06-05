import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import UploadModal from './UploadModal'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/images', () => ({
  initiateUpload: vi.fn(),
  putToR2: vi.fn(),
  completeUpload: vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

import { initiateUpload, putToR2, completeUpload } from '@/lib/images'
import { toast } from 'sonner'

function renderModal(folderId: string | null = null, onUploadSuccess?: (id: string) => void) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const onOpenChange = vi.fn()
  render(
    <QueryClientProvider client={queryClient}>
      <UploadModal open={true} onOpenChange={onOpenChange} folderId={folderId} onUploadSuccess={onUploadSuccess} />
    </QueryClientProvider>,
  )
  return { onOpenChange }
}

function makeImageFile(name = 'photo.jpg', type = 'image/jpeg'): File {
  return new File(['bytes'], name, { type })
}

describe('UploadModal', () => {
  beforeEach(() => vi.clearAllMocks())

  it('closes modal and shows success toast on successful upload', async () => {
    const user = userEvent.setup()
    vi.mocked(initiateUpload).mockResolvedValueOnce({ id: 'upload-1', upload_url: 'https://r2.example.com/upload', r2_path: 'path' })
    vi.mocked(putToR2).mockResolvedValueOnce(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1' })
    const onUploadSuccess = vi.fn()

    const { onOpenChange } = renderModal(null, onUploadSuccess)

    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(input, makeImageFile())
    await user.click(screen.getByRole('button', { name: /upload/i }))

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('Image uploaded successfully')
    })
    expect(onOpenChange).toHaveBeenCalledWith(false)
    expect(onUploadSuccess).toHaveBeenCalledWith('img-1')
  })

  it('uploads with description and source_url when Add details is filled', async () => {
    const user = userEvent.setup()
    vi.mocked(initiateUpload).mockResolvedValueOnce({ id: 'upload-1', upload_url: 'https://r2.example.com/upload', r2_path: 'path' })
    vi.mocked(putToR2).mockResolvedValueOnce(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1' })

    renderModal('folder-1')

    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    const file = makeImageFile()
    await user.upload(input, file)

    await user.click(screen.getByRole('button', { name: /add details/i }))

    const notes = screen.getByPlaceholderText('Notes')
    const sourceUrl = screen.getByPlaceholderText('https://…')
    await user.type(notes, 'A nice sunset photo')
    await user.type(sourceUrl, 'https://example.com/ref')

    await user.click(screen.getByRole('button', { name: /upload/i }))

    await waitFor(() => {
      expect(initiateUpload).toHaveBeenCalledWith(
        expect.any(Function),
        expect.objectContaining({
          description: 'A nice sunset photo',
          sourceUrl: 'https://example.com/ref',
          folderId: 'folder-1',
        }),
      )
    })
    expect(toast.success).toHaveBeenCalledWith('Image uploaded successfully')
  })

  it('shows error toast and keeps modal open when upload fails', async () => {
    const user = userEvent.setup()
    vi.mocked(initiateUpload).mockRejectedValueOnce(new Error('network error'))

    const { onOpenChange } = renderModal()

    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(input, makeImageFile())

    await user.click(screen.getByRole('button', { name: /upload/i }))

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Upload failed. Please try again.')
    })
    expect(onOpenChange).not.toHaveBeenCalledWith(false)
  })
})
