import { useState, useEffect, useRef, useCallback } from 'react'

export function useUploadTriggers() {
  const [uploadOpen, setUploadOpen] = useState(false)
  const [uploadInitialFile, setUploadInitialFile] = useState<File | null>(null)
  const [batchUploadOpen, setBatchUploadOpen] = useState(false)
  const [batchInitialFiles, setBatchInitialFiles] = useState<File[]>([])

  // Mirror uploadOpen in a ref so the paste listener — attached once — can read
  // the latest value without being re-subscribed on every toggle.
  const uploadOpenRef = useRef(uploadOpen)
  useEffect(() => { uploadOpenRef.current = uploadOpen }, [uploadOpen])

  useEffect(() => {
    function handlePaste(event: ClipboardEvent) {
      const tag = (event.target as HTMLElement).tagName
      if (tag === 'INPUT' || tag === 'TEXTAREA') return
      if (uploadOpenRef.current) return
      const items = Array.from(event.clipboardData?.items ?? [])
      const imageItem = items.find((item) => item.kind === 'file' && item.type.startsWith('image/'))
      if (!imageItem) return
      const file = imageItem.getAsFile()
      if (!file) return
      setUploadInitialFile(file)
      setUploadOpen(true)
    }
    document.addEventListener('paste', handlePaste)
    return () => document.removeEventListener('paste', handlePaste)
  }, [])

  const openUpload = useCallback((file?: File) => {
    setUploadInitialFile(file ?? null)
    setUploadOpen(true)
  }, [])

  const closeUpload = useCallback(() => {
    setUploadOpen(false)
    setUploadInitialFile(null)
  }, [])

  const openBatchUpload = useCallback((files?: File[]) => {
    setBatchInitialFiles(files ?? [])
    setBatchUploadOpen(true)
  }, [])

  const closeBatchUpload = useCallback(() => setBatchUploadOpen(false), [])

  return {
    uploadOpen,
    uploadInitialFile,
    batchUploadOpen,
    batchInitialFiles,
    openUpload,
    closeUpload,
    openBatchUpload,
    closeBatchUpload,
  }
}
