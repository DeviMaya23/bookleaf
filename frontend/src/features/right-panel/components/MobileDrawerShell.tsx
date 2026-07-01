import type { ReactNode } from 'react'

interface MobileDrawerShellProps {
  onClose: () => void
  children: ReactNode
}

export default function MobileDrawerShell({ onClose, children }: MobileDrawerShellProps) {
  return (
    <>
      <div
        data-testid="mobile-drawer-shell-backdrop"
        className="fixed inset-0 z-40 bg-black/35 transition-opacity duration-200"
        onClick={onClose}
      />
      <div className="fixed inset-x-0 bottom-0 z-50 max-h-[80vh] flex flex-col bg-background border-t rounded-t-xl overflow-hidden transform transition-transform duration-200 translate-y-0">
        {children}
      </div>
    </>
  )
}
