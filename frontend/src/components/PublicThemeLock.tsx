import { Outlet } from 'react-router-dom'

export default function PublicThemeLock() {
  return (
    <div data-theme="warm">
      <Outlet />
    </div>
  )
}
