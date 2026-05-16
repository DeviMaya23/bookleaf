import { useEffect, useState } from 'react'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import type { UserProfile } from '@kinde/js-utils'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { ChevronUp } from 'lucide-react'

function getInitials(profile: UserProfile): string {
  const first = profile.givenName?.[0] ?? ''
  const last = profile.familyName?.[0] ?? ''
  return (first + last).toUpperCase() || '?'
}

function getFullName(profile: UserProfile): string {
  return [profile.givenName, profile.familyName].filter(Boolean).join(' ')
}

export default function ProfileMenu() {
  const { getUserProfile, logout } = useKindeAuth()
  const [profile, setProfile] = useState<UserProfile | null>(null)

  useEffect(() => {
    getUserProfile().then(setProfile)
  }, [getUserProfile])

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <button className="w-full flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent text-left">
          <Avatar className="h-7 w-7 shrink-0">
            <AvatarImage src={profile?.picture ?? undefined} />
            <AvatarFallback className="text-xs">
              {profile ? getInitials(profile) : '…'}
            </AvatarFallback>
          </Avatar>
          <span className="truncate font-medium text-foreground">
            {profile ? getFullName(profile) : '…'}
          </span>
          <ChevronUp className="ml-auto h-3.5 w-3.5 text-muted-foreground shrink-0" />
        </button>
      </DropdownMenuTrigger>
      <DropdownMenuContent side="top" align="start" className="w-[220px]">
        {profile?.email && (
          <div className="px-2 py-1.5 text-xs text-muted-foreground truncate">{profile.email}</div>
        )}
        <DropdownMenuItem
          onClick={() => logout()}
          className="text-destructive focus:text-destructive cursor-pointer"
        >
          Sign out
        </DropdownMenuItem>
      </DropdownMenuContent>
    </DropdownMenu>
  )
}
