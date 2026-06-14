import { Link } from 'react-router-dom'
import SimplePageLayout from '@/components/SimplePageLayout'

export default function AiNotesPage() {
  return (
    <SimplePageLayout title="AI Notes">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">How AI is used here</h2>
        <p className="text-muted-foreground">Currently, Bookleaf uses Google's Vision API to provide folder suggestions when you upload an image.
          The API returns labels describing the content of the image, which Bookleaf uses to suggest relevant folders for organisation. The labels are stored in our database.
        </p>
        <br></br>
        <p className="text-muted-foreground">This feature is off by default and can be enabled in Settings → Advanced.</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">
          What gets sent to Google Vision API
        </h2>
        <p className="text-muted-foreground">When you enable the AI folder suggestions feature, a thumbnail version (same aspect ratio, max 600px) of your uploaded image will be sent to Google's Vision API for analysis.</p>
      </section>

      <p className="text-muted-foreground">
        See also: <Link to="/privacy" className="underline">Privacy Policy</Link>
      </p>
    </SimplePageLayout>
  )
}
