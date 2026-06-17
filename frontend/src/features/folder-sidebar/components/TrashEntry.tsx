import { useState } from 'react'
import { toast } from 'sonner'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import {
  ContextMenu,
  ContextMenuContent,
  ContextMenuItem,
  ContextMenuTrigger,
} from '@/components/ui/context-menu'
import { emptyTrash } from '@/lib/images'
import EmptyTrashDialog from './EmptyTrashDialog'
import { FOLDER_ICONS, SYSTEM_ICON_KEYS } from '../lib/folderIcons'

interface TrashEntryProps {
  active: boolean
  iconsEnabled: boolean
  onClick: () => void
}

export default function TrashEntry({ active, iconsEnabled, onClick }: TrashEntryProps) {
  const Icon = FOLDER_ICONS[SYSTEM_ICON_KEYS.trash]
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [open, setOpen] = useState(false)

  const emptyTrashMutation = useMutation({
    mutationFn: () => emptyTrash(getToken),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['images', 'trash'] })
      toast.success('Trash emptied')
    },
    onError: () => toast.error('Failed to empty trash'),
  })

  return (
    <div className="mt-[8px]">
      <ContextMenu>
        <ContextMenuTrigger>
          <div
            className={`flex items-center gap-1.5 px-3 py-1 rounded-md cursor-pointer text-sm select-none ${
              active
                ? 'bg-accent text-accent-foreground font-medium'
                : 'text-muted-foreground/60 hover:bg-accent hover:text-accent-foreground'
            }`}
            onClick={onClick}
          >
            {iconsEnabled && <Icon className="w-3.5 h-3.5 shrink-0" />}
            Trash
          </div>
        </ContextMenuTrigger>
        <ContextMenuContent>
          <ContextMenuItem
            onClick={() => setOpen(true)}
            className="text-destructive focus:text-destructive"
          >
            Empty trash
          </ContextMenuItem>
        </ContextMenuContent>
      </ContextMenu>

      <EmptyTrashDialog
        open={open}
        onCancel={() => setOpen(false)}
        onConfirm={() => {
          emptyTrashMutation.mutate()
          setOpen(false)
        }}
      />
    </div>
  )
}
