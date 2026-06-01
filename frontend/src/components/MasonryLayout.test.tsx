import { render, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import MasonryLayout, { computeMasonryLayout } from './MasonryLayout'
import type { Image } from '@/lib/images'

function makeImage(id: string, width: number, height: number): Image {
  return {
    id,
    title: `Image ${id}`,
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_ids: [],
    thumbnail_url: null,
    width,
    height,
    file_size: null,
    tags: [],
    position: null,
    created_at: '',
    updated_at: '',
  }
}

describe('computeMasonryLayout', () => {
  it('returns correct column count for a given container width', () => {
    const { numCols, colWidth } = computeMasonryLayout(700)
    expect(numCols).toBe(3)
    expect(colWidth).toBeCloseTo((700 - 12 * 2) / 3)
  })

  it('falls back to 1 column when container width is 0', () => {
    const { numCols } = computeMasonryLayout(0)
    expect(numCols).toBe(1)
  })
})

describe('MasonryLayout', () => {
  it('renders all image titles', () => {
    const images = [makeImage('a', 800, 600), makeImage('b', 400, 600), makeImage('c', 800, 400)]
    render(
      <MasonryLayout
        images={images}
        containerWidth={700}
        renderCard={(image) => <div key={image.id}>{image.title}</div>}
      />
    )
    expect(screen.getByText('Image a')).toBeInTheDocument()
    expect(screen.getByText('Image b')).toBeInTheDocument()
    expect(screen.getByText('Image c')).toBeInTheDocument()
  })
})
