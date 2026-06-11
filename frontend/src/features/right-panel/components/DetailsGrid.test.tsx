import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import DetailsGrid from './DetailsGrid'
import type { Image } from '@/lib/images'

function makeImage(overrides?: Partial<Image>): Image {
  return {
    id: 'img-1',
    title: 'Sunset photo',
    description: 'A nice sunset',
    mime_type: 'image/jpeg',
    source_url: 'https://example.com',
    folder_ids: [],
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 1920,
    height: 1080,
    file_size: 2 * 1024 * 1024,
    tags: [],
    position: null,
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
    ...overrides,
  }
}

describe('DetailsGrid — success scenario', () => {
  it('renders size in MB, dimensions, and the formatted date', () => {
    render(<DetailsGrid image={makeImage()} />)

    const expectedDate = new Date('2026-05-01T00:00:00Z').toLocaleDateString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
    })

    expect(screen.getByText('2.0 MB')).toBeInTheDocument()
    expect(screen.getByText('1920 × 1080')).toBeInTheDocument()
    expect(screen.getByText(expectedDate)).toBeInTheDocument()
  })

  it('renders size in KB and B for smaller file sizes', () => {
    const { rerender } = render(<DetailsGrid image={makeImage({ file_size: 2048 })} />)
    expect(screen.getByText('2.0 KB')).toBeInTheDocument()

    rerender(<DetailsGrid image={makeImage({ file_size: 512 })} />)
    expect(screen.getByText('512 B')).toBeInTheDocument()
  })
})

describe('DetailsGrid — fallback scenario', () => {
  it('renders a dash for null file size and missing dimensions', () => {
    render(<DetailsGrid image={makeImage({ file_size: null, width: null, height: null })} />)

    const dashes = screen.getAllByText('—')
    expect(dashes).toHaveLength(2)
  })
})
