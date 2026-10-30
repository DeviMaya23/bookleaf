import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import type { Folder } from '@/lib/folders'
import type { AppView } from '@/lib/view'
import type { FolderNode } from '../lib/folderTree'
import { useFolderMutations } from './useFolderMutations'

type NameDialogState =
  | { mode: 'create-root' }
  | { mode: 'create-sub'; parent: Folder }
  | { mode: 'rename'; target: Folder }
  | null

export function useFolderSidebarDialogs(view: AppView, onMobileClose?: () => void) {
  const navigate = useNavigate()
  const { createMutation, renameMutation, deleteMutation, changeIconMutation } = useFolderMutations()

  const [nameDialog, setNameDialog] = useState<NameDialogState>(null)
  const [deleteTarget, setDeleteTarget] = useState<Folder | null>(null)

  function handleDelete() {
    if (!deleteTarget) return
    const targetId = deleteTarget.id
    const isViewingTarget = view.type === 'folder' && view.id === targetId
    deleteMutation.mutate(targetId, {
      onSuccess: () => {
        if (isViewingTarget) navigate('/app')
      },
    })
    setDeleteTarget(null)
  }

  function handleFolderSelect(folder: FolderNode, onFolderSelect?: () => void) {
    const isActive = view.type === 'folder' && view.id === folder.id
    if (!isActive) onFolderSelect?.()
    navigate(`/app/folders/${folder.id}`)
    onMobileClose?.()
  }

  function handleViewDetails(folder: Folder, onFolderViewDetails?: () => void) {
    const isActive = view.type === 'folder' && view.id === folder.id
    if (!isActive) navigate(`/app/folders/${folder.id}`)
    onFolderViewDetails?.()
    onMobileClose?.()
  }

  function handleNameDialogSubmit(name: string) {
    if (!nameDialog) return
    if (nameDialog.mode === 'create-root') {
      createMutation.mutate({ name })
    } else if (nameDialog.mode === 'create-sub') {
      createMutation.mutate({ name, parentId: nameDialog.parent.id })
    } else {
      renameMutation.mutate({ id: nameDialog.target.id, name })
    }
  }

  const nameDialogTitle =
    nameDialog?.mode === 'create-sub'
      ? 'New subfolder'
      : nameDialog?.mode === 'rename'
      ? 'Rename folder'
      : 'New folder'
  const nameDialogInitialValue = nameDialog?.mode === 'rename' ? nameDialog.target.name : ''

  return {
    nameDialog,
    setNameDialog,
    deleteTarget,
    setDeleteTarget,
    handleDelete,
    handleFolderSelect,
    handleViewDetails,
    handleNameDialogSubmit,
    nameDialogTitle,
    nameDialogInitialValue,
    changeIconMutation,
  }
}
