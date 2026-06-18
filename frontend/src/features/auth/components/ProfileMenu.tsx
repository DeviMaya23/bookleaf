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
import { ChevronUp } from 'lucide-react'

export default function ProfileMenu() {
  const { getUserProfile, logout } = useKindeAuth()
  const [settingsOpen, setSettingsOpen] = useState(false)

  const { data: profile } = useQuery({
    queryKey: ['profile'],
    queryFn: () => getUserProfile(),
    staleTime: Infinity,
  })

  return (
    <>
      <DropdownMenu>
        <DropdownMenuTrigger className="w-full flex items-center gap-2 rounded-md px-2 py-1.5 text-sm hover:bg-accent text-left">
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
