import SimplePageLayout from '@/components/SimplePageLayout'

const LOREM =
  'Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat.'

export default function AboutPage() {
  return (
    <SimplePageLayout title="About Bookleaf">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">What Bookleaf is</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Why it exists</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Who's behind it</h2>
        <p className="text-muted-foreground">{LOREM}</p>
      </section>
    </SimplePageLayout>
  )
}
