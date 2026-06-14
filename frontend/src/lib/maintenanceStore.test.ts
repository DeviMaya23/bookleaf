import { renderHook, act } from '@testing-library/react'
import { describe, it, expect } from 'vitest'
import { useMaintenanceActive, setMaintenanceActive } from './maintenanceStore'

describe('maintenanceStore', () => {
  it('defaults to false', () => {
    const { result } = renderHook(() => useMaintenanceActive())

    expect(result.current).toBe(false)
  })

  it('updates subscribed components when setMaintenanceActive is called', () => {
    const { result } = renderHook(() => useMaintenanceActive())

    act(() => setMaintenanceActive(true))
    expect(result.current).toBe(true)

    act(() => setMaintenanceActive(false))
    expect(result.current).toBe(false)
  })
})
