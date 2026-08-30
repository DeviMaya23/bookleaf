import type { ReactNode } from 'react'
import { Link } from 'react-router-dom'

interface SimplePageLayoutProps {
  title: string
  children: ReactNode
}

export default function SimplePageLayout({ title, children }: SimplePageLayoutProps) {
  return (
    <div className="min-h-dvh bg-background text-foreground">
      <nav className="border-b border-border px-8 py-4">
        <Link to="/" className="text-xl font-bold font-title">
          Bookleaf
        </Link>
      </nav>

      <main className="mx-auto max-w-2xl px-8 py-12">
        <h1 className="mb-8 font-serif text-3xl font-semibold">{title}</h1>
        {children}
      </main>
    </div>
  )
}
