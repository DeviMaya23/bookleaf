import { useCallback, useEffect, useRef, useState, type RefObject } from 'react'
import type { Image } from '@/lib/images'

interface UseImageTransformResult {
  containerRef: RefObject<HTMLDivElement | null>
  transform: string
  zoom: number
  setZoom: (zoom: number) => void
  dragging: boolean
  dragHandlers: {
    onMouseDown: (e: React.MouseEvent) => void
    onMouseMove: (e: React.MouseEvent) => void
    onMouseUp: () => void
    onMouseLeave: () => void
  }
  toggleFlip: () => void
  rotate: () => void
  resetTo1to1: () => void
}

export function useImageTransform(image: Image): UseImageTransformResult {
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

  const transform = `translate(-50%, -50%) translate(${pan.x}px, ${pan.y}px) scale(${zoom}) scaleX(${flipped ? -1 : 1}) rotate(${rotation}deg)`

  const dragHandlers = {
    onMouseDown: (e: React.MouseEvent) => {
      setDragging(true)
      dragOriginRef.current = { x: e.clientX, y: e.clientY }
      panSnapshotRef.current = panRef.current
    },
    onMouseMove: (e: React.MouseEvent) => {
      if (!dragging || !dragOriginRef.current) return
      const origin = dragOriginRef.current
      const snapshot = panSnapshotRef.current
      setPan({ x: snapshot.x + (e.clientX - origin.x), y: snapshot.y + (e.clientY - origin.y) })
    },
    onMouseUp: () => { setDragging(false); dragOriginRef.current = null },
    onMouseLeave: () => { setDragging(false); dragOriginRef.current = null },
  }

  const toggleFlip = useCallback(() => setFlipped((f) => !f), [])
  const rotate = useCallback(() => setRotation((r) => ((r + 90) % 360) as 0 | 90 | 180 | 270), [])
  const resetTo1to1 = useCallback(() => { setZoom(1); setPan({ x: 0, y: 0 }) }, [])

  return {
    containerRef,
    transform,
    zoom,
    setZoom,
    dragging,
    dragHandlers,
    toggleFlip,
    rotate,
    resetTo1to1,
  }
}
