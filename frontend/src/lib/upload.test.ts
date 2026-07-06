import { vi, describe, it, expect, beforeEach } from 'vitest'

vi.mock('@/lib/images', () => ({
  initiateUpload: vi.fn(),
  putToR2: vi.fn(),
  completeUpload: vi.fn(),
}))

vi.mock('@/lib/thumbnail', () => ({
  generateThumbnail: vi.fn().mockResolvedValue({
    blob: new Blob(['thumb'], { type: 'image/jpeg' }),
    width: 1920,
    height: 1080,
  }),
  convertHeicToJpeg: vi.fn().mockResolvedValue(new Blob(['jpeg'], { type: 'image/jpeg' })),
}))

vi.mock('@/lib/browser', () => ({
  isSafari: vi.fn(),
}))

import { validateImageFile, fileBaseName, uploadImageFile } from './upload'
import { initiateUpload, putToR2, completeUpload } from '@/lib/images'
import { generateThumbnail, convertHeicToJpeg } from '@/lib/thumbnail'
import { isSafari } from '@/lib/browser'

const getToken = vi.fn().mockResolvedValue('token')

function makeFile(name: string, type: string): File {
  return new File(['bytes'], name, { type })
}

const defaultInitiateResult = {
  id: 'upload-1',
  upload_url: 'https://r2.example.com/upload',
  thumbnail_upload_url: 'https://r2.example.com/thumb',
  r2_path: 'path',
}

describe('validateImageFile', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(isSafari).mockReturnValue(false)
  })

  it('returns null for an accepted file type', () => {
    expect(validateImageFile(makeFile('photo.png', 'image/png'))).toBeNull()
  })

  it('returns unsupported_type for an unaccepted file type', () => {
    expect(validateImageFile(makeFile('doc.pdf', 'application/pdf'))).toBe('unsupported_type')
  })

  it('returns heic_safari_only for HEIC files on non-Safari browsers', () => {
    vi.mocked(isSafari).mockReturnValue(false)
    expect(validateImageFile(makeFile('photo.heic', 'image/heic'))).toBe('heic_safari_only')
  })

  it('returns null for HEIC files on Safari', () => {
    vi.mocked(isSafari).mockReturnValue(true)
    expect(validateImageFile(makeFile('photo.heic', 'image/heic'))).toBeNull()
  })
})

describe('fileBaseName', () => {
  it('strips the extension', () => {
    expect(fileBaseName('sunset.jpg')).toBe('sunset')
  })
})

describe('uploadImageFile', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    vi.mocked(isSafari).mockReturnValue(false)
  })

  it('completes the 4-step pipeline and returns the complete-upload result', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-new', duplicates: [] })

    const file = makeFile('sunset.jpg', 'image/jpeg')
    const result = await uploadImageFile(getToken, { file, folderId: 'folder-1' })

    expect(initiateUpload).toHaveBeenCalledWith(getToken, {
      title: 'sunset',
      mimeType: 'image/jpeg',
      folderId: 'folder-1',
      description: undefined,
      sourceUrl: undefined,
    })
    expect(generateThumbnail).toHaveBeenCalledWith(file)
    expect(putToR2).toHaveBeenCalledWith('https://r2.example.com/upload', file)
    expect(putToR2).toHaveBeenCalledWith('https://r2.example.com/thumb', expect.any(Blob))
    expect(completeUpload).toHaveBeenCalledWith(getToken, 'upload-1', { width: 1920, height: 1080, fileSize: file.size })
    expect(result).toEqual({ image_id: 'img-new', duplicates: [] })
  })

  it('defaults title to the file base name when not provided', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

    const file = makeFile('vacation-photo.jpg', 'image/jpeg')
    await uploadImageFile(getToken, { file, folderId: null })

    expect(initiateUpload).toHaveBeenCalledWith(getToken, expect.objectContaining({ title: 'vacation-photo' }))
  })

  it('uses the provided title, description, and sourceUrl when given', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

    const file = makeFile('vacation-photo.jpg', 'image/jpeg')
    await uploadImageFile(getToken, {
      file,
      folderId: 'folder-1',
      title: 'My Title',
      description: 'A nice photo',
      sourceUrl: 'https://example.com/ref',
    })

    expect(initiateUpload).toHaveBeenCalledWith(getToken, {
      title: 'My Title',
      mimeType: 'image/jpeg',
      folderId: 'folder-1',
      description: 'A nice photo',
      sourceUrl: 'https://example.com/ref',
    })
  })

  it('accepts webp files', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

    const file = makeFile('image.webp', 'image/webp')
    await uploadImageFile(getToken, { file, folderId: null })

    expect(initiateUpload).toHaveBeenCalledWith(getToken, expect.objectContaining({ mimeType: 'image/webp' }))
  })

  it('accepts avif files', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

    const file = makeFile('image.avif', 'image/avif')
    await uploadImageFile(getToken, { file, folderId: null })

    expect(initiateUpload).toHaveBeenCalledWith(getToken, expect.objectContaining({ mimeType: 'image/avif' }))
  })

  it('converts HEIC files to JPEG before uploading', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockResolvedValueOnce({ image_id: 'img-1', duplicates: [] })

    const file = makeFile('photo.heic', 'image/heic')
    await uploadImageFile(getToken, { file, folderId: null })

    expect(convertHeicToJpeg).toHaveBeenCalledWith(file)
    expect(initiateUpload).toHaveBeenCalledWith(getToken, expect.objectContaining({ mimeType: 'image/jpeg' }))
    expect(generateThumbnail).toHaveBeenCalledWith(expect.any(Blob))
  })

  it('propagates errors from initiateUpload', async () => {
    vi.mocked(initiateUpload).mockRejectedValueOnce(new Error('server error'))

    const file = makeFile('photo.png', 'image/png')
    await expect(uploadImageFile(getToken, { file, folderId: null })).rejects.toThrow('server error')
  })

  it('propagates errors from putToR2 and does not call completeUpload', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2)
      .mockResolvedValueOnce(undefined)
      .mockRejectedValueOnce(new Error('thumbnail upload failed'))

    const file = makeFile('photo.png', 'image/png')
    await expect(uploadImageFile(getToken, { file, folderId: null })).rejects.toThrow('thumbnail upload failed')
    expect(completeUpload).not.toHaveBeenCalled()
  })

  it('propagates errors from completeUpload', async () => {
    vi.mocked(initiateUpload).mockResolvedValueOnce(defaultInitiateResult)
    vi.mocked(putToR2).mockResolvedValue(undefined)
    vi.mocked(completeUpload).mockRejectedValueOnce(new Error('complete failed'))

    const file = makeFile('photo.png', 'image/png')
    await expect(uploadImageFile(getToken, { file, folderId: null })).rejects.toThrow('complete failed')
  })
})
