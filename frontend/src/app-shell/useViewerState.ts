import { useState, useCallback, useEffect } from 'react'
import type { Image } from '@/lib/images'

interface Params {
  isCoarsePointer: boolean
  viewKey: string
}

export function useViewerState({ isCoarsePointer, viewKey }: Params) {
  const [selectedImage, setSelectedImage] = useState<Image | null>(null)
  const [autoFocusTitle, setAutoFocusTitle] = useState(false)
  const [viewerImage, setViewerImage] = useState<Image | null>(null)
  const [lightboxImage, setLightboxImage] = useState<Image | null>(null)

  useEffect(() => {
    setViewerImage(null)
    setSelectedImage(null)
    setAutoFocusTitle(false)
    setLightboxImage(null)
  }, [viewKey])

  const handleImageSelect = useCallback((img: Image) => {
    if (isCoarsePointer) {
      setLightboxImage(img)
      return
    }
    setAutoFocusTitle(false)
    setSelectedImage(img)
  }, [isCoarsePointer])

  const handleViewDetails = useCallback((img: Image) => {
    setAutoFocusTitle(false)
    setSelectedImage(img)
  }, [])

  const handleImageDoubleClick = useCallback((img: Image) => {
    if (isCoarsePointer) return
    setSelectedImage(img)
    setViewerImage(img)
  }, [isCoarsePointer])

  const handleImageDeleted = useCallback((id: string) => {
    if (selectedImage?.id === id) { setSelectedImage(null); setAutoFocusTitle(false) }
    if (viewerImage?.id === id) setViewerImage(null)
    if (lightboxImage?.id === id) setLightboxImage(null)
  }, [selectedImage, viewerImage, lightboxImage])

  const clearImageSelection = useCallback(() => {
    setSelectedImage(null)
    setAutoFocusTitle(false)
  }, [])

  const selectImageAutoFocus = useCallback((img: Image) => {
    setAutoFocusTitle(true)
    setSelectedImage(img)
  }, [])

  const closeViewer = useCallback(() => setViewerImage(null), [])
  const closeLightbox = useCallback(() => setLightboxImage(null), [])

  return {
    selectedImage,
    autoFocusTitle,
    viewerImage,
    lightboxImage,
    handleImageSelect,
    handleViewDetails,
    handleImageDoubleClick,
    handleImageDeleted,
    clearImageSelection,
    selectImageAutoFocus,
    closeViewer,
    closeLightbox,
  }
}
