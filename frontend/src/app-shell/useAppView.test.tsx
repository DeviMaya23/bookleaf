import { describe, it, expect } from 'vitest'
import { renderHook } from '@testing-library/react'
import { MemoryRouter, Routes, Route } from 'react-router-dom'
import type { AppView } from '@/lib/view'
import { useAppView } from './useAppView'

function viewAt(path: string): AppView {
  const wrapper = ({ children }: { children: React.ReactNode }) => (
    <MemoryRouter initialEntries={[path]}>
      <Routes>
        <Route path="/folders/:folderId" element={children} />
        <Route path="*" element={children} />
      </Routes>
    </MemoryRouter>
  )
  return renderHook(() => useAppView(), { wrapper }).result.current
}

describe('useAppView', () => {
  it('maps routes to the matching view', () => {
    expect(viewAt('/')).toEqual({ type: 'all' })
    expect(viewAt('/unsorted')).toEqual({ type: 'unsorted' })
    expect(viewAt('/trash')).toEqual({ type: 'trash' })
    expect(viewAt('/folders/folder-1')).toEqual({ type: 'folder', id: 'folder-1' })
  })
})
