import { useRef, useState } from 'react'
import { X } from 'lucide-react'
import type { Folder } from '@/lib/folders'

interface FolderInputProps {
  folders: Folder[]
  onChange: (folders: Folder[]) => void
  disabled?: boolean
  suggestions?: Folder[]
}

export default function FolderInput({ folders, onChange, disabled, suggestions = [] }: FolderInputProps) {
  const [val, setVal] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const [showDropdown, setShowDropdown] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const blurTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const filtered = suggestions.filter(
    (s) =>
      val.trim().length > 0 &&
      s.name.toLowerCase().includes(val.trim().toLowerCase()) &&
      !folders.some((f) => f.id === s.id),
  )

  const dropdownVisible = showDropdown && filtered.length > 0

  const add = (folder: Folder) => {
    onChange([...folders, folder])
    setVal('')
    setSelectedIndex(-1)
    setShowDropdown(false)
  }

  const remove = (id: string) => onChange(folders.filter((f) => f.id !== id))

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'ArrowDown') {
      e.preventDefault()
      if (!dropdownVisible) return
      setSelectedIndex((i) => (i + 1) % filtered.length)
    } else if (e.key === 'ArrowUp') {
      e.preventDefault()
      if (!dropdownVisible) return
      setSelectedIndex((i) => (i <= 0 ? filtered.length - 1 : i - 1))
    } else if (e.key === 'Enter') {
      e.preventDefault()
      if (dropdownVisible && selectedIndex >= 0) {
        add(filtered[selectedIndex])
      }
    } else if (e.key === 'Escape') {
      setVal('')
      setSelectedIndex(-1)
      setShowDropdown(false)
    } else if (e.key === 'Backspace' && val === '' && folders.length) {
      remove(folders[folders.length - 1].id)
    }
  }

  const onValChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setVal(e.target.value)
    setSelectedIndex(-1)
    setShowDropdown(true)
  }

  const onBlur = () => {
    blurTimerRef.current = setTimeout(() => {
      setVal('')
      setShowDropdown(false)
      setSelectedIndex(-1)
    }, 150)
  }

  const onSuggestionMouseDown = (folder: Folder) => {
    if (blurTimerRef.current) clearTimeout(blurTimerRef.current)
    add(folder)
    inputRef.current?.focus()
  }

  return (
    <div className="relative">
      <div
        onClick={() => !disabled && inputRef.current?.focus()}
        className={`flex flex-wrap gap-1.5 px-2 py-1.5 border border-border/50 rounded-lg bg-muted/30 min-h-[38px] items-center cursor-text transition-colors focus-within:border-border ${disabled ? 'opacity-50 cursor-default' : ''}`}
      >
        {folders.map((f) => (
          <span
            key={f.id}
            className="inline-flex items-center gap-1 bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs select-none"
          >
            {f.name}
            {!disabled && (
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation()
                  remove(f.id)
                }}
                className="text-muted-foreground hover:text-foreground transition-colors ml-0.5"
                aria-label={`Remove folder ${f.name}`}
              >
                <X className="w-2.5 h-2.5" />
              </button>
            )}
          </span>
        ))}
        <input
          ref={inputRef}
          value={val}
          onChange={onValChange}
          onKeyDown={onKeyDown}
          onFocus={() => setShowDropdown(true)}
          onBlur={onBlur}
          placeholder={folders.length === 0 ? 'Add to folder…' : ''}
          disabled={disabled}
          className="border-none outline-none bg-transparent text-xs min-w-[70px] flex-1 py-0.5 px-0.5"
        />
      </div>

      {dropdownVisible && (
        <ul className="absolute z-10 left-0 right-0 mt-1 bg-popover border border-border rounded-lg shadow-md overflow-hidden">
          {filtered.map((s, i) => (
            <li
              key={s.id}
              onMouseDown={() => onSuggestionMouseDown(s)}
              className={`px-3 py-1.5 text-xs cursor-pointer select-none ${
                i === selectedIndex
                  ? 'bg-accent text-accent-foreground'
                  : 'hover:bg-muted'
              }`}
            >
              {s.name}
            </li>
          ))}
        </ul>
      )}
    </div>
  )
}
