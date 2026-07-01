import { useState } from 'react'
import type { Folder } from '@/lib/folders'

interface SelectionFolderPickerProps {
  folders: Folder[]
  onPick: (folderId: string) => void
  disabled?: boolean
}

export default function SelectionFolderPicker({ folders, onPick, disabled }: SelectionFolderPickerProps) {
  const [search, setSearch] = useState('')

  const filtered = folders.filter((f) => f.name.toLowerCase().includes(search.trim().toLowerCase()))

  return (
    <div>
      <input
        value={search}
        onChange={(e) => setSearch(e.target.value)}
        placeholder="Search folders…"
        disabled={disabled}
        className="w-full rounded-md border bg-background px-2 py-1.5 text-sm outline-none focus:ring-1 focus:ring-primary/40 disabled:opacity-50"
      />
      <div className="max-h-32 overflow-y-auto rounded-lg border border-border/50 mt-1.5">
        {filtered.length === 0 ? (
          <p className="px-3 py-2 text-sm text-muted-foreground">
            {folders.length === 0 ? 'No folders yet' : 'No folders match your search'}
          </p>
        ) : (
          filtered.map((folder) => (
            <button
              key={folder.id}
              type="button"
              disabled={disabled}
              onClick={() => onPick(folder.id)}
              className="w-full text-left px-3 py-2 text-sm hover:bg-muted/60 disabled:opacity-50 disabled:pointer-events-none border-b border-border/30 last:border-b-0"
            >
              {folder.name}
            </button>
          ))
        )}
      </div>
    </div>
  )
}
