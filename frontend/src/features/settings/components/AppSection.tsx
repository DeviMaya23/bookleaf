import { useTheme, type Theme } from '@/hooks/useTheme'

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
    </div>
  )
}
