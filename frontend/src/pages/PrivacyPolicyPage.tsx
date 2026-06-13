import { Link } from 'react-router-dom'
import SimplePageLayout from '@/components/SimplePageLayout'

const LOREM =
  'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.'

export default function PrivacyPolicyPage() {
  return (
    <SimplePageLayout title="Privacy Policy">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">What we collect</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">How we use it</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Your choices</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <p className="text-muted-foreground">
        See also: <Link to="/ai-notes" className="underline">AI Notes</Link>
      </p>
    </SimplePageLayout>
  )
}
