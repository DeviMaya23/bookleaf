import { useState, useEffect, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { ImageIcon, ExternalLink, X } from 'lucide-react'
import { toast } from 'sonner'
import { updateImage } from '@/lib/images'
import { resolveOrCreateTags } from '@/lib/tags'
import type { Image } from '@/lib/images'
import type { Tag } from '@/lib/tags'
import FolderInput from './FolderInput'
import TagInput from './TagInput'
import FolderPanelContent from './FolderPanelContent'
import DetailsGrid from './DetailsGrid'
import DownloadButton from './DownloadButton'
import { useFieldAutosave } from '../hooks/useFieldAutosave'
import { useImageDetailsData } from '../hooks/useImageDetailsData'

type RightPanelProps =
  | { mode: 'image'; image: Image; onClose: () => void; autoFocusTitle?: boolean }
  | { mode: 'folder'; folder: { id: string; name: string; description: string | null }; onClose: () => void }

export default function RightPanel(props: RightPanelProps) {
  return (
    <aside className="w-80 flex-shrink-0 border-l h-screen flex flex-col bg-background overflow-hidden">
      {props.mode === 'folder' ? (
        <FolderPanelContent key={props.folder.id} folder={props.folder} onClose={props.onClose} />
      ) : (
        <ImagePanelBody key={props.image.id} image={props.image} onClose={props.onClose} autoFocusTitle={props.autoFocusTitle} />
      )}
    </aside>
  )
}

interface ImagePanelBodyProps {
  image: Image
  onClose: () => void
  autoFocusTitle?: boolean
}

function ImagePanelBody({ image, onClose, autoFocusTitle }: ImagePanelBodyProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const { imageDetail, allFolders, allTags, selectedFolders, setFolders, isFoldersSaving } =
    useImageDetailsData(image)

  const [tags, setTags] = useState<Tag[]>(image.tags ?? [])

  const saveMutation = useMutation({
    mutationFn: (params: Parameters<typeof updateImage>[2]) =>
      updateImage(getToken, image.id, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      toast.success('Saved')
    },
    onError: () => {
      toast.error('Failed to save')
    },
  })

  const tagSaveMutation = useMutation({
    mutationFn: (tagIds: string[]) =>
      updateImage(getToken, image.id, { tags: tagIds }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images'] })
      toast.success('Saved')
    },
    onError: () => {
      toast.error('Failed to save tags')
    },
  })

  const titleField = useFieldAutosave(
    image.title,
    (value) => saveMutation.mutate({ title: value }),
    { isEmpty: (value) => value.trim() === '' },
  )
  const descriptionField = useFieldAutosave(
    image.description ?? '',
    (value) => saveMutation.mutate({ description: value || null }),
  )
  const sourceUrlField = useFieldAutosave(
    image.source_url ?? '',
    (value) => saveMutation.mutate({ source_url: value || null }),
  )

  const handleTagsChange = async (incoming: Tag[]) => {
    let resolved: Tag[]
    try {
      resolved = await resolveOrCreateTags(getToken, incoming, allTags, queryClient)
    } catch (err) {
      toast.error(err instanceof Error ? err.message : 'Failed to resolve tags')
      return
    }
    setTags(resolved)
    tagSaveMutation.mutate(resolved.map((t) => t.id))
  }

  const titleInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (autoFocusTitle) {
      titleInputRef.current?.focus()
    }
  }, [image.id, autoFocusTitle])

  const thumbnailUrl = imageDetail?.thumbnail_url ?? image.thumbnail_url

  return (
    <>
      {/* Thumbnail */}
      <div className="relative flex-shrink-0 max-h-[33vh]">
        <div>
          {thumbnailUrl ? (
            <img
              src={thumbnailUrl}
              alt={image.title}
              className="max-h-[33vh] w-auto mx-auto block"
            />
          ) : (
            <div className="w-full aspect-video flex items-center justify-center bg-muted">
              <ImageIcon className="w-10 h-10 text-muted-foreground" />
            </div>
          )}
        </div>
        <button
          onClick={onClose}
          className="absolute top-2 right-2 w-7 h-7 rounded-full bg-black/40 text-white flex items-center justify-center hover:bg-black/60 transition-colors"
          aria-label="Close panel"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      {/* Scrollable body */}
      <div className="flex-1 overflow-y-auto">
        {/* Title */}
        <div className="px-4 pt-4 pb-3 border-b">
          <input
            ref={titleInputRef}
            value={titleField.value}
            onChange={(e) => titleField.onChange(e.target.value)}
            onBlur={titleField.onBlur}
            className="w-full text-base font-semibold bg-transparent border-b border-transparent focus:border-border focus:bg-muted/40 outline-none px-0 py-0.5 transition-colors"
          />
        </div>

        {/* Notes */}
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Notes</p>
          <textarea
            value={descriptionField.value}
            onChange={(e) => descriptionField.onChange(e.target.value)}
            onBlur={descriptionField.onBlur}
            placeholder="Add a note…"
            rows={3}
            className="w-full resize-none text-sm bg-muted/30 border border-border/50 rounded-lg px-3 py-2 outline-none focus:border-border transition-colors"
          />
        </div>

        {/* Source URL */}
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Source</p>
          <div className="flex gap-2 items-center">
            <input
              value={sourceUrlField.value}
              onChange={(e) => sourceUrlField.onChange(e.target.value)}
              onBlur={sourceUrlField.onBlur}
              placeholder="https://…"
              className="flex-1 text-sm bg-muted/30 border border-border/50 rounded-lg px-3 py-1.5 outline-none focus:border-border transition-colors min-w-0"
            />
            <a
              href={sourceUrlField.value || undefined}
              target="_blank"
              rel="noreferrer"
              onClick={(e) => !sourceUrlField.value && e.preventDefault()}
              className={`flex-shrink-0 flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
                sourceUrlField.value
                  ? 'bg-foreground text-background border-foreground hover:opacity-80'
                  : 'bg-muted text-muted-foreground border-border cursor-default'
              }`}
            >
              Open <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>

        {/* Folders */}
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Folders</p>
          <FolderInput
            folders={selectedFolders}
            onChange={setFolders}
            disabled={isFoldersSaving}
            suggestions={(allFolders ?? []).filter((f) => !selectedFolders.some((s) => s.id === f.id))}
          />
        </div>

        {/* Tags */}
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Tags</p>
          <TagInput
            tags={tags}
            onChange={handleTagsChange}
            disabled={tagSaveMutation.isPending}
            suggestions={allTags.filter((t) => !tags.some((applied) => applied.id === t.id))}
          />
        </div>

        {/* Details */}
        <DetailsGrid image={image} />
      </div>

      {/* Sticky footer — Download */}
      <DownloadButton imageId={image.id} />
    </>
  )
}
