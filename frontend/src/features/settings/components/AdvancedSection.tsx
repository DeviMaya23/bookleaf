import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { Switch } from '@/components/ui/switch'
import { getMe, updateMe } from '@/features/auth/lib/me'

export default function AdvancedSection() {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [tooltipOpen, setTooltipOpen] = useState(false)

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  const updateMutation = useMutation({
    mutationFn: (visionEnabled: boolean) => updateMe(getToken, { vision_enabled: visionEnabled }),
    onSuccess: (updatedMe) => {
      queryClient.setQueryData(['me'], updatedMe)
    },
    onError: () => {
      toast.error('Failed to update settings')
    },
  })

  const visionEnabled = me?.vision_enabled ?? false

  return (
    <div>
      <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        AI
      </p>
      <div className="flex items-start justify-between gap-4 border-b border-border py-3">
        <div>
          <div className="mb-1 flex items-center gap-1.5">
            <span className="text-sm font-medium text-foreground">AI folder suggestions</span>
            <div className="relative inline-flex">
              <button
                type="button"
                onMouseEnter={() => setTooltipOpen(true)}
                onMouseLeave={() => setTooltipOpen(false)}
                className="flex size-[15px] items-center justify-center rounded-full border border-foreground/20 bg-muted text-[9px] font-bold leading-none text-muted-foreground"
                aria-label="What does this do?"
              >
                ?
              </button>
              {tooltipOpen && (
                <div
                  role="tooltip"
                  className="pointer-events-none absolute top-full left-0 z-50 mt-2 w-[230px] rounded-md bg-foreground px-3 py-2 text-xs leading-relaxed text-primary-foreground shadow-lg"
                >
                  Enables folder suggestions on single-file upload, using Google's Vision API.
                </div>
              )}
            </div>
          </div>
          <p className="text-xs leading-relaxed text-muted-foreground" data-testid="vision-description">
            {visionEnabled ? 'Gets folder suggestions on single-file upload' : 'Folder suggestions off'}
          </p>
        </div>
        <Switch
          checked={visionEnabled}
          onCheckedChange={(checked) => updateMutation.mutate(checked)}
          disabled={updateMutation.isPending}
        />
      </div>
    </div>
  )
}
