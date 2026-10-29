import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { getTagsWithCount } from '@/lib/tags'
import TagRow from './TagRow'

function groupAlphabetically<T extends { name: string }>(items: T[]): Map<string, T[]> {
  const groups = new Map<string, T[]>()
  for (const item of items) {
    const letter = item.name[0]?.toUpperCase() ?? '#'
    const key = /[A-Z]/.test(letter) ? letter : '#'
    const existing = groups.get(key)
    if (existing) existing.push(item)
    else groups.set(key, [item])
  }
  return groups
}

export default function TagsView() {
  const { getToken } = useKindeAuth()
  const navigate = useNavigate()
  const [search, setSearch] = useState('')

  const { data: tags = [], isLoading } = useQuery({
    queryKey: ['tags', 'with-count'],
    queryFn: () => getTagsWithCount(getToken),
    staleTime: 30_000,
  })

  const filtered = search.trim()
    ? tags.filter((t) => t.name.toLowerCase().includes(search.trim().toLowerCase()))
    : tags

  const sorted = [...filtered].sort((a, b) => a.name.localeCompare(b.name))
  const groups = groupAlphabetically(sorted)
  const sortedKeys = [...groups.keys()].sort((a, b) => {
    if (a === '#') return 1
    if (b === '#') return -1
    return a.localeCompare(b)
  })

  function handleTagClick(tagId: string) {
    navigate('/app', { state: { filterTagIds: [tagId] } })
  }

  return (
    <div className="max-w-2xl">
      <div className="mb-4">
        <input
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          placeholder="Search tags…"
          className="w-full max-w-xs rounded-md border bg-background px-3 py-1.5 text-sm outline-none focus:ring-1 focus:ring-primary/40"
        />
      </div>

      {isLoading && (
        <p className="text-sm text-muted-foreground">Loading…</p>
      )}

      {!isLoading && sorted.length === 0 && (
        <p className="text-sm text-muted-foreground">No tags yet.</p>
      )}

      {sortedKeys.map((letter) => {
        const group = groups.get(letter)!
        return (
          <div key={letter} className="mb-4">
            <div className="flex items-center gap-2 mb-1">
              <span className="text-xs font-semibold tracking-widest uppercase text-muted-foreground/50">
                {letter}
              </span>
              <span className="text-xs text-muted-foreground/50">({group.length})</span>
            </div>
            <div className="space-y-0.5">
              {group.map((tag) => (
                <TagRow
                  key={tag.id}
                  tag={tag}
                  onClick={() => handleTagClick(tag.id)}
                />
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}
