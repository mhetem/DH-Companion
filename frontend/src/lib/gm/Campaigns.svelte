<script>
  import {
    AdjustCampaignFear,
    DeleteCampaign,
    ListCampaigns,
    SaveCampaign,
    SetCampaignFear,
    errorMessage
  } from './api.js'
  import Countdowns from './Countdowns.svelte'
  import FearTracker from './FearTracker.svelte'
  import Notes from './Notes.svelte'
  import SessionLog from './SessionLog.svelte'

  // The last campaign opened is remembered the same way the runner remembers its
  // dice panel — reopening the section lands you back where you were.
  const LAST_KEY = 'gm.campaigns.last'

  let campaigns = $state([])
  let selectedId = $state(null)
  let loading = $state(true)
  let error = $state('')
  let saving = $state(false)
  let busy = $state(false)
  let tab = $state('sessions')
  let editingId = $state(null)
  let form = $state(blank())
  let adding = $state(false)

  const selected = $derived(campaigns.find((c) => c.id === selectedId) ?? null)

  function blank() {
    return { name: '', description: '' }
  }

  async function refresh() {
    try {
      campaigns = (await ListCampaigns()) ?? []
      if (!campaigns.some((c) => c.id === selectedId)) {
        const remembered = Number(localStorage.getItem(LAST_KEY))
        selectedId = campaigns.some((c) => c.id === remembered) ? remembered : (campaigns[0]?.id ?? null)
      }
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

  function select(campaign) {
    selectedId = campaign.id
    localStorage.setItem(LAST_KEY, String(campaign.id))
  }

  function edit(campaign) {
    editingId = campaign.id
    adding = true
    form = { name: campaign.name, description: campaign.description }
  }

  function reset() {
    editingId = null
    adding = false
    form = blank()
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      const saved = await SaveCampaign({ id: editingId, name: form.name, description: form.description })
      reset()
      await refresh()
      select(saved)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }

  async function remove(campaign) {
    if (!confirm(`Delete “${campaign.name}”? Its sessions, notes and countdowns go with it.`)) return
    try {
      await DeleteCampaign(campaign.id)
      if (editingId === campaign.id) reset()
      if (selectedId === campaign.id) selectedId = null
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    }
  }

  async function fear(fn) {
    busy = true
    try {
      const value = await fn()
      campaigns = campaigns.map((c) => (c.id === selectedId ? { ...c, currentFear: value } : c))
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  const adjustFear = (delta) => fear(() => AdjustCampaignFear(selectedId, delta))
  const setFear = (value) => fear(() => SetCampaignFear(selectedId, value))

  const TABS = [
    { id: 'sessions', label: 'Sessions' },
    { id: 'notes', label: 'Notes' },
    { id: 'clocks', label: 'Countdowns' }
  ]
</script>

<div class="campaigns">
  <header>
    <div>
      <h2>Campaigns</h2>
      <p class="blurb">Fear carries between fights here, not inside one — a campaign owns its sessions, notes and clocks.</p>
    </div>
    <button class="btn ghost" onclick={() => (adding ? reset() : (adding = true))}>
      {adding ? 'Cancel' : '+ New campaign'}
    </button>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if adding}
    <form onsubmit={submit}>
      <label>
        <span>Name</span>
        <input bind:value={form.name} placeholder="Age of Umbra" required />
      </label>
      <label>
        <span>Description</span>
        <input bind:value={form.description} placeholder="A grim march through the drowned south" />
      </label>
      <button class="btn primary" type="submit" disabled={saving || !form.name.trim()}>
        {editingId === null ? 'Create' : 'Save changes'}
      </button>
    </form>
  {/if}

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if !campaigns.length}
    <p class="empty">No campaigns yet. Create one to start logging sessions and notes.</p>
  {:else}
    <div class="picker">
      {#each campaigns as campaign (campaign.id)}
        <button class="tabbtn" class:on={campaign.id === selectedId} onclick={() => select(campaign)}>
          {campaign.name}
        </button>
      {/each}
    </div>
  {/if}

  {#if selected}
    <div class="detail">
      <div class="summary">
        <div class="about">
          <h3 class="cname">{selected.name}</h3>
          {#if selected.description}<p class="desc">{selected.description}</p>{/if}
          <div class="rowbtns">
            <button class="btn ghost" onclick={() => edit(selected)}>Edit</button>
            <button class="btn danger" onclick={() => remove(selected)}>Delete</button>
          </div>
        </div>
        <FearTracker fear={selected.currentFear} max={selected.fearMax} {busy} onadjust={adjustFear} onset={setFear} />
      </div>

      <div class="tabs">
        {#each TABS as t (t.id)}
          <button class="tabbtn" class:on={tab === t.id} onclick={() => (tab = t.id)}>{t.label}</button>
        {/each}
      </div>

      {#key selected.id}
        <div class="panel">
          {#if tab === 'sessions'}
            <SessionLog campaignId={selected.id} />
          {:else if tab === 'notes'}
            <Notes campaignId={selected.id} />
          {:else}
            <Countdowns campaignId={selected.id} />
          {/if}
        </div>
      {/key}
    </div>
  {/if}
</div>

<style>
  .campaigns {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.75rem;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  /* Children keep their natural height so the page itself is what scrolls, and are
     capped for line length — recaps and note bodies are prose. */
  .campaigns > :global(*) {
    flex-shrink: 0;
    width: 100%;
    max-width: 68rem;
  }


  header {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
  }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .blurb {
    margin: 0.25rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .error {
    margin: 0;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  form {
    display: flex;
    align-items: flex-end;
    gap: 0.6rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  form label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
  }

  form label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .picker,
  .tabs {
    display: flex;
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .tabs {
    border-bottom: 1px solid var(--line);
    padding-bottom: 0.4rem;
  }

  .tabbtn {
    padding: 0.3rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel);
    color: var(--muted);
    font: inherit;
    font-size: 0.8rem;
    cursor: pointer;
  }

  .tabbtn:hover { color: var(--text); }

  .tabbtn.on {
    border-color: var(--fear);
    color: var(--fear);
  }

  /* The page is the only scroll container — the tab panels below grow to their
     content and push this one taller, rather than each scrolling in its own box. */
  .detail {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }

  .summary {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
  }

  .about { flex: 1; min-width: 0; }

  .cname {
    margin: 0;
    font-size: 1rem;
  }

  .desc {
    margin: 0.2rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .rowbtns {
    display: flex;
    gap: 0.35rem;
    margin-top: 0.4rem;
  }

  .panel {
    display: flex;
    flex-direction: column;
  }

  .empty {
    margin: 0;
    font-size: 0.85rem;
    color: var(--muted);
  }
</style>
