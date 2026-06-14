import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Link, Navigate, useLocation } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import landingHero from '@/assets/landing-hero.png'

const FEATURES = [
  {
    num: '01',
    title: 'Upload and organise',
    body: 'Add images from your device, drag them into folders.',
  },
  {
    num: '02',
    title: 'Clean gallery and viewer',
    body: 'Your images in a tidy grid, with full-resolution viewing.',
  },
  {
    num: '03',
    title: 'Optional smart features',
    body: 'Off by default, always will be.',
  },
]

export default function LandingPage() {
  const { isAuthenticated, isLoading, login } = useKindeAuth()
  const location = useLocation()
  const errorMessage = (location.state as { error?: string } | null)?.error

  if (!isLoading && isAuthenticated) return <Navigate to="/app" replace />

  return (
    <div className="relative grid h-dvh grid-rows-[auto_1fr_auto_auto] overflow-hidden bg-background text-foreground">
      {errorMessage && (
        <div className="absolute inset-x-0 top-0 z-10 bg-destructive/10 px-8 py-2 text-center text-sm text-destructive">
          {errorMessage}
        </div>
      )}

      <nav className="flex items-center justify-between border-b border-border px-8 py-4">
        <span className="font-serif text-lg font-semibold">Bookleaf</span>
        <Button variant="secondary" onClick={() => login()}>
          Sign in
        </Button>
      </nav>

      <section className="grid grid-cols-2 items-center gap-12 overflow-hidden px-8 py-6">
        <div className="max-w-md">
          <p className="mb-2 text-xs font-medium uppercase tracking-widest text-muted-foreground">
            (Still in beta)
          </p>
          <h1 className="mb-4 font-serif text-4xl font-semibold leading-tight">
            Noiseless web image library
          </h1>
          <p className="mb-6 text-base text-muted-foreground">
            It's just you and your images here.
          </p>
          <Button size="lg" onClick={() => login()}>
            Get started
          </Button>
        </div>
        <div className="aspect-[4/3] overflow-hidden rounded-lg border border-border shadow-lg">
          <img
            src={landingHero}
            alt="Bookleaf app screenshot"
            className="h-full w-full object-cover"
          />
        </div>
      </section>

      <section className="grid grid-cols-3 gap-8 bg-secondary px-8 py-8">
        {FEATURES.map((feature) => (
          <div key={feature.num}>
            <span className="text-xs font-semibold text-muted-foreground">{feature.num}</span>
            <h3 className="mt-1 mb-2 font-serif text-lg font-semibold">{feature.title}</h3>
            <p className="text-sm text-muted-foreground">{feature.body}</p>
          </div>
        ))}
      </section>

      <footer className="flex items-center justify-center gap-4 border-t border-border px-8 py-4 text-sm text-muted-foreground">
        <Link to="/about">About</Link>
        <Link to="/privacy">Privacy</Link>
        <Link to="/ai-notes">AI Notes</Link>
      </footer>
    </div>
  )
}
