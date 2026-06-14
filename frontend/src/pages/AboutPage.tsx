import SimplePageLayout from '@/components/SimplePageLayout'

const WHY =
  "I want a noiseless online image gallery with specific features (some still in development). So here we are."

const ABOUT = 
  'Bookleaf is a web-based personal image library/moodboarding tool. The idea is a low-noise space for browsing through your images.'

export default function AboutPage() {
  return (
    <SimplePageLayout title="About Bookleaf">
      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">What is this?</h2>
        <p className="text-muted-foreground">{ABOUT}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Why it exists</h2>
        <p className="text-muted-foreground">{WHY}</p>
      </section>

      <section className="mb-8">
        <h2 className="mb-2 font-serif text-xl font-semibold">Contact</h2>
        <p className="text-muted-foreground">Email me:</p>
        <a href="mailto:bookleaf@evimay.me" className="text-primary">
          bookleaf@evimay.me
        </a>
      </section>
    </SimplePageLayout>
  )
}
