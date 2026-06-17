import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { useTheme, type Theme } from '@/hooks/useTheme'
import { Switch } from '@/components/ui/switch'
import { getMe, updateMe } from '@/features/auth/lib/me'

const THEME_OPTIONS: Record<Theme, { label: string; sub: string; swatches: string[] }> = {
  warm: {
    label: 'Parchment',
    sub: 'Warm parchment',
    swatches: ['#FAF8F4', '#E5DED6', '#2D2A26'],
  },
  lumen: {
    label: 'Lumen',
    sub: 'Bright and clean',
    swatches: ['#FFFFFF', '#F1F1F1', '#1A1A1A'],
  },
  sunless: {
    label: 'Sunless',
    sub: 'Mostly black',
    swatches: ['#121212', '#2A2A2A', '#EDEDED'],
  },
}

export default function AppSection() {
  const { theme, setTheme } = useTheme()
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  const updateMutation = useMutation({
    mutationFn: (folderIconsEnabled: boolean) => updateMe(getToken, { folder_icons_enabled: folderIconsEnabled }),
    onSuccess: (updatedMe) => {
      queryClient.setQueryData(['me'], updatedMe)
    },
    onError: () => {
      toast.error('Failed to update settings')
    },
  })

  const folderIconsEnabled = me?.folder_icons_enabled ?? true

  return (
    <div>
      <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        Theme
      </p>
      <div className="flex flex-col gap-2">
        {Object.entries(THEME_OPTIONS).map(([key, option]) => (
          <label
            key={key}
            className="flex items-center gap-3 rounded-lg border border-ring bg-accent px-3.5 py-2.5"
          >
            <input
              type="radio"
              name="theme"
              checked={theme === key}
              onChange={() => setTheme(key as Theme)}
              aria-label={`${option.label} theme`}
              className="shrink-0 accent-foreground"
            />
            <div className="flex gap-1">
              {option.swatches.map((color) => (
                <div
                  key={color}
                  className="size-3.5 rounded-sm border border-foreground/10"
                  style={{ backgroundColor: color }}
                />
              ))}
            </div>
            <div>
              <p className="text-sm font-medium text-foreground">{option.label}</p>
              <p className="text-xs text-muted-foreground">{option.sub}</p>
            </div>
          </label>
        ))}
      </div>

      <p className="mt-5 mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        Folders
      </p>
      <div className="flex items-center justify-between gap-4 border-b border-border py-3">
        <div>
          <p className="text-sm font-medium text-foreground">Folder icons</p>
          <p className="text-xs leading-relaxed text-muted-foreground">
            Show icons next to folders in the sidebar
          </p>
        </div>
        <Switch
          checked={folderIconsEnabled}
          onCheckedChange={(checked) => updateMutation.mutate(checked)}
          disabled={updateMutation.isPending}
        />
      </div>
    </div>
  )
}
