import SimplePageLayout from '@/components/SimplePageLayout'
import { EXTENSION_FIREFOX_URL, EXTENSION_CHROME_URL } from '@/lib/downloads'

export default function ExtensionsPage() {
  return (
    <SimplePageLayout title="Browser Extension">
      <section className="mb-8">
        <p className="mb-3 text-sm text-muted-foreground">
          The extension adds a right-click "Save to Bookleaf" option for easy saving of image content (PNG, JPG, GIF, WEBP, AVIF).
          To use it, install the extension for your browser and log in with your Bookleaf account.
          Right-click on any image and select "Save to Bookleaf" to save it directly to your library.
        </p>
      </section>
      <section className="mb-8">
        <div className="flex items-start gap-2.5 rounded-lg border border-border bg-muted px-3.5 py-2.5 text-xs leading-relaxed text-muted-foreground">
          <span className="mt-0.5 shrink-0 rounded bg-secondary px-1.5 py-0.5 text-[9px] font-bold uppercase tracking-wider text-muted-foreground">
            BETA
          </span>
          <span>
            The browser extension is in early access. Chrome requires manual developer-mode setup.
          </span>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Firefox</h2>
        <a href={EXTENSION_FIREFOX_URL} className="text-sm font-medium text-primary underline underline-offset-2 hover:opacity-75 transition-opacity">
          Download for Firefox (.xpi)
        </a>
        <div className="mt-4">
          <p className="mb-1 text-sm font-medium text-foreground">Manual install (if needed)</p>
          <ol className="list-decimal pl-5 text-sm text-muted-foreground space-y-1">
            <li>Open Firefox and navigate to <code className="rounded bg-muted px-1 py-px font-mono text-xs">about:addons</code></li>
            <li>Click the gear icon near the top right</li>
            <li>Select "Install Add-on From File"</li>
            <li>Choose the downloaded <code className="rounded bg-muted px-1 py-px font-mono text-xs">.xpi</code> file</li>
          </ol>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Chrome</h2>
        <p className="mb-3 text-sm text-muted-foreground">
          Chrome Web Store approval is pending. In the meantime, you can load the extension in
          developer mode.
        </p>
        <a href={EXTENSION_CHROME_URL} className="text-sm font-medium text-primary underline underline-offset-2 hover:opacity-75 transition-opacity">
          Download for Chrome (.zip)
        </a>
        <div className="mt-4">
          <p className="mb-1 text-sm font-medium text-foreground">Install steps</p>
          <ol className="list-decimal pl-5 text-sm text-muted-foreground space-y-1">
            <li>Download and unzip the file</li>
            <li>Navigate to <code className="rounded bg-muted px-1 py-px font-mono text-xs">chrome://extensions</code></li>
            <li>Enable Developer mode (toggle in the top right)</li>
            <li>Click "Load unpacked"</li>
            <li>Select the unzipped folder</li>
          </ol>
        </div>
      </section>
    </SimplePageLayout>
  )
}
