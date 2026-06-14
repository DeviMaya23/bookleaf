import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { useKindeAuth } from '@kinde-oss/kinde-auth-react'
import { toast } from 'sonner'
import { Avatar, AvatarFallback, AvatarImage } from '@/components/ui/avatar'
import { Button } from '@/components/ui/button'
import { getInitials, getFullName } from '@/features/auth/lib/profile'
import { deleteMe } from '@/features/auth/lib/me'

export default function AccountSection() {
  const { getUserProfile, logout, getToken } = useKindeAuth()
  const [confirming, setConfirming] = useState(false)
  const [confirmEmail, setConfirmEmail] = useState('')

  const { data: profile } = useQuery({
    queryKey: ['profile'],
    queryFn: () => getUserProfile(),
    staleTime: Infinity,
  })

  const deleteMutation = useMutation({
    mutationFn: () => deleteMe(getToken),
    onSuccess: () => {
      logout()
    },
    onError: () => {
      toast.error('Failed to delete account')
    },
  })

  const emailMatches =
    !!profile?.email && confirmEmail.trim().toLowerCase() === profile.email.trim().toLowerCase()

  return (
    <div>
      <div className="mb-6">
        <p className="mb-2.5 text-[10px] font-semibold uppercase tracking-wider text-muted-foreground">
          Profile
        </p>
        <div className="flex items-center gap-4">
          <Avatar className="size-[52px]">
            <AvatarImage src={profile?.picture ?? undefined} />
            <AvatarFallback className="text-base">{profile ? getInitials(profile) : '…'}</AvatarFallback>
          </Avatar>
          <div className="flex flex-1 flex-col gap-1.5">
            <div className="font-medium">{profile ? getFullName(profile) : '…'}</div>
            <div className="rounded-md border bg-input px-2.5 py-1.5 text-sm text-muted-foreground">
              {profile?.email}
            </div>
          </div>
        </div>
      </div>

      <div className="border-b border-border pb-5">
        <Button variant="outline" onClick={() => logout()}>
          Log out
        </Button>
      </div>

      <div className="pt-5">
        <p className="mb-1.5 text-sm font-medium text-destructive">Delete account</p>
        <p className="mb-3.5 text-xs leading-relaxed text-muted-foreground">
          Permanently deletes your account and all associated data. This cannot be undone.
        </p>
        {!confirming ? (
          <Button variant="destructive" onClick={() => setConfirming(true)}>
            Delete account…
          </Button>
        ) : (
          <div className="flex flex-col gap-3">
            <p className="text-sm text-muted-foreground">
              Type <span className="font-medium text-foreground">{profile?.email}</span> to confirm.
            </p>
            <input
              className="w-full rounded-md border bg-background px-3 py-2 text-sm outline-none focus:ring-2 focus:ring-ring"
              placeholder="Type your account email"
              value={confirmEmail}
              onChange={(e) => setConfirmEmail(e.target.value)}
              autoFocus
            />
            <div className="flex gap-2">
              <Button variant="outline" onClick={() => { setConfirming(false); setConfirmEmail('') }}>
                Cancel
              </Button>
              <Button
                variant="destructive"
                disabled={!emailMatches || deleteMutation.isPending}
                onClick={() => deleteMutation.mutate()}
              >
                Delete account
              </Button>
            </div>
          </div>
        )}
      </div>
    </div>
  )
}
