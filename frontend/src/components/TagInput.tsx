import { useRef, useState } from 'react'
import { X } from 'lucide-react'

interface Tag {
  id: string
  name: string
}

interface TagInputProps {
  tags: Tag[]
  onChange: (tags: Tag[]) => void
  disabled?: boolean
  suggestions?: Tag[]
}

export default function TagInput({ tags, onChange, disabled, suggestions = [] }: TagInputProps) {
  const [val, setVal] = useState('')
  const [selectedIndex, setSelectedIndex] = useState(-1)
  const [showDropdown, setShowDropdown] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const blurTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null)

  const filtered = suggestions.filter(
    (s) =>
      val.trim().length > 0 &&
      s.name.toLowerCase().includes(val.trim().toLowerCase()) &&
      !tags.some((t) => t.id === s.id),
  )

  const dropdownVisible = showDropdown && filtered.length > 0

  const commitTag = (tag: Tag) => {
    onChange([...tags, tag])
    setVal('')
    setSelectedIndex(-1)
    setShowDropdown(false)
  }

  const commitRaw = (raw: string) => {
    const name = raw.trim().toLowerCase().replace(/,/g, '')
    if (!name) return
    if (tags.some((t) => t.name === name)) {
      setVal('')
      return
    }
    onChange([...tags, { id: '', name }])
    setVal('')
    setSelectedIndex(-1)
    setShowDropdown(false)
  }

  const remove = (id: string, name: string) =>
    onChange(tags.filter((t) => (id ? t.id !== id : t.name !== name)))

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
        commitTag(filtered[selectedIndex])
      } else {
        commitRaw(val)
      }
    } else if (e.key === ',') {
      e.preventDefault()
      commitRaw(val)
    } else if (e.key === 'Escape') {
      setSelectedIndex(-1)
      setShowDropdown(false)
    } else if (e.key === 'Backspace' && val === '' && tags.length) {
      const last = tags[tags.length - 1]
      remove(last.id, last.name)
    }
  }

  const onValChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setVal(e.target.value)
    setSelectedIndex(-1)
    setShowDropdown(true)
  }

  const onBlur = () => {
    blurTimerRef.current = setTimeout(() => {
      if (val.trim()) commitRaw(val)
      setShowDropdown(false)
      setSelectedIndex(-1)
    }, 150)
  }

  const onSuggestionMouseDown = (tag: Tag) => {
    // Cancel the blur timer so the dropdown doesn't close before click fires
    if (blurTimerRef.current) clearTimeout(blurTimerRef.current)
    commitTag(tag)
    inputRef.current?.focus()
  }

  return (
    <div className="relative">
      <div
        onClick={() => !disabled && inputRef.current?.focus()}
        className={`flex flex-wrap gap-1.5 px-2 py-1.5 border border-border/50 rounded-lg bg-muted/30 min-h-[38px] items-center cursor-text transition-colors focus-within:border-border ${disabled ? 'opacity-50 cursor-default' : ''}`}
      >
        {tags.map((t) => (
          <span
            key={t.id || t.name}
            className="inline-flex items-center gap-1 bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs select-none"
          >
            {t.name}
            {!disabled && (
              <button
                type="button"
                onClick={(e) => {
                  e.stopPropagation()
                  remove(t.id, t.name)
                }}
                className="text-muted-foreground hover:text-foreground transition-colors ml-0.5"
                aria-label={`Remove tag ${t.name}`}
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
          placeholder={tags.length === 0 ? 'Add tags…' : ''}
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
