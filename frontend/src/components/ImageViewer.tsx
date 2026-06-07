import { useCallback, useEffect, useRef, useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { ChevronLeft, FlipHorizontal, RotateCw } from 'lucide-react'
import { getImage } from '@/lib/images'
import type { Image } from '@/lib/images'

interface ImageViewerProps {
  image: Image
  onClose: () => void
}

export default function ImageViewer({ image, onClose }: ImageViewerProps) {
  const { getToken } = useKindeAuth()

  const { data: imageDetail } = useQuery({
    queryKey: ['image', image.id],
    queryFn: () => getImage(getToken, image.id),
  })

  // Transform state
  const [zoom, setZoom] = useState(0.5)
  const [pan, setPan] = useState({ x: 0, y: 0 })
  const [rotation, setRotation] = useState<0 | 90 | 180 | 270>(0)
  const [flipped, setFlipped] = useState(false)

  // Refs for stale-closure-safe access in event handlers
  const containerRef = useRef<HTMLDivElement>(null)
  const zoomRef = useRef(zoom)
  const panRef = useRef(pan)
  const rotationRef = useRef<0 | 90 | 180 | 270>(0)
  useEffect(() => { zoomRef.current = zoom }, [zoom])
  useEffect(() => { panRef.current = pan }, [pan])
  useEffect(() => { rotationRef.current = rotation }, [rotation])

  // Drag pan state
  const [dragging, setDragging] = useState(false)
  const dragOriginRef = useRef<{ x: number; y: number } | null>(null)
  const panSnapshotRef = useRef({ x: 0, y: 0 })

  const calcFit = useCallback((rot: number): number => {
    const el = containerRef.current
    if (!el || !image.width || !image.height) return 0.5
    const containerW = el.clientWidth || el.getBoundingClientRect().width
    const containerH = el.clientHeight || el.getBoundingClientRect().height
    if (!containerW || !containerH) return 0.5
    const isSwapped = rot % 180 !== 0
    const fitW = isSwapped ? image.height : image.width
    const fitH = isSwapped ? image.width : image.height
    return Math.min(containerW / fitW!, containerH / fitH!) * 0.9
  }, [image.width, image.height])

  // Reset all transforms when image changes (also covers fit-on-open, since this
  // effect fires on mount too)
  useEffect(() => {
    setZoom(calcFit(0))
    setPan({ x: 0, y: 0 })
    setRotation(0)
    setFlipped(false)
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [image.id])

  // ResizeObserver: recalculate fit when container resizes
  useEffect(() => {
    const el = containerRef.current
    if (!el) return
    const observer = new ResizeObserver(() => {
      setZoom(calcFit(rotationRef.current))
      setPan({ x: 0, y: 0 })
    })
    observer.observe(el)
    return () => observer.disconnect()
  }, [calcFit])

  // Reset zoom/pan after rotation; use a ref to detect actual changes
  const prevRotationRef = useRef<number | null>(null)
  useEffect(() => {
    if (prevRotationRef.current === null) {
      prevRotationRef.current = rotation
      return
    }
    if (prevRotationRef.current === rotation) return
    prevRotationRef.current = rotation
    setZoom(calcFit(rotation))
    setPan({ x: 0, y: 0 })
  }, [rotation, calcFit])

  // Wheel zoom — native listener (passive: false required for preventDefault in Chrome)
  useEffect(() => {
    const el = containerRef.current
    if (!el) return

    function handleWheel(e: WheelEvent) {
      e.preventDefault()
      const currentZoom = zoomRef.current
      const currentPan = panRef.current
      const factor = e.deltaY < 0 ? 1.1 : 1 / 1.1
      const newZoom = Math.min(Math.max(currentZoom * factor, 0.05), 8)
      const rect = el!.getBoundingClientRect()
      const cx = e.clientX - rect.left - rect.width / 2
      const cy = e.clientY - rect.top - rect.height / 2
      const newPanX = cx - (cx - currentPan.x) * (newZoom / currentZoom)
      const newPanY = cy - (cy - currentPan.y) * (newZoom / currentZoom)
      setZoom(newZoom)
      setPan({ x: newPanX, y: newPanY })
    }

    el.addEventListener('wheel', handleWheel, { passive: false })
    return () => el.removeEventListener('wheel', handleWheel)
  }, [])

  // Esc to close
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key === 'Escape') onClose()
    }
    window.addEventListener('keydown', handleKeyDown)
    return () => window.removeEventListener('keydown', handleKeyDown)
  }, [onClose])

  const displaySrc = imageDetail?.image_url ?? image.thumbnail_url

  const transform = `translate(-50%, -50%) translate(${pan.x}px, ${pan.y}px) scale(${zoom}) scaleX(${flipped ? -1 : 1}) rotate(${rotation}deg)`

  return (
    <div className="flex flex-col w-full h-full">
      <div className="flex items-center h-11 px-2 gap-2 flex-shrink-0 border-b">
        <button onClick={onClose} aria-label="Back" className="p-1.5 rounded hover:bg-muted transition-colors">
          <ChevronLeft className="w-5 h-5" />
        </button>
        <input
          type="range"
          min={5}
          max={800}
          value={Math.round(zoom * 100)}
          onChange={(e) => setZoom(Number(e.target.value) / 100)}
          className="w-32"
        />
        <span className="text-sm tabular-nums">{Math.round(zoom * 100)}%</span>
        <div className="h-4 w-px bg-border mx-1" />
        <button
          aria-label="Flip horizontal"
          className="p-1.5 rounded hover:bg-muted transition-colors cursor-pointer"
          onClick={() => setFlipped((f) => !f)}
        >
          <FlipHorizontal className="w-4 h-4" />
        </button>
        <button
          aria-label="Rotate 90° clockwise"
          className="p-1.5 rounded hover:bg-muted transition-colors cursor-pointer"
          onClick={() => setRotation((r) => ((r + 90) % 360) as 0 | 90 | 180 | 270)}
        >
          <RotateCw className="w-4 h-4" />
        </button>
        <button
          className="text-sm px-2 py-1 rounded hover:bg-muted transition-colors cursor-pointer"
          onClick={() => { setZoom(1); setPan({ x: 0, y: 0 }) }}
        >
          1:1
        </button>
        <div className="flex-1" />
      </div>
      <div
        ref={containerRef}
        className="flex-1 overflow-hidden relative"
        style={{ cursor: dragging ? 'grabbing' : 'grab' }}
        onMouseDown={(e) => {
          setDragging(true)
          dragOriginRef.current = { x: e.clientX, y: e.clientY }
          panSnapshotRef.current = panRef.current
        }}
        onMouseMove={(e) => {
          if (!dragging || !dragOriginRef.current) return
          const origin = dragOriginRef.current
          const snapshot = panSnapshotRef.current
          setPan({ x: snapshot.x + (e.clientX - origin.x), y: snapshot.y + (e.clientY - origin.y) })
        }}
        onMouseUp={() => { setDragging(false); dragOriginRef.current = null }}
        onMouseLeave={() => { setDragging(false); dragOriginRef.current = null }}
      >
        {displaySrc && (
          <img
            src={displaySrc}
            alt={image.title}
            draggable={false}
            style={{ position: 'absolute', left: '50%', top: '50%', transform }}
          />
        )}
        <div className="absolute bottom-4 left-1/2 -translate-x-1/2 bg-black/50 text-white text-xs px-3 py-1 rounded-full whitespace-nowrap pointer-events-none">
          {`${image.title} · ${image.width} × ${image.height}`}
        </div>
      </div>
    </div>
  )
}
