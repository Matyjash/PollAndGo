import { FormEvent, useEffect, useState } from 'react'
import { createPoll, getPoll, listPolls, vote } from './api'
import type { Poll } from './types'
import { useVoterId } from './useVoterId'

type Screen = 'home' | 'create' | 'poll'

export default function App() {
  const voterId = useVoterId()
  const [screen, setScreen] = useState<Screen>('home')
  const [poll, setPoll] = useState<Poll | null>(null)
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const [userPolls, setUserPolls] = useState<Poll[]>([])
  const [pollsLoading, setPollsLoading] = useState(true)
  const [pollsError, setPollsError] = useState('')
  const [resultsOnly, setResultsOnly] = useState(false)

  useEffect(() => {
    listPolls(voterId)
      .then(setUserPolls)
      .catch((cause) => setPollsError(cause instanceof Error ? cause.message : 'Unable to load your polls.'))
      .finally(() => setPollsLoading(false))
    const match = window.location.pathname.match(/^\/poll\/([a-zA-Z0-9]+)$/)
    if (match) void openPoll(match[1])
    // The voter ID is stable for the life of the page.
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [])

  async function openPoll(id: string, showResults = false) {
    setLoading(true)
    setError('')
    setResultsOnly(showResults)
    try {
      const loaded = await getPoll(id.trim().toUpperCase(), voterId)
      setPoll(loaded)
      setScreen('poll')
      window.history.pushState({}, '', `/poll/${loaded.id}`)
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Unable to open that poll.')
      setScreen('home')
    } finally { setLoading(false) }
  }

  function goHome() {
    setScreen('home')
    setPoll(null)
    setError('')
    setResultsOnly(false)
    window.history.pushState({}, '', '/')
  }

  return (
    <div className="app-shell">
      <header>
        <button className="brand" onClick={goHome} aria-label="Poll and Go home">
          <span className="brand-mark">P</span><span>Poll <b>&amp;</b> Go</span>
        </button>
      </header>
      <main>
        {screen === 'home' && <Home onOpen={(id) => openPoll(id)} onOpenResults={(id) => openPoll(id, true)} onCreate={() => { setError(''); setScreen('create') }} loading={loading} error={error} polls={userPolls} pollsLoading={pollsLoading} pollsError={pollsError} />}
        {screen === 'create' && <CreatePoll voterId={voterId} onCancel={goHome} onCreated={(created) => { setPoll(created); setUserPolls((current) => [created, ...current]); setScreen('poll'); window.history.pushState({}, '', `/poll/${created.id}`) }} />}
        {screen === 'poll' && poll && <PollView poll={poll} voterId={voterId} resultsOnly={resultsOnly} onChange={(updated) => { setPoll(updated); setUserPolls((current) => [updated, ...current.filter((item) => item.id !== updated.id)]) }} />}
      </main>
    </div>
  )
}

function Home({ onOpen, onOpenResults, onCreate, loading, error, polls, pollsLoading, pollsError }: { onOpen: (id: string) => void; onOpenResults: (id: string) => void; onCreate: () => void; loading: boolean; error: string; polls: Poll[]; pollsLoading: boolean; pollsError: string }) {
  const [id, setId] = useState('')
  function submit(event: FormEvent) { event.preventDefault(); if (id.trim()) onOpen(id) }
  return (
    <section className="hero">
      <div className="eyebrow"><span /> QUICK, SIMPLE POLLS</div>
      <h1>Ask. Share.<br /><em>Decide.</em></h1>
      <div className="home-grid">
        <button className="create-card" onClick={onCreate}>
          <span className="circle-icon">+</span>
          <span><strong>Create a poll</strong><small>Start with your own question</small></span>
          <span className="arrow">→</span>
        </button>
        <form className="join-card" onSubmit={submit}>
          <label htmlFor="poll-code">Already have a code?</label>
          <div className="join-row">
            <input id="poll-code" value={id} onChange={(event) => setId(event.target.value.toUpperCase())} placeholder="ENTER POLL ID" maxLength={12} autoComplete="off" />
            <button disabled={loading || !id.trim()}>{loading ? '…' : 'Go →'}</button>
          </div>
          {error && <p className="error" role="alert">{error}</p>}
        </form>
      </div>
      <section className="my-polls">
        <div className="my-polls-heading">
          <div><span>YOUR ACTIVITY</span><h2>Your polls</h2></div>
          {!pollsLoading && <small>{polls.length} {polls.length === 1 ? 'poll' : 'polls'}</small>}
        </div>
        {pollsLoading && <p className="polls-message">Loading your polls...</p>}
        {!pollsLoading && pollsError && <p className="error" role="alert">{pollsError}</p>}
        {!pollsLoading && !pollsError && polls.length === 0 && <p className="polls-message">Polls you create or vote in will appear here.</p>}
        <div className="poll-list">
          {polls.map((poll) => {
            const totalVotes = poll.options.reduce((sum, option) => sum + option.voteCount, 0)
            return <button key={poll.id} onClick={() => onOpenResults(poll.id)}>
              <span className="poll-list-main"><strong>{poll.title}</strong><small>{poll.createdByUser ? 'Created by you' : 'Voted'} · {totalVotes} {totalVotes === 1 ? 'vote' : 'votes'}</small></span>
              <span className="poll-list-code">{poll.id}</span>
              <span className="arrow" aria-hidden="true">→</span>
            </button>
          })}
        </div>
      </section>
    </section>
  )
}

function CreatePoll({ voterId, onCancel, onCreated }: { voterId: string; onCancel: () => void; onCreated: (poll: Poll) => void }) {
  const [title, setTitle] = useState('')
  const [options, setOptions] = useState(['', ''])
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)

  function updateOption(index: number, value: string) { setOptions((current) => current.map((item, i) => i === index ? value : item)) }
  async function submit(event: FormEvent) {
    event.preventDefault(); setLoading(true); setError('')
    try { onCreated(await createPoll(title.trim(), options.map((option) => option.trim()), voterId)) }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Unable to create the poll.') }
    finally { setLoading(false) }
  }

  return (
    <section className="panel create-panel">
      <button className="back" onClick={onCancel}>← Back</button>
      <div className="section-number">01</div>
      <h2>Create your poll</h2>
      <p>Write one clear question and give people a few good choices.</p>
      <form onSubmit={submit}>
        <label htmlFor="title">Your question</label>
        <input id="title" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Where should we have lunch?" maxLength={200} required autoFocus />
        <div className="options-heading"><label>Options</label><span>{options.length} / 20</span></div>
        <div className="option-inputs">
          {options.map((option, index) => (
            <div className="option-input" key={index}>
              <span>{String(index + 1).padStart(2, '0')}</span>
              <input value={option} onChange={(e) => updateOption(index, e.target.value)} placeholder={`Option ${index + 1}`} maxLength={120} required />
              {options.length > 2 && <button type="button" onClick={() => setOptions((current) => current.filter((_, i) => i !== index))} aria-label={`Remove option ${index + 1}`}>×</button>}
            </div>
          ))}
        </div>
        {options.length < 20 && <button className="add-option" type="button" onClick={() => setOptions((current) => [...current, ''])}>+ Add another option</button>}
        {error && <p className="error" role="alert">{error}</p>}
        <button className="primary full" disabled={loading}>{loading ? 'Creating…' : 'Create poll →'}</button>
      </form>
    </section>
  )
}

