import { useState, useCallback, useEffect } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { bulkAddImagesToFolder, bulkTrashImages, bulkExportImages } from '@/lib/images'
import { bulkAddTags, resolveOrCreateTags } from '@/lib/tags'
import type { Tag } from '@/lib/tags'

interface Params {
  viewKey: string
  clearImageSelection: () => void
}

export function useSelectionMode({ viewKey, clearImageSelection }: Params) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [selectMode, setSelectMode] = useState(false)
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set())
  const [mainSelectedId, setMainSelectedId] = useState<string | null>(null)

  useEffect(() => {
    setSelectMode(false)
    setSelectedIds(new Set())
    setMainSelectedId(null)
  }, [viewKey])

  const exitSelectMode = useCallback(() => {
    setSelectMode(false)
    setSelectedIds(new Set())
    setMainSelectedId(null)
  }, [])

  const bulkTrashMutation = useMutation({
    mutationFn: (imageIds: string[]) => bulkTrashImages(getToken, imageIds),
    onSuccess: (result, imageIds) => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      if (result.succeeded_count < imageIds.length) {
        toast.warning(`Moved ${result.succeeded_count} of ${imageIds.length} images to trash`)
      } else {
        toast.success(`Moved ${result.succeeded_count} image${result.succeeded_count === 1 ? '' : 's'} to trash`)
      }
      exitSelectMode()
    },
    onError: () => {
      toast.error('Failed to move images to trash')
    },
  })

  const handleSelectModeToggle = useCallback((pressed: boolean) => {
    setSelectMode(pressed)
    if (pressed) {
      clearImageSelection()
    } else {
      setSelectedIds(new Set())
      setMainSelectedId(null)
    }
  }, [clearImageSelection])

  const handleApplySelection = useCallback(async ({ folderIds, tags }: { folderIds: string[]; tags: Tag[] }) => {
    const imageIds = Array.from(selectedIds)
    const ops: Promise<void>[] = []

    if (folderIds.length > 0) {
      for (const folderId of folderIds) {
        ops.push(
          bulkAddImagesToFolder(getToken, imageIds, folderId).then((result) => {
            queryClient.invalidateQueries({ queryKey: ['images'] })
            if (result.succeeded_count < imageIds.length) {
              toast.warning(`Added ${result.succeeded_count} of ${imageIds.length} images to folder`)
            }
          }).catch(() => {
            toast.error('Failed to add images to folder')
            throw new Error('folder')
          })
        )
      }
    }

    if (tags.length > 0) {
      ops.push(
        (async () => {
          const allTags = queryClient.getQueryData<Tag[]>(['tags']) ?? []
          const resolved = await resolveOrCreateTags(getToken, tags, allTags, queryClient)
          const resolvedTagIds = resolved.map((t) => t.id)
          const result = await bulkAddTags(getToken, imageIds, resolvedTagIds)
          queryClient.invalidateQueries({ queryKey: ['images'] })
          if (result.succeeded_count < imageIds.length) {
            toast.warning(`Tagged ${result.succeeded_count} of ${imageIds.length} images`)
          }
        })().catch(() => {
          toast.error('Failed to add tags to images')
          throw new Error('tags')
        })
      )
    }

    const results = await Promise.allSettled(ops)
    const allSucceeded = results.every((r) => r.status === 'fulfilled')
    if (allSucceeded && ops.length > 0) {
      toast.success(`Applied to ${imageIds.length} image${imageIds.length === 1 ? '' : 's'}`)
      exitSelectMode()
    }
  }, [selectedIds, getToken, queryClient, exitSelectMode])

  const handleMoveSelectionToTrash = useCallback(() => {
    bulkTrashMutation.mutate(Array.from(selectedIds))
  }, [bulkTrashMutation, selectedIds])

  const handleDownloadSelection = useCallback(async () => {
    try {
      const blob = await bulkExportImages(getToken, Array.from(selectedIds))
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = 'bookleaf-export.zip'
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
    } catch {
      toast.error('Failed to download images')
    }
  }, [getToken, selectedIds])

  const handleSelectionChange = useCallback((ids: Set<string>, anchorId: string | null) => {
    setSelectedIds(ids)
    setMainSelectedId(anchorId)
  }, [])

  return {
    selectMode,
    selectedIds,
    mainSelectedId,
    handleSelectModeToggle,
    exitSelectMode,
    handleApplySelection,
    handleMoveSelectionToTrash,
    handleDownloadSelection,
    handleSelectionChange,
  }
}
