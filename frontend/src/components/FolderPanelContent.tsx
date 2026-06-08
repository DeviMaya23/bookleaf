import { useState, useEffect, useRef } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { X } from 'lucide-react'
import { toast } from 'sonner'
import { updateFolderDetails } from '@/lib/folders'

interface FolderPanelContentProps {
  folder: { id: string; name: string; description: string | null }
  onClose: () => void
}

export default function FolderPanelContent({ folder, onClose }: FolderPanelContentProps) {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const [name, setName] = useState(folder.name)
  const [description, setDescription] = useState(folder.description ?? '')

  const origName = useRef(folder.name)
  const origDescription = useRef(folder.description ?? '')

  useEffect(() => {
    setName(folder.name)
    setDescription(folder.description ?? '')
    origName.current = folder.name
    origDescription.current = folder.description ?? ''
  }, [folder.id, folder.name, folder.description])

  const saveMutation = useMutation({
    mutationFn: (params: Parameters<typeof updateFolderDetails>[2]) =>
      updateFolderDetails(getToken, folder.id, params),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['folders'] })
      toast.success('Saved')
    },
    onError: () => {
      toast.error('Failed to save')
    },
  })

  const handleNameBlur = () => {
    if (name.trim() === '') {
      setName(origName.current)
      return
    }
    if (name !== origName.current) {
      origName.current = name
      saveMutation.mutate({ name })
    }
  }

  const handleDescriptionBlur = () => {
    if (description !== origDescription.current) {
      origDescription.current = description
      saveMutation.mutate({ description: description || null })
    }
  }

  return (
    <>
      <div className="relative flex-shrink-0 border-b px-4 pt-4 pb-3">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          onBlur={handleNameBlur}
          className="w-full text-base font-semibold bg-transparent border-b border-transparent focus:border-border focus:bg-muted/40 outline-none px-0 py-0.5 transition-colors"
        />
        <button
          onClick={onClose}
          className="absolute top-2 right-2 w-7 h-7 rounded-full bg-black/40 text-white flex items-center justify-center hover:bg-black/60 transition-colors"
          aria-label="Close panel"
        >
          <X className="w-3.5 h-3.5" />
        </button>
      </div>

      <div className="flex-1 overflow-y-auto">
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
      </div>
    </>
  )
}
