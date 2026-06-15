import { render, screen, fireEvent } from '@testing-library/react'
import { vi, describe, it, expect } from 'vitest'
import Lightbox from './Lightbox'
import type { SharedImage } from '@/lib/share'

function makeImages(): SharedImage[] {
  return [
    { title: 'First', thumbnail_url: null, full_res_url: 'https://example.com/1.jpg', download_url: 'https://example.com/1.jpg?download', width: 100, height: 100 },
    { title: 'Second', thumbnail_url: null, full_res_url: 'https://example.com/2.jpg', download_url: 'https://example.com/2.jpg?download', width: 100, height: 100 },
    { title: 'Third', thumbnail_url: null, full_res_url: 'https://example.com/3.jpg', download_url: 'https://example.com/3.jpg?download', width: 100, height: 100 },
  ]
}

describe('Lightbox', () => {
  it('opens at the given index', () => {
    render(<Lightbox images={makeImages()} index={1} onClose={vi.fn()} onNavigate={vi.fn()} />)

    expect(screen.getByAltText('Second')).toBeInTheDocument()
    expect(screen.getByText('2 / 3')).toBeInTheDocument()
  })

  it('calls onNavigate(1) when ArrowRight is pressed and not at the last image', () => {
    const onNavigate = vi.fn()
    render(<Lightbox images={makeImages()} index={0} onClose={vi.fn()} onNavigate={onNavigate} />)

    fireEvent.keyDown(window, { key: 'ArrowRight' })

    expect(onNavigate).toHaveBeenCalledWith(1)
  })

  it('does not call onNavigate when ArrowRight is pressed at the last image', () => {
    const onNavigate = vi.fn()
    render(<Lightbox images={makeImages()} index={2} onClose={vi.fn()} onNavigate={onNavigate} />)

    fireEvent.keyDown(window, { key: 'ArrowRight' })

    expect(onNavigate).not.toHaveBeenCalled()
  })

  it('calls onNavigate(-1) when ArrowLeft is pressed and not at the first image', () => {
    const onNavigate = vi.fn()
    render(<Lightbox images={makeImages()} index={1} onClose={vi.fn()} onNavigate={onNavigate} />)

    fireEvent.keyDown(window, { key: 'ArrowLeft' })

    expect(onNavigate).toHaveBeenCalledWith(-1)
  })

  it('does not call onNavigate when ArrowLeft is pressed at the first image', () => {
    const onNavigate = vi.fn()
    render(<Lightbox images={makeImages()} index={0} onClose={vi.fn()} onNavigate={onNavigate} />)

    fireEvent.keyDown(window, { key: 'ArrowLeft' })

    expect(onNavigate).not.toHaveBeenCalled()
  })

  it('calls onClose when Escape is pressed', () => {
    const onClose = vi.fn()
    render(<Lightbox images={makeImages()} index={0} onClose={onClose} onNavigate={vi.fn()} />)

    fireEvent.keyDown(window, { key: 'Escape' })

    expect(onClose).toHaveBeenCalled()
  })

  it('calls onClose when the backdrop is clicked', () => {
    const onClose = vi.fn()
    const { container } = render(<Lightbox images={makeImages()} index={0} onClose={onClose} onNavigate={vi.fn()} />)

    fireEvent.click(container.firstChild as Element)

    expect(onClose).toHaveBeenCalled()
  })

  it('does not call onClose when the image is clicked', () => {
    const onClose = vi.fn()
    render(<Lightbox images={makeImages()} index={0} onClose={onClose} onNavigate={vi.fn()} />)

    fireEvent.click(screen.getByAltText('First'))

    expect(onClose).not.toHaveBeenCalled()
  })
})
