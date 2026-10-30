import React from 'react'

export const contextMenuMock = {
  ContextMenu: ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children),
  ContextMenuTrigger: ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children),
  ContextMenuContent: ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children),
  ContextMenuSeparator: () => React.createElement('hr'),
  ContextMenuItem: ({
    children,
    onClick,
    className,
  }: {
    children: React.ReactNode
    onClick?: (e: React.MouseEvent) => void
    className?: string
  }) => React.createElement('button', { role: 'menuitem', className, onClick }, children),
  ContextMenuSub: ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children),
  ContextMenuSubTrigger: ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children),
  ContextMenuSubContent: ({ children }: { children: React.ReactNode }) =>
    React.createElement(React.Fragment, null, children),
}
