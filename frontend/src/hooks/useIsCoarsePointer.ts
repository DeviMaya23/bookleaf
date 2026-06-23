import { useEffect, useState } from 'react'

const QUERY = '(pointer: coarse)'

export function useIsCoarsePointer(): boolean {
  const [isCoarse, setIsCoarse] = useState(() => window.matchMedia(QUERY).matches)

  useEffect(() => {
    const mediaQueryList = window.matchMedia(QUERY)
    function handleChange(event: MediaQueryListEvent) {
      setIsCoarse(event.matches)
    }
    mediaQueryList.addEventListener('change', handleChange)
    return () => mediaQueryList.removeEventListener('change', handleChange)
  }, [])

  return isCoarse
}
