import TokenInput from './TokenInput'

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

export default function TagInput({ tags, onChange, disabled, suggestions }: TagInputProps) {
  const createFromText = (raw: string): Tag | null => {
    const name = raw.trim().toLowerCase().replace(/,/g, '')
    if (!name) return null
    if (tags.some((t) => t.name === name)) return null
    return { id: '', name }
  }

  return (
    <TokenInput<Tag>
      items={tags}
      onChange={onChange}
      disabled={disabled}
      suggestions={suggestions}
      placeholder="Add tags…"
      createFromText={createFromText}
    />
  )
}
