import { useSyncExternalStore } from 'react'

let maintenanceActive = false
const listeners = new Set<() => void>()

function subscribe(listener: () => void): () => void {
  listeners.add(listener)
  return () => listeners.delete(listener)
}

function getSnapshot(): boolean {
  return maintenanceActive
}

export function setMaintenanceActive(active: boolean): void {
  if (maintenanceActive === active) return
  maintenanceActive = active
  listeners.forEach((listener) => listener())
}

export function useMaintenanceActive(): boolean {
  return useSyncExternalStore(subscribe, getSnapshot)
}
