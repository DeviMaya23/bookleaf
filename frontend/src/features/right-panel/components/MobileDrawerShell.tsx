import type { ReactNode } from 'react'
import { Drawer } from '@base-ui/react/drawer'

interface MobileDrawerShellProps {
  open: boolean
  onClose: () => void
  children: ReactNode
}

export default function MobileDrawerShell({ open, onClose, children }: MobileDrawerShellProps) {
  return (
    <Drawer.Root open={open} onOpenChange={(isOpen) => { if (!isOpen) onClose() }}>
      <Drawer.Portal>
        <Drawer.Backdrop
          data-testid="mobile-drawer-shell-backdrop"
          className="fixed inset-0 z-40 bg-black/35 transition-opacity duration-200"
        />
        <Drawer.Popup className="fixed inset-x-0 bottom-0 z-50 max-h-[80vh] flex flex-col bg-background border-t rounded-t-xl overflow-hidden outline-none">
          <Drawer.Title className="sr-only">Details panel</Drawer.Title>
          {children}
        </Drawer.Popup>
      </Drawer.Portal>
    </Drawer.Root>
  )
}
