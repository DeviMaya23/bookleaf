import { useKindeAuth } from '@kinde-oss/kinde-auth-react'

export default function LogoutButton() {
  const { logout } = useKindeAuth()

  return (
    <button
      onClick={() => logout()}
      className="rounded-md px-3 py-1.5 text-sm font-medium text-foreground hover:bg-accent"
    >
      Sign out
    </button>
  )
}
