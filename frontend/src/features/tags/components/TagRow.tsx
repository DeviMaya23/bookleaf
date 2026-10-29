import { useRef, useState, useEffect } from 'react'
import { MoreHorizontal, Pencil, Trash2 } from 'lucide-react'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { useRenameMutation, useDeleteMutation } from '../hooks/useTagMutations'
import type { TagWithCount } from '@/lib/tags'

type RowMode = 'view' | 'rename' | 'delete-confirm'

interface TagRowProps {
  tag: TagWithCount
  onClick: () => void
}

export default function TagRow({ tag, onClick }: TagRowProps) {
  const [mode, setMode] = useState<RowMode>('view')
  const [renameValue, setRenameValue] = useState(tag.name)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (mode === 'rename') {
      setRenameValue(tag.name)
      setTimeout(() => {
        inputRef.current?.focus()
        inputRef.current?.select()
      }, 0)
    }
  }, [mode, tag.name])

  const rename = useRenameMutation(() => setMode('view'))
  const del = useDeleteMutation(() => {})

  function handleRenameSubmit() {
    const trimmed = renameValue.trim()
    if (!trimmed || trimmed === tag.name) {
      setMode('view')
      return
    }
    rename.mutate(tag.id, trimmed)
  }

  if (mode === 'rename') {
    return (
      <div className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-accent/50">
        <div className="flex-1 min-w-0">
          <input
            ref={inputRef}
            value={renameValue}
            onChange={(e) => {
              setRenameValue(e.target.value)
              rename.clearDuplicateError()
            }}
            onKeyDown={(e) => {
              if (e.key === 'Enter') handleRenameSubmit()
              if (e.key === 'Escape') setMode('view')
            }}
            className="w-full bg-background border rounded px-2 py-0.5 text-sm outline-none focus:ring-1 focus:ring-primary/40"
          />
          {rename.duplicateError && (
            <p className="text-xs text-destructive mt-0.5">That tag already exists. Try another name.</p>
          )}
        </div>
        <button
          className="text-xs text-muted-foreground hover:text-foreground shrink-0"
          onClick={handleRenameSubmit}
          disabled={rename.isPending}
        >
          Save
        </button>
        <button
          className="text-xs text-muted-foreground hover:text-foreground shrink-0"
          onClick={() => setMode('view')}
        >
          Cancel
        </button>
      </div>
    )
  }

  if (mode === 'delete-confirm') {
    return (
      <div className="flex items-center gap-2 px-3 py-1.5 rounded-md bg-accent/50">
        <span className="flex-1 text-sm text-muted-foreground">Remove tag?</span>
        <button
          className="text-xs text-muted-foreground hover:text-foreground shrink-0"
          onClick={() => setMode('view')}
        >
          Cancel
        </button>
        <button
          className="text-xs text-destructive hover:text-destructive/80 font-medium shrink-0"
          onClick={() => del.mutate(tag.id)}
          disabled={del.isPending}
        >
          Delete
        </button>
      </div>
    )
  }

  return (
    <div className="group flex items-center gap-2 px-3 py-1.5 rounded-md hover:bg-accent/50">
      <button
        className="flex-1 text-left text-sm truncate cursor-pointer"
        onClick={onClick}
      >
        {tag.name}
      </button>
      <div className="relative shrink-0 flex items-center justify-end w-6">
        <span className="text-xs text-muted-foreground group-hover:invisible">{tag.image_count}</span>
        <DropdownMenu>
          <DropdownMenuTrigger className="invisible group-hover:visible absolute inset-0 flex items-center justify-center rounded text-muted-foreground hover:text-foreground hover:bg-accent">
            <MoreHorizontal className="w-3.5 h-3.5" />
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end">
            <DropdownMenuItem onClick={() => setMode('rename')}>
              <Pencil className="w-3.5 h-3.5" />
              Rename
            </DropdownMenuItem>
            <DropdownMenuItem
              onClick={() => setMode('delete-confirm')}
              className="text-destructive focus:text-destructive"
            >
              <Trash2 className="w-3.5 h-3.5" />
              Delete
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </div>
    </div>
  )
}
