import TokenInput from './TokenInput'
import type { Folder } from '@/lib/folders'

interface FolderInputProps {
  folders: Folder[]
  onChange: (folders: Folder[]) => void
  disabled?: boolean
  suggestions?: Folder[]
}

export default function FolderInput({ folders, onChange, disabled, suggestions }: FolderInputProps) {
  return (
    <TokenInput<Folder>
      items={folders}
      onChange={onChange}
      disabled={disabled}
      suggestions={suggestions}
      placeholder="Add to folder…"
    />
  )
}
