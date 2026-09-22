import type { Poll } from './types'

const API_URL = (import.meta.env.VITE_API_URL ?? '').replace(/\/$/, '')

async function request<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${API_URL}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) {
    throw new Error(typeof body.error === 'string' ? body.error : 'Something went wrong. Please try again.')
  }
  return body as T
}

export function createPoll(title: string, options: string[], creatorId: string): Promise<Poll> {
  return request('/api/polls', { method: 'POST', body: JSON.stringify({ title, options, creatorId }) })
}

export function listPolls(voterId: string): Promise<Poll[]> {
  return request(`/api/polls?voterId=${encodeURIComponent(voterId)}`)
}

export function getPoll(id: string, voterId: string): Promise<Poll> {
  return request(`/api/polls/${encodeURIComponent(id)}?voterId=${encodeURIComponent(voterId)}`)
}

export function vote(pollId: string, optionId: string, voterId: string): Promise<Poll> {
  return request(`/api/polls/${encodeURIComponent(pollId)}/votes`, {
    method: 'POST',
    body: JSON.stringify({ optionId, voterId }),
  })
}

