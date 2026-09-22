import { useState } from 'react'

const STORAGE_KEY = 'pollandgo-voter-id'

function createId(): string {
  if (typeof crypto.randomUUID === 'function') return crypto.randomUUID()
  const bytes = crypto.getRandomValues(new Uint8Array(16))
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
}

function loadVoterId(): string {
  const existing = localStorage.getItem(STORAGE_KEY)
  if (existing) return existing
  const generated = createId()
  localStorage.setItem(STORAGE_KEY, generated)
  return generated
}

export function useVoterId(): string {
  const [voterId] = useState(loadVoterId)
  return voterId
}

