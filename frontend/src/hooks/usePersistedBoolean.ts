import { useState } from 'react'

export function usePersistedBoolean(key: string, defaultValue: boolean): [boolean, (v: boolean) => void] {
  const [value, setValue] = useState<boolean>(() => {
    try {
      const stored = localStorage.getItem(key)
      if (stored === 'true') return true
      if (stored === 'false') return false
    } catch {
      // ignore
    }
    return defaultValue
  })

  function set(v: boolean) {
    setValue(v)
    try {
      localStorage.setItem(key, String(v))
    } catch {
      // ignore
    }
  }

  return [value, set]
}
