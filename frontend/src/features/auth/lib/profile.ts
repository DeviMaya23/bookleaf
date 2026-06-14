import type { UserProfile } from '@kinde/js-utils'

export function getInitials(profile: UserProfile): string {
  const first = profile.givenName?.[0] ?? ''
  const last = profile.familyName?.[0] ?? ''
  return (first + last).toUpperCase() || '?'
}

export function getFullName(profile: UserProfile): string {
  return [profile.givenName, profile.familyName].filter(Boolean).join(' ')
}
