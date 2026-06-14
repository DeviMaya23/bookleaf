import { useTheme, type Theme } from '@/hooks/useTheme'

const THEME_OPTIONS: Record<Theme, { label: string; sub: string; swatches: string[] }> = {
  warm: {
    label: 'Default',
    sub: 'Warm parchment',
    swatches: ['#FAF8F4', '#EEE9E1', '#2D2A26'],
  },
}

export default function AppSection() {
  const { theme } = useTheme()
  const option = THEME_OPTIONS[theme]

  return (
    <div>
      <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
        Theme
      </p>
      <div className="flex items-center gap-3 rounded-lg border border-ring bg-accent px-3.5 py-2.5">
        <input
          type="radio"
          checked
          readOnly
          disabled
          aria-label={`${option.label} theme (selected)`}
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
      </div>
    </div>
  )
}
