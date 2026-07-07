import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { Switch } from '@/components/ui/switch'
import { getMe, updateMe, backfillVisionLabels } from '@/features/auth/lib/me'
import VisionBackfillConfirmDialog from './VisionBackfillConfirmDialog'

export default function AdvancedSection() {
  const { getToken } = useKindeAuth()
  const queryClient = useQueryClient()
  const [tooltipOpen, setTooltipOpen] = useState(false)
  const [categorisationTooltipOpen, setCategorisationTooltipOpen] = useState(false)
  const [confirmOpen, setConfirmOpen] = useState(false)

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  const enableWithBackfillMutation = useMutation({
    mutationFn: async () => {
      const updatedMe = await updateMe(getToken, { vision_enabled: true })
      try {
        await backfillVisionLabels(getToken)
      } catch {
        return { updatedMe, backfillFailed: true }
      }
      return { updatedMe, backfillFailed: false }
    },
    onSuccess: ({ updatedMe, backfillFailed }) => {
      setConfirmOpen(false)
      queryClient.setQueryData(['me'], updatedMe)
      if (backfillFailed) {
        toast.warning('Smart Features enabled, but existing images could not be queued for processing. Try again later.')
      }
    },
    onError: () => {
      setConfirmOpen(false)
      toast.error('Failed to update settings')
    },
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

  const updateCategorisationMutation = useMutation({
    mutationFn: (aiCategorisationEnabled: boolean) =>
      updateMe(getToken, { ai_categorisation_enabled: aiCategorisationEnabled }),
    onSuccess: (updatedMe) => {
      queryClient.setQueryData(['me'], updatedMe)
    },
    onError: () => {
      toast.error('Failed to update settings')
    },
  })

  const visionEnabled = me?.vision_enabled ?? false
  const aiCategorisationEnabled = me?.ai_categorisation_enabled ?? false
  const categorisationCount = me?.ai_categorisation_count_this_month ?? 0

  return (
    <div>
      <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        Smart Features
      </p>
      <div className="flex items-start justify-between gap-4 py-3">
        <div>
          <div className="mb-1 flex items-center gap-1.5">
            <span className="text-sm font-medium text-foreground">Smart Features</span>
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
                  Uses Google's Vision API to label your images, enabling smart search by image content.
                </div>
              )}
            </div>
          </div>
          <p className="text-xs leading-relaxed text-muted-foreground" data-testid="vision-description">
            {visionEnabled ? 'Smart search available in filter option' : 'Smart features off'}
          </p>
        </div>
        <Switch
          data-testid="vision-switch"
          checked={visionEnabled}
          onCheckedChange={(checked) => {
            if (checked) {
              setConfirmOpen(true)
            } else {
              updateMutation.mutate(false)
            }
          }}
          disabled={updateMutation.isPending || enableWithBackfillMutation.isPending}
        />
      </div>
      <div className={`ml-3 flex items-start justify-between gap-4 border-l border-border pl-3 pb-3 transition-opacity${!visionEnabled ? ' opacity-50' : ''}`}>
        <div>
          <div className="mb-1 flex items-center gap-1.5">
            <span className="text-sm font-medium text-foreground">AI auto-categorisation</span>
            <div className="relative inline-flex">
              <button
                type="button"
                onMouseEnter={() => setCategorisationTooltipOpen(true)}
                onMouseLeave={() => setCategorisationTooltipOpen(false)}
                className="flex size-[15px] items-center justify-center rounded-full border border-foreground/20 bg-muted text-[9px] font-bold leading-none text-muted-foreground"
                aria-label="What does AI auto-categorisation do?"
              >
                ?
              </button>
              {categorisationTooltipOpen && (
                <div
                  role="tooltip"
                  className="pointer-events-none absolute top-full left-0 z-50 mt-2 w-[230px] rounded-md bg-foreground px-3 py-2 text-xs leading-relaxed text-primary-foreground shadow-lg"
                >
                  {visionEnabled
                    ? "Automatically categorises newly uploaded images using Anthropic's AI model with vision labels and folder metadata."
                    : 'Requires Smart Features to be enabled.'}
                </div>
              )}
            </div>
          </div>
          <p
            className="text-xs leading-relaxed text-muted-foreground"
            data-testid="categorisation-counter"
          >
            {categorisationCount} / 50 this month
          </p>
        </div>
        <Switch
          data-testid="ai-categorisation-switch"
          checked={aiCategorisationEnabled}
          onCheckedChange={(checked) => updateCategorisationMutation.mutate(checked)}
          disabled={updateCategorisationMutation.isPending || !visionEnabled}
        />
      </div>
      <VisionBackfillConfirmDialog
        open={confirmOpen}
        onCancel={() => setConfirmOpen(false)}
        onConfirm={() => enableWithBackfillMutation.mutate()}
        isPending={enableWithBackfillMutation.isPending}
      />
    </div>
  )
}