function PollView({ poll, voterId, resultsOnly, onChange }: { poll: Poll; voterId: string; resultsOnly: boolean; onChange: (poll: Poll) => void }) {
  const [selected, setSelected] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const total = poll.options.reduce((sum, option) => sum + option.voteCount, 0)
  const hasVoted = Boolean(poll.userVotedOptionId)

  async function submit(event: FormEvent) {
    event.preventDefault(); if (!selected || hasVoted) return
    setLoading(true); setError('')
    try { onChange(await vote(poll.id, selected, voterId)) }
    catch (cause) { setError(cause instanceof Error ? cause.message : 'Unable to save your vote.') }
    finally { setLoading(false) }
  }

  async function copyLink() {
    await navigator.clipboard.writeText(`${window.location.origin}/poll/${poll.id}`)
  }

  return (
    <section className="panel poll-panel">
      <div className="poll-meta"><span>LIVE POLL</span><button onClick={copyLink}>Copy link</button></div>
      <div className="share-code"><small>SHARE THIS CODE</small><strong>{poll.id}</strong></div>
      <h2>{poll.title}</h2>
      {hasVoted || resultsOnly ? (
        <div className="results" aria-live="polite">
          <p className="success">{hasVoted ? '✓ Your vote is in' : 'Current results'}</p>
          {poll.options.map((option) => {
            const percentage = total === 0 ? 0 : Math.round(option.voteCount / total * 100)
            return <div className={`result ${option.id === poll.userVotedOptionId ? 'chosen' : ''}`} key={option.id}>
              <div><span>{option.label}</span><b>{percentage}%</b></div>
              <div className="bar"><i style={{ width: `${percentage}%` }} /></div>
              <small>{option.voteCount} {option.voteCount === 1 ? 'vote' : 'votes'}</small>
            </div>
          })}
          <p className="total">{total} total {total === 1 ? 'vote' : 'votes'}</p>
        </div>
      ) : (
        <form className="ballot" onSubmit={submit}>
          <p>Choose one option</p>
          {poll.options.map((option) => <label className={selected === option.id ? 'selected' : ''} key={option.id}>
            <input type="radio" name="option" value={option.id} checked={selected === option.id} onChange={() => setSelected(option.id)} />
            <span className="radio" /><span>{option.label}</span>
          </label>)}
          {error && <p className="error" role="alert">{error}</p>}
          <button className="primary full" disabled={!selected || loading}>{loading ? 'Submitting…' : 'Submit vote →'}</button>
        </form>
      )}
    </section>
  )
}

