import { useState } from 'react'
import { ChevronRightIcon, DownloadIcon } from 'lucide-react'
import { cn } from '@/lib/utils'
import { EXTENSION_FIREFOX_URL, EXTENSION_CHROME_URL } from '@/lib/downloads'

function ChromeIcon() {
  return (
    <svg
      width="28"
      height="28"
      viewBox="0 0 24 24"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      {/* Red sector — top, from 10 o'clock to 2 o'clock through 12 */}
      <path d="M12 12L3.34 7A10 10 0 0 1 20.66 7Z" fill="#EA4335" />
      {/* Green sector — bottom-right, from 2 o'clock to 6 o'clock */}
      <path d="M12 12L20.66 7A10 10 0 0 1 12 22Z" fill="#34A853" />
      {/* Yellow sector — bottom-left, from 6 o'clock to 10 o'clock */}
      <path d="M12 12L12 22A10 10 0 0 1 3.34 7Z" fill="#FBBC05" />
      <circle cx="12" cy="12" r="7" fill="white" />
      <circle cx="12" cy="12" r="5" fill="#4285F4" />
    </svg>
  )
}

function FirefoxIcon() {
  return (
    <svg
      width="28"
      height="28"
      viewBox="0 0 28 28"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <circle cx="14" cy="14" r="14" fill="#FF6611" fillOpacity="0.12" />
      <circle cx="14" cy="14" r="9" fill="#FF6611" fillOpacity="0.18" />
      <text
        x="14"
        y="18.5"
        textAnchor="middle"
        fontSize="13"
        fill="#E55A00"
        fontWeight="700"
        fontFamily="sans-serif"
      >
        ff
      </text>
    </svg>
  )
}

interface DownloadCardProps {
  icon: React.ReactNode
  label: string
  sub: string
  href: string
}

function DownloadCard({ icon, label, sub, href }: DownloadCardProps) {
  return (
    <a
      href={href}
      className="flex items-center gap-3 rounded-lg border border-border px-3.5 py-2.5 text-inherit no-underline transition-colors hover:border-ring hover:bg-accent"
    >
      <div className="shrink-0">{icon}</div>
      <div className="min-w-0 flex-1">
        <p className="text-sm font-medium text-foreground">{label}</p>
        <p className="mt-0.5 text-xs text-muted-foreground">{sub}</p>
      </div>
      <DownloadIcon className="size-3.5 shrink-0 text-muted-foreground" />
    </a>
  )
}

export default function ExtensionsSection() {
  const [ffManualOpen, setFfManualOpen] = useState(false)

  return (
    <div className="flex flex-col gap-6">
      {/* Beta notice */}
      <div className="flex items-start gap-2.5 rounded-lg border border-border bg-muted px-3.5 py-2.5 text-xs leading-relaxed text-muted-foreground">
        <span className="mt-0.5 shrink-0 rounded bg-secondary px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
          BETA
        </span>
        <span>
          The browser extension is in early access. Chrome requires manual developer-mode setup.
        </span>
      </div>

      {/* Firefox */}
      <div>
        <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
          Firefox
        </p>
        <DownloadCard
          href={EXTENSION_FIREFOX_URL}
          label="Download for Firefox"
          sub=".xpi"
          icon={<FirefoxIcon />}
        />
        <button
          type="button"
          data-testid="ff-manual-toggle"
          onClick={() => setFfManualOpen((o) => !o)}
          className="mt-2.5 inline-flex items-center gap-1.5 text-xs text-muted-foreground transition-colors hover:text-foreground"
        >
          <ChevronRightIcon
            className={cn('size-3 transition-transform', ffManualOpen && 'rotate-90')}
          />
          Manual install instructions
        </button>
        {ffManualOpen && (
          <ol
            data-testid="ff-manual-steps"
            className="mt-2.5 flex list-decimal flex-col gap-1.5 pl-4 text-xs leading-relaxed text-muted-foreground"
          >
            <li>
              Open <code className="rounded bg-muted px-1 py-px font-mono text-[10px]">about:addons</code> in Firefox
            </li>
            <li>Click the gear icon → "Install Add-on From File"</li>
            <li>Choose the downloaded .xpi file</li>
          </ol>
        )}
      </div>

      <div className="border-t border-border" />

      {/* Chrome */}
      <div>
        <div className="mb-3 flex items-center gap-2">
          <p className="text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
            Chrome
          </p>
          <span className="rounded bg-secondary px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
            DEVELOPER MODE
          </span>
        </div>
        <DownloadCard
          href={EXTENSION_CHROME_URL}
          label="Download for Chrome"
          sub=".zip · Requires manual setup below"
          icon={<ChromeIcon />}
        />
        <ol className="mt-2.5 flex list-decimal flex-col gap-1.5 pl-4 text-xs leading-relaxed text-muted-foreground">
          <li>Unzip the downloaded file</li>
          <li>
            Go to <code className="rounded bg-muted px-1 py-px font-mono text-[10px]">chrome://extensions</code>
          </li>
          <li>
            Enable <strong className="font-medium text-foreground">Developer mode</strong>{' '}
            (top-right toggle)
          </li>
          <li>
            Click <strong className="font-medium text-foreground">Load unpacked</strong> and select
            the unzipped folder
          </li>
        </ol>
      </div>
    </div>
  )
}
