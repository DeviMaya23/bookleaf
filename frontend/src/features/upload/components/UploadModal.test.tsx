import { render, screen, waitFor, fireEvent } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import UploadModal from './UploadModal'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/upload', () => ({
  validateImageFile: vi.fn().mockReturnValue(null),
  fileBaseName: vi.fn((name: string) => name.replace(/\.[^.]+$/, '')),
  uploadImageFile: vi.fn(),
}))

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}))

import { uploadImageFile, validateImageFile } from '@/lib/upload'
import { toast } from 'sonner'

function renderModal(folderId: string | null = null, initialFile?: File) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const onOpenChange = vi.fn()
  render(
    <QueryClientProvider client={queryClient}>
      <UploadModal open={true} onOpenChange={onOpenChange} folderId={folderId} initialFile={initialFile} />
    </QueryClientProvider>,
  )
  return { onOpenChange }
}

function makeImageFile(name = 'photo.jpg', type = 'image/jpeg'): File {
  return new File(['bytes'], name, { type })
}

describe('UploadModal', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    // mockReset clears the once queue so unconsumed once values from a
    // previous test (e.g. a drop that fired no dataTransfer files) don't
    // carry over and corrupt the next test's mock sequence.
    vi.mocked(validateImageFile).mockReset()
    vi.mocked(validateImageFile).mockReturnValue(null)
  })

  it('closes modal and shows success toast on successful upload', async () => {
    const user = userEvent.setup()
    vi.mocked(uploadImageFile).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

    const { onOpenChange } = renderModal(null)

    const file = makeImageFile()
    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(input, file)
    await user.click(screen.getByRole('button', { name: /upload/i }))

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('Image uploaded successfully')
    })
    expect(uploadImageFile).toHaveBeenCalledWith(expect.any(Function), {
      file,
      folderId: null,
      title: 'photo',
      description: undefined,
      sourceUrl: undefined,
    })
    expect(onOpenChange).toHaveBeenCalledWith(false)
  })

  it('uploads with description and source_url when Add details is filled', async () => {
    const user = userEvent.setup()
    vi.mocked(uploadImageFile).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

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
      expect(uploadImageFile).toHaveBeenCalledWith(
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

  it('stages file and focuses title input when initialFile is provided', async () => {
    const file = makeImageFile('image.png', 'image/png')
    renderModal(null, file)

    await waitFor(() => {
      expect(screen.getByRole('img', { name: 'image.png' })).toBeInTheDocument()
    })
    const titleInput = screen.getByPlaceholderText('image')
    expect(titleInput).toHaveValue('')
    await waitFor(() => expect(titleInput).toHaveFocus())
  })

  it('does not focus title input when opened without initialFile', () => {
    renderModal()

    const titleInput = screen.getByPlaceholderText('Title')
    expect(titleInput).not.toHaveFocus()
  })

  it('rejects an unsupported initialFile type with inline error and stages no file', () => {
    vi.mocked(validateImageFile).mockReturnValueOnce('unsupported_type')
    const file = makeImageFile('document.pdf', 'application/pdf')

    renderModal(null, file)

    expect(screen.getByText('Unsupported file type.')).toBeInTheDocument()
    expect(screen.queryByRole('img', { name: 'document.pdf' })).not.toBeInTheDocument()
  })

  it('shows oversized inline error and disables upload button for an oversized file via initialFile', () => {
    vi.mocked(validateImageFile).mockReturnValueOnce('oversized')
    const file = makeImageFile('huge.jpg', 'image/jpeg')

    renderModal(null, file)

    expect(screen.getByText('File is too large. Maximum size is 50 MB.')).toBeInTheDocument()
    expect(screen.queryByRole('img', { name: 'huge.jpg' })).not.toBeInTheDocument()
    expect(screen.getByRole('button', { name: /upload/i })).toBeDisabled()
  })

  it('shows oversized inline error when an oversized file is selected via the picker', async () => {
    const user = userEvent.setup()
    vi.mocked(validateImageFile).mockReturnValueOnce('oversized')

    renderModal()

    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(input, makeImageFile('huge.jpg', 'image/jpeg'))

    expect(screen.getByText('File is too large. Maximum size is 50 MB.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /upload/i })).toBeDisabled()
  })

  it('shows oversized inline error when an oversized file is dropped onto the drop zone', () => {
    vi.mocked(validateImageFile).mockReturnValueOnce('oversized')
    const file = makeImageFile('huge.jpg', 'image/jpeg')

    renderModal()

    const dropZone = screen.getByRole('button', { name: /drop an image/i })
    fireEvent.drop(dropZone, { dataTransfer: { files: [file] } })

    expect(screen.getByText('File is too large. Maximum size is 50 MB.')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /upload/i })).toBeDisabled()
  })

  it('clears oversized error when a valid file is selected afterwards', async () => {
    const user = userEvent.setup()
    vi.mocked(validateImageFile)
      .mockReturnValueOnce('oversized')
      .mockReturnValueOnce(null)

    const oversizedFile = makeImageFile('huge.jpg', 'image/jpeg')
    renderModal(null, oversizedFile)

    expect(screen.getByText('File is too large. Maximum size is 50 MB.')).toBeInTheDocument()

    const input = document.querySelector('input[type="file"]') as HTMLInputElement
    await user.upload(input, makeImageFile('small.jpg', 'image/jpeg'))

    expect(screen.queryByText('File is too large. Maximum size is 50 MB.')).not.toBeInTheDocument()
    expect(screen.getByRole('img', { name: 'small.jpg' })).toBeInTheDocument()
  })

  it('shows error toast and keeps modal open when upload fails', async () => {
    const user = userEvent.setup()
    vi.mocked(uploadImageFile).mockRejectedValueOnce(new Error('network error'))

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
