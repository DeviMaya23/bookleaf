import { Link } from 'react-router-dom'
import SimplePageLayout from '@/components/SimplePageLayout'

export default function AiNotesPage() {
  return (
    <SimplePageLayout title="AI Notes">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">AI in Bookleaf's smart features</h2>
        <p className="text-muted-foreground">Currently, Bookleaf uses the combination of Google's Vision API and Anthropic's AI model to provide 'smart' searches and/or automatically categorise when you upload an image.
        </p>
        <br></br>
        <p className="text-muted-foreground">This feature is off by default and can be enabled in Settings → Advanced.</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">
          What gets sent to Google Vision API
        </h2>
        <p className="text-muted-foreground">When you enable the smart features, a thumbnail version (same aspect ratio, max 600px) of your uploaded image will be sent to Google's Vision API for analysis.</p>
        <br></br>
        <p className="text-muted-foreground">The application will backfill any images missing Vision labelling.</p>
      </section>


      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">
          What gets sent to Anthropic's AI model
        </h2>
        <p className="text-muted-foreground">When you enable the AI categorisation feature, the results from Google's Vision API along with your folder list and metadata (folder name, folder description) will be sent to Anthropic's AI model for analysis.</p>
        <br></br>
        <p className="text-muted-foreground">To refine folder suggestions, the model may also request additional information, such as metadata of images inside of a folder. The actual images and thumbnails are not used in this process.</p>
      </section>

      <p className="text-muted-foreground">
        See also: <Link to="/privacy" className="underline">Privacy Policy</Link>
      </p>
    </SimplePageLayout>
  )
}
