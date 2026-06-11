import { useEffect, useRef, useState } from 'react'

interface UseFieldAutosaveOptions<T> {
  isEmpty?: (value: T) => boolean
}

interface UseFieldAutosaveResult<T> {
  value: T
  onChange: (value: T) => void
  onBlur: () => void
}

// Tracks a locally-editable copy of `value`, resetting it whenever the source
// `value` changes (e.g. a different image is selected). On blur it saves only
// if the value actually changed; if `options.isEmpty` reports the value empty,
// it reverts to the last saved value instead of saving (the title field's
// "don't save empty" rule).
export function useFieldAutosave<T>(
  value: T,
  onSave: (value: T) => void,
  options?: UseFieldAutosaveOptions<T>,
): UseFieldAutosaveResult<T> {
  const [local, setLocal] = useState(value)
  const orig = useRef(value)

  useEffect(() => {
    setLocal(value)
    orig.current = value
  }, [value])

  const onBlur = () => {
    if (options?.isEmpty?.(local)) {
      setLocal(orig.current)
      return
    }
    if (local !== orig.current) {
      orig.current = local
      onSave(local)
    }
  }

  return { value: local, onChange: setLocal, onBlur }
}
