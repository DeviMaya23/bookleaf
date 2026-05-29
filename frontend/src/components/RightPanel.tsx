import { useState, useEffect, useRef } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Loader2, ImageIcon, Download, ExternalLink, X } from 'lucide-react'
import { toast } from 'sonner'
import { Button } from '@/components/ui/button'
import {
  Dialog,
  DialogContent,
  DialogTitle,
} from '@/components/ui/dialog'
import { getImage, updateImage, downloadImage } from '@/lib/images'
import { getFolders } from '@/lib/folders'
import { getTags, createTag } from '@/lib/tags'
import TagInput from '@/components/TagInput'
import type { Image } from '@/lib/images'
import type { Tag } from '@/lib/tags'

function formatFileSize(bytes: number | null): string {
  if (bytes === null) return '—'
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(iso: string): string {
  return new Date(iso).toLocaleDateString(undefined, {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  })
}

interface RightPanelProps {
  image: Image
  onClose: () => void
  autoFocusTitle?: boolean
}

export default function RightPanel({ image, onClose, autoFocusTitle }: RightPanelProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [lightboxOpen, setLightboxOpen] = useState(false)
  const [title, setTitle] = useState(image.title)
  const [description, setDescription] = useState(image.description ?? '')
  const [sourceUrl, setSourceUrl] = useState(image.source_url ?? '')
  const [tags, setTags] = useState<Tag[]>(image.tags ?? [])

  // Reset local state when a different image is selected
  useEffect(() => {
    setTitle(image.title)
    setDescription(image.description ?? '')
    setSourceUrl(image.source_url ?? '')
    setTags(image.tags ?? [])
    setLightboxOpen(false)
  }, [image.id])

  const { data: imageDetail, isLoading: isLoadingDetail } = useQuery({
    queryKey: ['image', image.id],
    queryFn: () => getImage(getToken, image.id),
    enabled: lightboxOpen,
    staleTime: 0,
  })

  const { data: folders } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  const { data: allTags = [] } = useQuery({
    queryKey: ['tags'],
    queryFn: () => getTags(getToken),
    staleTime: 60_000,
  })

  const folderName = image.folder_id
    ? (folders?.find((f) => f.id === image.folder_id)?.name ?? '—')
    : 'Unsorted'

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

  const handleTagsChange = async (incoming: Tag[]) => {
    // Resolve any tag that has no ID (newly typed by user)
    const resolved: Tag[] = []
    for (const t of incoming) {
      if (t.id) {
        resolved.push(t)
        continue
      }
      const existing = allTags.find(
        (a) => a.name.toLowerCase() === t.name.toLowerCase(),
      )
      if (existing) {
        resolved.push(existing)
      } else {
        try {
          const created = await createTag(getToken, t.name)
          queryClient.setQueryData<Tag[]>(['tags'], (prev = []) => [...prev, created])
          resolved.push(created)
        } catch {
          toast.error(`Failed to create tag "${t.name}"`)
          return
        }
      }
    }
    setTags(resolved)
    tagSaveMutation.mutate(resolved.map((t) => t.id))
  }

  const [isDownloading, setIsDownloading] = useState(false)

  const handleDownload = async () => {
    setIsDownloading(true)
    try {
      const url = await downloadImage(getToken, image.id)
      const a = document.createElement('a')
      a.href = url
      a.download = ''
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
    } catch {
      toast.error('Failed to download image')
    } finally {
      setIsDownloading(false)
    }
  }

  const titleInputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    if (autoFocusTitle) {
      titleInputRef.current?.focus()
    }
  }, [image.id, autoFocusTitle])

  // Refs to track original values for change detection on blur
  const origTitle = useRef(image.title)
  const origDescription = useRef(image.description ?? '')
  const origSourceUrl = useRef(image.source_url ?? '')

  useEffect(() => {
    origTitle.current = image.title
    origDescription.current = image.description ?? ''
    origSourceUrl.current = image.source_url ?? ''
  }, [image.id])

  const handleTitleBlur = () => {
    if (title.trim() === '') {
      setTitle(origTitle.current)
      return
    }
    if (title !== origTitle.current) {
      origTitle.current = title
      saveMutation.mutate({ title })
    }
  }

  const handleDescriptionBlur = () => {
    if (description !== origDescription.current) {
      origDescription.current = description
      saveMutation.mutate({ description: description || null })
    }
  }

  const handleSourceUrlBlur = () => {
    if (sourceUrl !== origSourceUrl.current) {
      origSourceUrl.current = sourceUrl
      saveMutation.mutate({ source_url: sourceUrl || null })
    }
  }

  return (
    <aside className="w-80 flex-shrink-0 border-l h-screen flex flex-col bg-background overflow-hidden">
      {/* Thumbnail */}
      <div className="relative flex-shrink-0">
        <div
          className="cursor-pointer"
          onClick={() => setLightboxOpen(true)}
        >
          {image.thumbnail_url ? (
            <img
              src={image.thumbnail_url}
              alt={image.title}
              className="w-full h-auto object-cover"
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
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            onBlur={handleTitleBlur}
            className="w-full text-base font-semibold bg-transparent border-b border-transparent focus:border-border focus:bg-muted/40 outline-none px-0 py-0.5 transition-colors"
          />
        </div>

        {/* Notes */}
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Notes</p>
          <textarea
            value={description}
            onChange={(e) => setDescription(e.target.value)}
            onBlur={handleDescriptionBlur}
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
              value={sourceUrl}
              onChange={(e) => setSourceUrl(e.target.value)}
              onBlur={handleSourceUrlBlur}
              placeholder="https://…"
              className="flex-1 text-sm bg-muted/30 border border-border/50 rounded-lg px-3 py-1.5 outline-none focus:border-border transition-colors min-w-0"
            />
            <a
              href={sourceUrl || undefined}
              target="_blank"
              rel="noreferrer"
              onClick={(e) => !sourceUrl && e.preventDefault()}
              className={`flex-shrink-0 flex items-center gap-1 px-2.5 py-1.5 rounded-lg text-xs font-medium border transition-colors ${
                sourceUrl
                  ? 'bg-foreground text-background border-foreground hover:opacity-80'
                  : 'bg-muted text-muted-foreground border-border cursor-default'
              }`}
            >
              Open <ExternalLink className="w-3 h-3" />
            </a>
          </div>
        </div>

        {/* Tags */}
        <div className="px-4 py-3 border-b">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2">Tags</p>
          <TagInput
            tags={tags}
            onChange={handleTagsChange}
            disabled={tagSaveMutation.isPending}
          />
        </div>

        {/* Details */}
        <div className="px-4 py-3">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-3">Details</p>
          <div className="grid grid-cols-2 gap-x-4 gap-y-3">
            {[
              ['Size', formatFileSize(image.file_size)],
              ['Dimensions', image.width && image.height ? `${image.width} × ${image.height}` : '—'],
              ['Folder', folderName],
              ['Added', formatDate(image.created_at)],
            ].map(([label, value]) => (
              <div key={label}>
                <p className="text-[10px] text-muted-foreground mb-0.5">{label}</p>
                <p className="text-xs font-medium truncate">{value}</p>
              </div>
            ))}
          </div>
        </div>
      </div>

      {/* Sticky footer — Download */}
      <div className="flex-shrink-0 border-t px-4 py-3 bg-background">
        <Button
          className="w-full"
          onClick={handleDownload}
          disabled={isDownloading}
        >
          {isDownloading ? (
            <>
              <Loader2 className="w-4 h-4 mr-2 animate-spin" />
              Downloading…
            </>
          ) : (
            <>
              <Download className="w-4 h-4 mr-2" />
              Download image
            </>
          )}
        </Button>
      </div>

      {/* Lightbox */}
      <Dialog open={lightboxOpen} onOpenChange={(open) => { if (!open) setLightboxOpen(false) }}>
        <DialogContent className="sm:max-w-fit p-0 overflow-hidden" showCloseButton={false}>
          <DialogTitle className="sr-only">{image.title}</DialogTitle>
          {isLoadingDetail ? (
            <div className="flex items-center justify-center w-64 h-64">
              <Loader2 className="w-8 h-8 animate-spin text-muted-foreground" />
            </div>
          ) : imageDetail ? (
            <img
              src={imageDetail.image_url}
              alt={image.title}
              className="max-h-[90vh] max-w-[90vw] object-contain"
            />
          ) : null}
        </DialogContent>
      </Dialog>
    </aside>
  )
}
