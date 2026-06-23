import { useState } from 'react'
import { useDraggable, useDroppable } from '@dnd-kit/core'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
  ContextMenuSub,
  ContextMenuSubTrigger,
  ContextMenuSubContent,
  ContextMenuSeparator,
} from '@/components/ui/context-menu'
import { getFolderSubtreeIds } from '@/lib/folders'
import type { Folder } from '@/lib/folders'
import type { AppView } from '@/lib/view'
import type { FolderNode } from '../lib/folderTree'
import { FOLDER_ICONS } from '../lib/folderIcons'
import FolderIconView from './FolderIconView'
import { useIsCoarsePointer } from '@/hooks/useIsCoarsePointer'

interface FolderItemProps {
  folder: FolderNode
  depth: number
  view: AppView
  folders: Folder[]
  activeDragType: string | null
  activeDragFolderId: string | null
  iconsEnabled: boolean
  onSelect: (folder: FolderNode) => void
  onRename: (folder: Folder) => void
  onDelete: (folder: Folder) => void
  onNewSubfolder: (folder: Folder) => void
  onChangeIcon: (folder: Folder, icon: string) => void
  onViewDetails?: (folder: Folder) => void
}

export default function FolderItem({
  folder, depth, view, folders,
  activeDragType, activeDragFolderId, iconsEnabled,
  onSelect, onRename, onDelete, onNewSubfolder, onChangeIcon, onViewDetails,
}: FolderItemProps) {
  const isCoarsePointer = useIsCoarsePointer()
  const [open, setOpen] = useState(depth === 0)
  const hasChildren = folder.children.length > 0
  const isActive = view.type === 'folder' && view.id === folder.id

  const subtreeIds = activeDragFolderId
    ? getFolderSubtreeIds(folders, activeDragFolderId)
    : new Set<string>()
  const isInvalidFolderTarget = activeDragType === 'folder' && subtreeIds.has(folder.id)

  const { setNodeRef: setDropRef, isOver } = useDroppable({
    id: `folder-drop-${folder.id}`,
    disabled: isInvalidFolderTarget,
    data: { type: 'folder', folderId: folder.id },
  })

  const { attributes, listeners, setNodeRef: setDragRef, isDragging } = useDraggable({
    id: `folder-drag-${folder.id}`,
    data: { type: 'folder', folderId: folder.id, name: folder.name, parentId: folder.parent_id },
  })

  const isImageDropTarget = activeDragType === 'image' && isOver
  const isFolderDropTarget = activeDragType === 'folder' && isOver && !isInvalidFolderTarget

  const setRef = (el: HTMLDivElement | null) => {
    setDropRef(el)
    setDragRef(el)
  }


  return (
    <div style={{ opacity: isDragging ? 0.4 : 1 }}>
      <ContextMenu>
        <ContextMenuTrigger>
          <div
            ref={setRef}
            {...listeners}
            {...attributes}
            style={{ paddingLeft: 8 + depth * 14 }}
            className={`flex items-center gap-1 pr-2 py-1 rounded-md cursor-pointer mb-0.5 text-sm select-none ${
              isActive
                ? 'bg-accent text-accent-foreground font-medium'
                : isImageDropTarget || isFolderDropTarget
                ? 'bg-accent text-accent-foreground ring-1 ring-primary/40'
                : 'text-muted-foreground hover:bg-accent hover:text-accent-foreground'
            }`}
            onClick={() => onSelect(folder)}
          >
            <span
              onClick={(e) => { e.stopPropagation(); setOpen((o) => !o) }}
              className={`w-3.5 h-3.5 shrink-0 flex items-center justify-center text-muted-foreground/50 text-[9px] transition-transform ${
                hasChildren ? '' : 'invisible'
              } ${open ? 'rotate-90' : ''}`}
            >
              ▶
            </span>
            {iconsEnabled && (
              <FolderIconView icon={folder.icon} className="w-3.5 h-3.5 shrink-0" />
            )}
            <span className="flex-1 truncate">{folder.name}</span>
          </div>
        </ContextMenuTrigger>
        <ContextMenuContent>
          {isCoarsePointer && (
            <>
              <ContextMenuItem onClick={() => onViewDetails?.(folder)}>View details</ContextMenuItem>
              <ContextMenuSeparator />
            </>
          )}
          <ContextMenuItem onClick={() => onNewSubfolder(folder)}>
            New subfolder
          </ContextMenuItem>
          <ContextMenuItem onClick={() => onRename(folder)}>
            Rename
          </ContextMenuItem>
          <ContextMenuSub>
            <ContextMenuSubTrigger>Change icon</ContextMenuSubTrigger>
            <ContextMenuSubContent className="max-h-72 overflow-y-auto">
              {Object.entries(FOLDER_ICONS).map(([key, Icon]) => (
                <ContextMenuItem key={key} onClick={() => onChangeIcon(folder, key)}>
                  <Icon className="w-3.5 h-3.5 shrink-0" />
                  {key}
                </ContextMenuItem>
              ))}
            </ContextMenuSubContent>
          </ContextMenuSub>
          <ContextMenuItem
            onClick={() => onDelete(folder)}
            className="text-destructive focus:text-destructive"
          >
            Delete
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>

      {hasChildren && open && (
        <div>
          {folder.children.map((child) => (
            <FolderItem
              key={child.id}
              folder={child}
              depth={depth + 1}
              view={view}
              folders={folders}
              activeDragType={activeDragType}
              activeDragFolderId={activeDragFolderId}
              iconsEnabled={iconsEnabled}
              onSelect={onSelect}
              onRename={onRename}
              onDelete={onDelete}
              onNewSubfolder={onNewSubfolder}
              onChangeIcon={onChangeIcon}
              onViewDetails={onViewDetails}
            />
          ))}
        </div>
      )}
    </div>
  )
}
