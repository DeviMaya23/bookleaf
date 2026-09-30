import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { useQuery } from '@tanstack/react-query'
import { Trash2, FolderPlus } from 'lucide-react'
import { getFolders } from '@/lib/folders'
import SelectionFolderPicker from './SelectionFolderPicker'

interface SelectionPanelBodyProps {
  selectedCount: number
  onAddToFolder: (folderId: string) => void
  onMoveToTrash: () => void
}

export default function SelectionPanelBody({ selectedCount, onAddToFolder, onMoveToTrash }: SelectionPanelBodyProps) {
  const { getToken } = useKindeAuth()

  const { data: folders = [] } = useQuery({
    queryKey: ['folders'],
    queryFn: () => getFolders(getToken),
    staleTime: 60_000,
  })

  return (
    <>
      <div className="flex-shrink-0 border-b px-4 pt-4 pb-3">
        <p className="text-base font-semibold">{selectedCount} selected</p>
      </div>

      <div className="flex-1 overflow-y-auto px-4 py-3 space-y-4">
        <div>
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground mb-2 flex items-center gap-1.5">
            <FolderPlus className="w-3.5 h-3.5" />
            Add to folder
          </p>
          <SelectionFolderPicker folders={folders} onPick={onAddToFolder} />
        </div>

        <button
          type="button"
          onClick={onMoveToTrash}
          className="w-full flex items-center justify-center gap-1.5 rounded-lg border border-destructive/40 text-destructive px-3 py-2 text-sm font-medium hover:bg-destructive/10 transition-colors"
        >
          <Trash2 className="w-3.5 h-3.5" />
          Move to trash
        </button>
      </div>
    </>
  )
}
