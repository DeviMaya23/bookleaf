import { render, fireEvent, screen } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useImageTransform } from './useImageTransform'
import type { Image } from '@/lib/images'

function makeImage(overrides?: Partial<Image>): Image {
  return {
    id: 'img-1',
    title: 'Sunset photo',
    description: null,
    mime_type: 'image/jpeg',
    source_url: null,
    folder_ids: [],
    thumbnail_url: 'https://example.com/thumb.jpg',
    width: 1920,
    height: 1080,
    file_size: null,
    tags: [],
    position: null,
    created_at: '2026-05-01T00:00:00Z',
    updated_at: '2026-05-01T00:00:00Z',
    ...overrides,
  }
}

function Harness({ image }: { image: Image }) {
  const { containerRef, transform, zoom, setZoom, dragging, dragHandlers, toggleFlip, rotate } = useImageTransform(image)
  return (
    <div>
      <div ref={containerRef} data-testid="container" {...dragHandlers}>
        <span data-testid="zoom">{zoom}</span>
        <span data-testid="transform">{transform}</span>
        <span data-testid="dragging">{String(dragging)}</span>
      </div>
      <button onClick={rotate}>rotate</button>
      <button onClick={toggleFlip}>flip</button>
      <button onClick={() => setZoom(2)}>setZoom</button>
    </div>
  )
}

function getZoom() {
  return Number(screen.getByTestId('zoom').textContent)
}

function getTransform() {
  return screen.getByTestId('transform').textContent ?? ''
}

function parsePan(transform: string) {
  const match = transform.match(/translate\((-?[\d.]+)px, (-?[\d.]+)px\)/)
  if (!match) throw new Error(`pan not found in transform: ${transform}`)
  return { x: Number(match[1]), y: Number(match[2]) }
}

function parseScale(transform: string) {
  const match = transform.match(/scale\((-?[\d.]+)\)/)
  if (!match) throw new Error(`scale not found in transform: ${transform}`)
  return Number(match[1])
}

describe('useImageTransform — wheel zoom', () => {
  it('zooms in and pans toward the cursor position', () => {
    render(<Harness image={makeImage()} />)

    const container = screen.getByTestId('container')
    Object.defineProperty(container, 'getBoundingClientRect', {
      configurable: true,
      value: () => ({ width: 800, height: 600, left: 0, top: 0 }),
    })

    const initialZoom = getZoom()

    fireEvent.wheel(container, { deltaY: -100, clientX: 500, clientY: 400 })

    const newZoom = getZoom()
    expect(newZoom).toBeGreaterThan(initialZoom)
    expect(newZoom).toBeCloseTo(initialZoom * 1.1, 5)

    const pan = parsePan(getTransform())
    // cursor at (500, 400) is offset (100, 100) from the container center (400, 300)
    expect(pan.x).toBeCloseTo(-10, 5)
    expect(pan.y).toBeCloseTo(-10, 5)
  })
})

describe('useImageTransform — drag to pan', () => {
  it('updates pan on mousemove while dragging, and stops on mouseup', () => {
    render(<Harness image={makeImage()} />)

    const container = screen.getByTestId('container')

    fireEvent.mouseDown(container, { clientX: 100, clientY: 100 })
    expect(screen.getByTestId('dragging').textContent).toBe('true')

    fireEvent.mouseMove(container, { clientX: 150, clientY: 130 })
    let pan = parsePan(getTransform())
    expect(pan.x).toBeCloseTo(50, 5)
    expect(pan.y).toBeCloseTo(30, 5)

    fireEvent.mouseUp(container)
    expect(screen.getByTestId('dragging').textContent).toBe('false')

    fireEvent.mouseMove(container, { clientX: 200, clientY: 200 })
    pan = parsePan(getTransform())
    expect(pan.x).toBeCloseTo(50, 5)
    expect(pan.y).toBeCloseTo(30, 5)
  })

  it('stops dragging on mouseleave', () => {
    render(<Harness image={makeImage()} />)

    const container = screen.getByTestId('container')

    fireEvent.mouseDown(container, { clientX: 0, clientY: 0 })
    fireEvent.mouseLeave(container)
    expect(screen.getByTestId('dragging').textContent).toBe('false')

    fireEvent.mouseMove(container, { clientX: 100, clientY: 100 })
    const pan = parsePan(getTransform())
    expect(pan.x).toBe(0)
    expect(pan.y).toBe(0)
  })
})

describe('useImageTransform — reset on image change', () => {
  it('resets zoom, pan, rotation and flip when the image prop changes', () => {
    const { rerender } = render(<Harness image={makeImage({ id: 'img-1' })} />)

    const container = screen.getByTestId('container')
    const initialZoom = getZoom()

    fireEvent.click(screen.getByRole('button', { name: 'rotate' }))
    fireEvent.click(screen.getByRole('button', { name: 'flip' }))
    fireEvent.mouseDown(container, { clientX: 0, clientY: 0 })
    fireEvent.mouseMove(container, { clientX: 40, clientY: 40 })
    fireEvent.mouseUp(container)
    fireEvent.click(screen.getByRole('button', { name: 'setZoom' }))

    expect(getTransform()).toContain('rotate(90deg)')
    expect(getTransform()).toContain('scaleX(-1)')
    expect(parsePan(getTransform())).toEqual({ x: 40, y: 40 })
    expect(parseScale(getTransform())).toBe(2)

    rerender(<Harness image={makeImage({ id: 'img-2' })} />)

    expect(getTransform()).toContain('rotate(0deg)')
    expect(getTransform()).toContain('scaleX(1)')
    expect(parsePan(getTransform())).toEqual({ x: 0, y: 0 })
    expect(getZoom()).toBe(initialZoom)
  })
})
