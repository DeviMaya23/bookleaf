import { EXTENSION_FIREFOX_URL, EXTENSION_CHROME_URL } from '@/lib/downloads'

export default function ExtensionsSection() {
  return (
    <div className="space-y-8">
      <div>
        <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
          Firefox
        </p>
        <a href={EXTENSION_FIREFOX_URL} className="text-sm text-primary">
          Download for Firefox (.xpi)
        </a>
        <div className="mt-3">
          <p className="mb-1 text-xs font-medium text-foreground">Manual install (if needed)</p>
          <ol className="list-decimal pl-4 text-xs text-muted-foreground space-y-1">
            <li>
              Open <code>about:addons</code> in Firefox
            </li>
            <li>Click the gear icon → "Install Add-on From File"</li>
            <li>Choose the downloaded .xpi file</li>
          </ol>
        </div>
      </div>

      <div>
        <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
          Chrome (Developer mode only)
        </p>
        <p className="mb-2 text-sm text-muted-foreground">
          Currently Bookleaf's Chrome extension is only available as a developer-mode extension.
        </p>
        <a href={EXTENSION_CHROME_URL} className="text-sm text-primary">
          Download for Chrome (.zip)
        </a>
        <div className="mt-3">
          <ol className="list-decimal pl-4 text-xs text-muted-foreground space-y-1">
            <li>Download and unzip the file</li>
            <li>
              Go to <code>chrome://extensions</code>
            </li>
            <li>Enable Developer mode</li>
            <li>Click "Load unpacked" and select the unzipped folder</li>
          </ol>
        </div>
      </div>
    </div>
  )
}
