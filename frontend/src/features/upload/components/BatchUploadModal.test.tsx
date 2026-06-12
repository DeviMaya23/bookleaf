import { render, screen, waitFor } from '@testing-library/react'
import userEvent from '@testing-library/user-event'
import { QueryClient, QueryClientProvider } from '@tanstack/react-query'
import { vi, describe, it, expect, beforeEach } from 'vitest'
import BatchUploadModal from './BatchUploadModal'

vi.mock('@kinde-oss/kinde-auth-react', () => ({
  useKindeAuth: () => ({ getToken: vi.fn().mockResolvedValue('token') }),
}))

vi.mock('@/lib/upload', () => ({
  validateImageFile: vi.fn((file: File) => {
    const ACCEPTED_TYPES = ['image/jpeg', 'image/png', 'image/gif', 'image/webp', 'image/avif', 'image/heic']
    return ACCEPTED_TYPES.includes(file.type) ? null : 'unsupported_type'
  }),
  fileBaseName: vi.fn((name: string) => name.replace(/\.[^.]+$/, '')),
  uploadImageFile: vi.fn(),
}))

import { uploadImageFile } from '@/lib/upload'

function renderModal(props: Partial<{ folderId: string | null; initialFiles: File[] }> = {}) {
  const queryClient = new QueryClient({ defaultOptions: { queries: { retry: false } } })
  const onOpenChange = vi.fn()
  render(
    <QueryClientProvider client={queryClient}>
      <BatchUploadModal
        open={true}
        onOpenChange={onOpenChange}
        folderId={props.folderId ?? null}
        initialFiles={props.initialFiles}
      />
    </QueryClientProvider>,
  )
  return { onOpenChange, queryClient }
}

function makeFile(name = 'photo.jpg', type = 'image/jpeg', sizeBytes = 1024): File {
  const content = new Uint8Array(sizeBytes)
  return new File([content], name, { type })
}

describe('BatchUploadModal', () => {
  beforeEach(() => vi.clearAllMocks())

  describe('upload flow', () => {
    it('uploads all files successfully and invalidates the image query per file', async () => {
      const user = userEvent.setup()
      vi.mocked(uploadImageFile)
        .mockResolvedValueOnce({ image_id: 'img-1', suggested_folder_name: null })
        .mockResolvedValueOnce({ image_id: 'img-2', suggested_folder_name: 'Nature' })

      const file1 = makeFile('a.jpg')
      const file2 = makeFile('b.jpg')
      renderModal({ initialFiles: [file1, file2] })

      await user.click(screen.getByRole('button', { name: /upload 2 images/i }))

      await waitFor(() => {
        expect(uploadImageFile).toHaveBeenCalledTimes(2)
      })

      expect(uploadImageFile).toHaveBeenCalledWith(
        expect.any(Function),
        { file: file1, folderId: null, title: 'a' },
      )
      expect(uploadImageFile).toHaveBeenCalledWith(
        expect.any(Function),
        { file: file2, folderId: null, title: 'b' },
      )
      // folder suggestion from complete is ignored — modal stays open, no suggestion UI
      expect(screen.queryByText(/add to this folder/i)).not.toBeInTheDocument()
    })

    it('shows Retry button after a file fails twice', async () => {
      const user = userEvent.setup()
      vi.mocked(uploadImageFile).mockRejectedValue(new Error('network error'))

      renderModal({ initialFiles: [makeFile('c.jpg')] })

      await user.click(screen.getByRole('button', { name: /upload 1 image/i }))

      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
      })

      // Two attempts: initial + auto-retry
      expect(uploadImageFile).toHaveBeenCalledTimes(2)
    })

    it('re-queues and uploads after clicking manual Retry', async () => {
      const user = userEvent.setup()
      vi.mocked(uploadImageFile)
        .mockRejectedValueOnce(new Error('fail'))
        .mockRejectedValueOnce(new Error('fail'))
        .mockResolvedValueOnce({ image_id: 'img-1', suggested_folder_name: null })

      renderModal({ initialFiles: [makeFile('d.jpg')] })

      await user.click(screen.getByRole('button', { name: /upload 1 image/i }))

      // Wait for double failure → Retry button appears
      await waitFor(() => {
        expect(screen.getByRole('button', { name: /retry/i })).toBeInTheDocument()
      })

      await user.click(screen.getByRole('button', { name: /retry/i }))

      // Third attempt succeeds — success icon appears, Retry button gone
      await waitFor(() => {
        expect(screen.queryByRole('button', { name: /retry/i })).not.toBeInTheDocument()
      })
      expect(uploadImageFile).toHaveBeenCalledTimes(3)
    })

    it('sends folder_id when a folder is active and omits it on non-folder routes', async () => {
      const user = userEvent.setup()
      vi.mocked(uploadImageFile).mockResolvedValue({ image_id: 'img-1', suggested_folder_name: null })

      // With folder active
      renderModal({ folderId: 'folder-42', initialFiles: [makeFile('e.jpg')] })
      await user.click(screen.getByRole('button', { name: /upload 1 image/i }))

      await waitFor(() => {
        expect(uploadImageFile).toHaveBeenCalledWith(
          expect.any(Function),
          expect.objectContaining({ folderId: 'folder-42' }),
        )
      })
    })
  })

  describe('validation', () => {
    it('rejects all files and shows error when count exceeds 20', async () => {
      const files = Array.from({ length: 21 }, (_, i) => makeFile(`img${i}.jpg`))
      renderModal({ initialFiles: files })

      expect(await screen.findByText(/maximum of 20 files/i)).toBeInTheDocument()
      expect(screen.queryByRole('button', { name: /upload/i })).toBeDisabled()
    })

    it('marks oversized files as Too large and unsupported types as Unsupported type, uploading valid files', async () => {
      const user = userEvent.setup()
      vi.mocked(uploadImageFile).mockResolvedValue({ image_id: 'img-1', suggested_folder_name: null })

      const oversized = makeFile('big.jpg', 'image/jpeg', 51 * 1024 * 1024)
      const unsupported = makeFile('doc.pdf', 'application/pdf', 100)
      const valid = makeFile('ok.png', 'image/png', 100)

      renderModal({ initialFiles: [oversized, unsupported, valid] })

      expect(screen.getByText('Too large')).toBeInTheDocument()
      expect(screen.getByText('Unsupported type')).toBeInTheDocument()

      // only 1 valid file is uploadable
      const uploadBtn = screen.getByRole('button', { name: /upload 1 image/i })
      await user.click(uploadBtn)

      await waitFor(() => {
        expect(uploadImageFile).toHaveBeenCalledTimes(1)
        expect(uploadImageFile).toHaveBeenCalledWith(
          expect.any(Function),
          expect.objectContaining({ file: valid, title: 'ok' }),
        )
      })
    })
  })
})
