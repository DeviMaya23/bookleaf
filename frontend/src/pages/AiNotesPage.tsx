import { Link } from 'react-router-dom'
import SimplePageLayout from '@/components/SimplePageLayout'

const LOREM =
  'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.'

export default function AiNotesPage() {
  return (
    <SimplePageLayout title="AI Notes">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">How AI is used here</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">
          What gets sent to Google Vision API
        </h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Opting out</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <p className="text-muted-foreground">
        See also: <Link to="/privacy" className="underline">Privacy Policy</Link>
      </p>
    </SimplePageLayout>
  )
}
