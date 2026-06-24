import type { ReactNode } from 'react'
import { Search, ArrowUpDown, ArrowUp, ArrowDown, Filter, X } from 'lucide-react'
import { cn } from '@/lib/utils'
import { buttonVariants } from '@/components/ui/button-variants'
import {
  DropdownMenu,
  DropdownMenuCheckboxItem,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ToggleGroup, ToggleGroupItem } from '@/components/ui/toggle-group'
import { MIME_TYPE_FILTER_OPTIONS } from '@/lib/images'
import type { AppView } from '@/lib/view'
import {
  FIELD_DEFAULT_DIRECTION,
  type GalleryControls,
  type SortBy,
  type SortDir,
} from '../hooks/useGalleryControls'

const SORT_FIELD_LABELS: Record<SortBy, string> = {
  manual: 'Manual',
  created_at: 'Date added',
  title: 'Name',
  deleted_at: 'Date deleted',
}

const DIR_LABELS: Record<'created_at' | 'title' | 'deleted_at', Record<SortDir, string>> = {
  created_at: { asc: 'Oldest first', desc: 'Newest first' },
  title: { asc: 'A → Z', desc: 'Z → A' },
  deleted_at: { asc: 'Oldest deleted first', desc: 'Newest deleted first' },
}

interface GalleryToolbarProps {
  view: AppView
  controls: GalleryControls
  // The focus-mode toggle and upload actions are owned by the app shell; the
  // toolbar only lays them out alongside the gallery search/sort/filter UI.
  focusToggle: ReactNode
  uploadActions: ReactNode
}

