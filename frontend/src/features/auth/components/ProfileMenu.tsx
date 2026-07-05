import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import SettingsModal from '@/features/settings/components/SettingsModal'
import { getInitials, getFullName } from '@/features/auth/lib/profile'
import { getMe } from '@/features/auth/lib/me'
import { ChevronUp } from 'lucide-react'

function dismissalKey() {
  const now = new Date()
  const year = now.getUTCFullYear()
  const month = String(now.getUTCMonth() + 1).padStart(2, '0')
  return `categorisation_limit_dismissed_${year}-${month}`
}

export default function ProfileMenu() {
  const { getUserProfile, getToken, logout } = useKindeAuth()
  const [settingsOpen, setSettingsOpen] = useState(false)
  const [dismissed, setDismissed] = useState(() => {
    try {
      return localStorage.getItem(dismissalKey()) !== null
    } catch {
      return false
    }
  })

  const { data: profile } = useQuery({
    queryKey: ['profile'],
    queryFn: () => getUserProfile(),
    staleTime: Infinity,
  })

  const { data: me } = useQuery({
    queryKey: ['me'],
    queryFn: () => getMe(getToken),
    staleTime: Infinity,
  })

  const showBadge =
    !dismissed &&
    (me?.ai_categorisation_enabled ?? false) &&
    (me?.ai_categorisation_count_this_month ?? 0) >= 50

  function handleTriggerClick() {
    if (showBadge) {
      try {
        localStorage.setItem(dismissalKey(), '1')
      } catch {
        // ignore
      }
      setDismissed(true)
    }
  }

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger
          onClick={handleTriggerClick}
          className="w-full flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent text-left"
        >
          <div className="relative shrink-0">
            <Avatar className="h-7 w-7">
              <AvatarImage src={profile?.picture ?? undefined} />
              <AvatarFallback className="text-xs">
                {profile ? getInitials(profile) : '…'}
              </AvatarFallback>
            </Avatar>
            {showBadge && (
              <span
                data-testid="categorisation-limit-badge"
                className="absolute -top-0.5 -right-0.5 h-2 w-2 rounded-full bg-destructive"
              />
            )}
          </div>
          <span className="truncate font-medium text-foreground">
            {profile ? getFullName(profile) : '…'}
          </span>
          <ChevronUp className="ml-auto h-3.5 w-3.5 text-muted-foreground shrink-0" />
        </DropdownMenuTrigger>
        <DropdownMenuContent side="top" align="start" className="w-[220px]">
          {profile?.email && (
            <div className="px-2 py-1.5 text-xs text-muted-foreground truncate">{profile.email}</div>
          )}
          <DropdownMenuItem onClick={() => setSettingsOpen(true)} className="hidden sm:flex cursor-pointer">
            Settings
          </DropdownMenuItem>
          <DropdownMenuSeparator />
          <DropdownMenuItem
            onClick={() => logout()}
            className="text-destructive focus:text-destructive cursor-pointer"
          >
            Sign out
          </DropdownMenuItem>
        </DropdownMenuContent>
      </DropdownMenu>
      <SettingsModal open={settingsOpen} onOpenChange={setSettingsOpen} />
    </>
  )
}
