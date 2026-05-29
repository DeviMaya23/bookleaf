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
}

export default function TagInput({ tags, onChange, disabled }: TagInputProps) {
  const [val, setVal] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const commit = (raw: string) => {
    const name = raw.trim().toLowerCase().replace(/,/g, '')
    if (!name) return
    if (tags.some((t) => t.name === name)) {
      setVal('')
      return
    }
    // Pass a placeholder id — caller resolves/creates the real ID
    onChange([...tags, { id: '', name }])
    setVal('')
  }

  const remove = (id: string, name: string) =>
    onChange(tags.filter((t) => (id ? t.id !== id : t.name !== name)))

  const onKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter' || e.key === ',') {
      e.preventDefault()
      commit(val)
    } else if (e.key === 'Backspace' && val === '' && tags.length) {
      const last = tags[tags.length - 1]
      remove(last.id, last.name)
    }
  }

  return (
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
        onChange={(e) => setVal(e.target.value)}
        onKeyDown={onKeyDown}
        onBlur={() => val.trim() && commit(val)}
        placeholder={tags.length === 0 ? 'Add tags…' : ''}
        disabled={disabled}
        className="border-none outline-none bg-transparent text-xs min-w-[70px] flex-1 py-0.5 px-0.5"
      />
    </div>
  )
}
