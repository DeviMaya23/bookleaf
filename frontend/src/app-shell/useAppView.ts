import { useParams, useLocation } from 'react-router-dom'
import type { AppView } from '@/lib/view'

// Derives the current gallery view from the route. The view/sort/filter default
// helpers (defaultSortForViewType, filterSectionsForViewType) live with
// useGalleryControls, since that hook is their only consumer.
export function useAppView(): AppView {
  const { folderId } = useParams<{ folderId: string }>()
  const { pathname } = useLocation()

  if (folderId) return { type: 'folder', id: folderId }
  if (pathname === '/app/unsorted') return { type: 'unsorted' }
  if (pathname === '/app/trash') return { type: 'trash' }
  return { type: 'all' }
}
