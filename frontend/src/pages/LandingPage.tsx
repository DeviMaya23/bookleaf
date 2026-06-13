import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Navigate, useLocation } from 'react-router-dom'
import { Button } from '@/components/ui/button'
import landingHero from '@/assets/landing-hero.png'

const FEATURES = [
  {
    num: '01',
    title: 'Organize your way',
    body: 'Sort images into nested folders and tag them so you can find what you need without digging through a messy camera roll.',
  },
  {
    num: '02',
    title: 'Built for browsing',
    body: 'A fast, focused gallery with zoom, pan, and rotation built in — no waiting around for thumbnails to load.',
  },
  {
    num: '03',
    title: 'Your library, your way',
    body: 'Bookleaf keeps your photos private and organized exactly how you want, with no clutter or algorithmic feeds.',
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
            Photo library, reimagined
          </p>
          <h1 className="mb-4 font-serif text-4xl font-semibold leading-tight">
            A calmer home for your photos
          </h1>
          <p className="mb-6 text-base text-muted-foreground">
            Bookleaf is a focused space to organize, browse, and revisit your images — without
            the noise of a social feed.
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

      <section className="grid grid-cols-3 gap-8 px-8 py-8" style={{ backgroundColor: '#F0EBE3' }}>
        {FEATURES.map((feature) => (
          <div key={feature.num}>
            <span className="text-xs font-semibold text-muted-foreground">{feature.num}</span>
            <h3 className="mt-1 mb-2 font-serif text-lg font-semibold">{feature.title}</h3>
            <p className="text-sm text-muted-foreground">{feature.body}</p>
          </div>
        ))}
      </section>

      <footer className="flex items-center justify-center border-t border-border px-8 py-4">
        <span className="font-serif text-sm font-semibold">Bookleaf</span>
      </footer>
    </div>
  )
}
