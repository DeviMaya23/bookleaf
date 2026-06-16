import SimplePageLayout from '@/components/SimplePageLayout'
import { EXTENSION_FIREFOX_URL, EXTENSION_CHROME_URL } from '@/lib/downloads'

export default function ExtensionsPage() {
  return (
    <SimplePageLayout title="Browser Extension">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Firefox</h2>
        <a href={EXTENSION_FIREFOX_URL} className="text-primary">
          Download for Firefox (.xpi)
        </a>
        <div className="mt-4">
          <p className="mb-1 text-sm font-medium text-foreground">Manual install (if needed)</p>
          <ol className="list-decimal pl-5 text-sm text-muted-foreground space-y-1">
            <li>Open Firefox and navigate to <code>about:addons</code></li>
            <li>Click the gear icon near the top right</li>
            <li>Select "Install Add-on From File"</li>
            <li>Choose the downloaded <code>.xpi</code> file</li>
          </ol>
        </div>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Chrome</h2>
        <p className="mb-3 text-muted-foreground">
          Chrome Web Store approval is pending. In the meantime, you can load the extension in
          developer mode.
        </p>
        <a href={EXTENSION_CHROME_URL} className="text-primary">
          Download for Chrome (.zip)
        </a>
        <div className="mt-4">
          <p className="mb-1 text-sm font-medium text-foreground">Install steps</p>
          <ol className="list-decimal pl-5 text-sm text-muted-foreground space-y-1">
            <li>Download and unzip the file</li>
            <li>Navigate to <code>chrome://extensions</code></li>
            <li>Enable Developer mode (toggle in the top right)</li>
            <li>Click "Load unpacked"</li>
            <li>Select the unzipped folder</li>
          </ol>
        </div>
      </section>
    </SimplePageLayout>
  )
}