export default function GalleryToolbar({ view, controls, focusToggle, uploadActions }: GalleryToolbarProps) {
  const {
    searchTerm,
    setSearchTerm,
    sortBy,
    sortDir,
    sortFieldOptions,
    sortActive,
    handleSortFieldChange,
    handleSortDirToggle,
    filterSections,
    filterCount,
    filterTagIds,
    setFilterTagIds,
    filterTagSearch,
    setFilterTagSearch,
    filteredTags,
    filterMimeTypes,
    setFilterMimeTypes,
    filterFolderIds,
    setFilterFolderIds,
    filterFolderSearch,
    setFilterFolderSearch,
    filteredFolders,
    activeFilterChips,
    clearAllFilters,
  } = controls

  return (
    <div className="flex flex-col gap-2 mb-4">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-2">
          {focusToggle}
          <div className="flex items-center gap-2 w-full max-w-xs">
            <div className="relative flex-1">
              <Search className="absolute left-2.5 top-1/2 -translate-y-1/2 w-3.5 h-3.5 text-muted-foreground" />
              <input
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
                placeholder="Search images by name…"
                className="w-full rounded-md border bg-background pl-8 pr-3 py-1.5 text-sm outline-none focus:ring-1 focus:ring-primary/40"
              />
            </div>
            <div className="hidden sm:flex">
              <DropdownMenu>
                <DropdownMenuTrigger
                  aria-label="Sort"
                  className={cn(buttonVariants({ variant: sortActive ? 'secondary' : 'outline', size: 'icon' }))}
                >
                  <ArrowUpDown className="w-3.5 h-3.5" />
                </DropdownMenuTrigger>
                <DropdownMenuContent align="start" className="w-48">
                  <DropdownMenuRadioGroup
                    value={sortBy}
                    onValueChange={(value) => handleSortFieldChange(value as SortBy)}
                  >
                    {sortFieldOptions.map((field) => (
                      <DropdownMenuRadioItem key={field} value={field}>
                        {SORT_FIELD_LABELS[field]}
                      </DropdownMenuRadioItem>
                    ))}
                  </DropdownMenuRadioGroup>
                  {sortBy !== 'manual' && (
                    <>
                      <DropdownMenuSeparator />
                      <DropdownMenuItem closeOnClick={false} onClick={handleSortDirToggle}>
                        {sortDir === 'asc' ? <ArrowUp className="w-4 h-4" /> : <ArrowDown className="w-4 h-4" />}
                        {DIR_LABELS[sortBy][sortDir ?? FIELD_DEFAULT_DIRECTION[sortBy]]}
                      </DropdownMenuItem>
                    </>
                  )}
                </DropdownMenuContent>
              </DropdownMenu>
            </div>
          </div>
          {view.type !== 'trash' && (
            <DropdownMenu>
              <DropdownMenuTrigger
                className={cn(buttonVariants({ variant: filterCount > 0 ? 'secondary' : 'outline' }))}
              >
                <Filter className="w-3.5 h-3.5" />
                Filters
                {filterCount > 0 && (
                  <span className="inline-flex items-center justify-center min-w-4 h-4 px-1 rounded-full bg-secondary-foreground/20 text-[10px] font-semibold leading-none">
                    {filterCount}
                  </span>
                )}
              </DropdownMenuTrigger>
              <DropdownMenuContent align="start" className="w-64">
                {filterSections.includes('mimeTypes') && (
                  <DropdownMenuGroup>
                    <DropdownMenuLabel>File type</DropdownMenuLabel>
                    <ToggleGroup
                      multiple
                      value={filterMimeTypes}
                      onValueChange={setFilterMimeTypes}
                      className="w-full flex-wrap gap-1.5 px-1.5 pb-1"
                    >
                      {MIME_TYPE_FILTER_OPTIONS.map((opt) => (
                        <ToggleGroupItem
                          key={opt.value}
                          value={opt.value}
                          className="rounded-full border border-border px-3 text-xs aria-pressed:border-secondary aria-pressed:bg-secondary aria-pressed:text-secondary-foreground"
                        >
                          {opt.label}
                        </ToggleGroupItem>
                      ))}
                    </ToggleGroup>
                  </DropdownMenuGroup>
                )}
                {filterSections.includes('tags') && (
                  <DropdownMenuGroup>
                    {filterSections.includes('mimeTypes') && <DropdownMenuSeparator />}
                    <DropdownMenuLabel>Tags</DropdownMenuLabel>
                    <div className="px-1.5 pb-1">
                      <input
                        value={filterTagSearch}
                        onChange={(e) => setFilterTagSearch(e.target.value)}
                        onKeyDown={(e) => e.stopPropagation()}
                        placeholder="Search tags…"
                        className="w-full rounded-md border bg-background px-2 py-1 text-xs outline-none focus:ring-1 focus:ring-primary/40"
                      />
                    </div>
                    <div className="max-h-40 overflow-y-auto">
                      {filteredTags.map((tag) => (
                        <DropdownMenuCheckboxItem
                          key={tag.id}
                          checked={filterTagIds.includes(tag.id)}
                          onCheckedChange={(checked) =>
                            setFilterTagIds((prev) => checked ? [...prev, tag.id] : prev.filter((id) => id !== tag.id))
                          }
                        >
                          {tag.name}
                        </DropdownMenuCheckboxItem>
                      ))}
                    </div>
                  </DropdownMenuGroup>
                )}
                {filterSections.includes('folders') && (
                  <DropdownMenuGroup>
                    <DropdownMenuSeparator />
                    <DropdownMenuLabel>Folder</DropdownMenuLabel>
                    <div className="px-1.5 pb-1">
                      <input
                        value={filterFolderSearch}
                        onChange={(e) => setFilterFolderSearch(e.target.value)}
                        onKeyDown={(e) => e.stopPropagation()}
                        placeholder="Search folders…"
                        className="w-full rounded-md border bg-background px-2 py-1 text-xs outline-none focus:ring-1 focus:ring-primary/40"
                      />
                    </div>
                    <div className="max-h-40 overflow-y-auto">
                      {filteredFolders.map((folder) => (
                        <DropdownMenuCheckboxItem
                          key={folder.id}
                          checked={filterFolderIds.includes(folder.id)}
                          onCheckedChange={(checked) =>
                            setFilterFolderIds((prev) => checked ? [...prev, folder.id] : prev.filter((id) => id !== folder.id))
                          }
                        >
                          {folder.name}
                        </DropdownMenuCheckboxItem>
                      ))}
                    </div>
                  </DropdownMenuGroup>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
        <div className="hidden sm:flex">{uploadActions}</div>
      </div>
      {activeFilterChips.length > 0 && (
        <div className="flex items-center gap-1.5 flex-wrap">
          {activeFilterChips.map((chip) => (
            <span
              key={chip.key}
              className="inline-flex items-center gap-1 bg-secondary text-secondary-foreground rounded px-2 py-0.5 text-xs"
            >
              {chip.label}
              <button
                type="button"
                onClick={chip.onRemove}
                className="text-muted-foreground hover:text-foreground transition-colors ml-0.5"
                aria-label={`Remove filter ${chip.label}`}
              >
                <X className="w-2.5 h-2.5" />
              </button>
            </span>
          ))}
          <button
            type="button"
            onClick={clearAllFilters}
            className="text-xs text-muted-foreground hover:text-foreground hover:underline"
          >
            Clear all
          </button>
        </div>
      )}
    </div>
  )
}
