<script>
  import { untrack } from 'svelte'
  import { gm } from '../../../wailsjs/go/models'
  import AdversaryDetail from './AdversaryDetail.svelte'
  import EnvironmentDetail from './EnvironmentDetail.svelte'
  import BudgetMeter from './BudgetMeter.svelte'
  import {
    ADVERSARY_TYPES,
    BrowseAdversaries,
    BrowseEnvironments,
    ComputeBudget,
    DIFFICULTIES,
    DeleteEncounter,
    GetEncounter,
    TIERS,
    SaveEncounter,
    costLabel,
    errorMessage
  } from './api.js'

  let { encounterId = null, parties = [], onsaved, ondeleted } = $props()

  // The parent keys this component by encounter, so an instance always edits the
  // one it was mounted with — read the id once instead of tracking it.
  const initialId = untrack(() => encounterId)

  let name = $state('')
  let partyId = $state('')
  let environmentSlug = $state('')
  let difficulty = $state('')
  let picks = $state([])

  let savedId = $state(initialId)
  let loading = $state(initialId !== null)
  let saving = $state(false)
  let error = $state('')
  let dirty = $state(false)

  // The whole roster is pulled once and filtered in the browser — 130-odd SRD
  // cards plus homebrew is small, and the picker stays instant while typing.
  let catalog = $state([])
  let environments = $state([])
  let search = $state('')
  let tierFilter = $state('')
  let typeFilter = $state('')
  let expandedSlug = $state('')

  const party = $derived(parties.find((p) => String(p.id) === String(partyId)) ?? null)
  const environment = $derived(environments.find((e) => e.slug === environmentSlug) ?? null)
  const totalCount = $derived(picks.reduce((sum, p) => sum + p.count, 0))

  let budget = $state(null)

  Promise.all([BrowseAdversaries({ tier: '', type: '' }), BrowseEnvironments({ tier: '', type: '' })])
    .then(([a, e]) => {
      catalog = a ?? []
      environments = e ?? []
    })
    .catch((e) => (error = errorMessage(e)))

  if (initialId !== null) {
    GetEncounter(initialId)
      .then(apply)
      .catch((e) => (error = errorMessage(e)))
      .finally(() => (loading = false))
  }

  function apply(view) {
    savedId = view.id
    name = view.name
    partyId = view.partyId == null ? '' : String(view.partyId)
    environmentSlug = view.environmentSlug ?? ''
    picks = (view.adversaries ?? []).map((a) => ({ ...a }))
    dirty = false
  }

  const visible = $derived.by(() => {
    const needle = search.trim().toLowerCase()
    return catalog.filter((card) => {
      if (tierFilter && card.tier !== tierFilter) return false
      if (typeFilter && card.type !== typeFilter) return false
      if (needle && !card.name.toLowerCase().includes(needle)) return false
      return true
    })
  })

  // Re-price on every edit: party, difficulty, and the roster all feed the meter.
  $effect(() => {
    const settings = party
      ? { partySize: party.size, partyTier: party.tier, difficulty }
      : null
    const roster = picks.map((p) => ({ ...p }))
    if (!settings) {
      budget = null
      return
    }
    let stale = false
    ComputeBudget(settings, roster)
      .then((b) => {
        if (!stale) budget = b
      })
      .catch((e) => {
        if (!stale) error = errorMessage(e)
      })
    return () => (stale = true)
  })

  function add(card) {
    const existing = picks.find((p) => p.slug === card.slug)
    if (existing) {
      existing.count += 1
    } else {
      picks.push({ ...card, count: 1 })
    }
    dirty = true
  }

  function step(pick, delta) {
    const next = pick.count + delta
    if (next < 1) {
      remove(pick)
      return
    }
    pick.count = next
    dirty = true
  }

  function remove(pick) {
    picks = picks.filter((p) => p.slug !== pick.slug)
    dirty = true
  }

  function touch() {
    dirty = true
  }

  async function save() {
    saving = true
    try {
      const input = gm.EncounterInput.createFrom({
        id: savedId,
        name: name.trim(),
        partyId: partyId === '' ? null : Number(partyId),
        environmentSlug: environmentSlug === '' ? null : environmentSlug,
        // Unresolved picks (a homebrew card deleted out from under the
        // encounter) ride along in the SRD bucket — lookup checks both lists.
        adversaries: toPicks(picks.filter((p) => p.source !== 'custom')),
        customAdversaries: toPicks(picks.filter((p) => p.source === 'custom'))
      })
      const view = await SaveEncounter(input)
      apply(view)
      error = ''
      onsaved?.(view)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }

  function toPicks(list) {
    return list.map((p) => ({ slug: p.slug, count: p.count }))
  }

  async function destroy() {
    if (savedId === null) return
    if (!confirm(`Delete encounter “${name}”?`)) return
    try {
      await DeleteEncounter(savedId)
      ondeleted?.(savedId)
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

<div class="builder">
  <header>
    <input
      class="title"
      bind:value={name}
      oninput={touch}
      placeholder="Encounter name"
      aria-label="Encounter name"
    />
    <select bind:value={partyId} onchange={touch} aria-label="Party">
      <option value="">No party</option>
      {#each parties as p (p.id)}
        <option value={String(p.id)}>{p.name} ({p.size} · T{p.tier})</option>
      {/each}
    </select>
    <select bind:value={difficulty} aria-label="Difficulty">
      {#each DIFFICULTIES as d (d.value)}
        <option value={d.value}>{d.label}</option>
      {/each}
    </select>
    <button class="btn primary" onclick={save} disabled={saving || !name.trim()}>
      {savedId === null ? 'Create' : dirty ? 'Save' : 'Saved'}
    </button>
    {#if savedId !== null}
      <button class="btn danger" onclick={destroy}>Delete</button>
    {/if}
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if loading}
    <p class="loading">Loading encounter…</p>
  {:else}
    <div class="body">
      <section class="picker">
        <div class="filters">
          <input type="search" placeholder="Search adversaries…" bind:value={search} />
          <select bind:value={tierFilter} aria-label="Tier">
            <option value="">All tiers</option>
            {#each TIERS as t (t)}
              <option value={t}>Tier {t}</option>
            {/each}
          </select>
          <select bind:value={typeFilter} aria-label="Type">
            <option value="">All types</option>
            {#each ADVERSARY_TYPES as t (t)}
              <option value={t}>{t}</option>
            {/each}
          </select>
        </div>

        <ul class="catalog">
          {#each visible as card (card.slug)}
            <li>
              <div class="row">
                <button
                  class="expand"
                  onclick={() => (expandedSlug = expandedSlug === card.slug ? '' : card.slug)}
                  aria-expanded={expandedSlug === card.slug}
                >
                  <span class="name">{card.name}</span>
                  <span class="meta">
                    Tier {card.tier} · {card.type} · {costLabel(card.type)}
                    {#if card.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
                  </span>
                </button>
                <button class="btn add" onclick={() => add(card)} aria-label="Add {card.name}">+</button>
              </div>
              {#if expandedSlug === card.slug}
                <div class="expanded">
                  <AdversaryDetail {card} />
                </div>
              {/if}
            </li>
          {:else}
            <li class="empty">No adversaries match these filters.</li>
          {/each}
        </ul>
      </section>

      <aside>
        <BudgetMeter {budget} />

        <section class="roster">
          <h3>Roster <span class="count">{totalCount}</span></h3>
          {#if !picks.length}
            <p class="empty">Nothing picked yet — add adversaries from the left.</p>
          {:else}
            <ul>
              {#each picks as pick (pick.slug)}
                <li class:unresolved={pick.unresolved}>
                  <div class="who">
                    <span class="name">{pick.name}</span>
                    <span class="meta">
                      {#if pick.unresolved}
                        Missing card — the homebrew it referenced was deleted
                      {:else}
                        Tier {pick.tier} · {pick.type} · {costLabel(pick.type)}
                      {/if}
                    </span>
                  </div>
                  <div class="stepper">
                    <button class="btn ghost" onclick={() => step(pick, -1)} aria-label="One fewer">−</button>
                    <span class="n">{pick.count}</span>
                    <button class="btn ghost" onclick={() => step(pick, 1)} aria-label="One more">+</button>
                  </div>
                  <button class="btn danger" onclick={() => remove(pick)} aria-label="Remove">×</button>
                </li>
              {/each}
            </ul>
          {/if}
        </section>

        <section class="environment">
          <h3>Environment</h3>
          <select
            bind:value={environmentSlug}
            onchange={touch}
            aria-label="Environment"
          >
            <option value="">None</option>
            {#each environments as e (e.slug)}
              <option value={e.slug}>{e.name} (T{e.tier})</option>
            {/each}
          </select>
          {#if environment}
            <div class="env-card">
              <EnvironmentDetail card={environment} />
            </div>
          {/if}
        </section>
      </aside>
    </div>
  {/if}
</div>

<style>
  .builder {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  header {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--line);
  }

  .title {
    flex: 1;
    font-size: 1rem;
  }

  .error {
    margin: 0;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--danger);
    font-size: 0.8rem;
    color: var(--danger);
  }

  .loading {
    padding: 1rem;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .body {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  .picker {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
    border-right: 1px solid var(--line);
  }

  .filters {
    display: flex;
    gap: 0.5rem;
    padding: 0.6rem 0.75rem;
    border-bottom: 1px solid var(--line);
  }

  .filters input[type='search'] { flex: 1; }

  .catalog {
    flex: 1;
    margin: 0;
    padding: 0.4rem;
    list-style: none;
    overflow-y: auto;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .expand {
    flex: 1;
    min-width: 0;
    padding: 0.4rem 0.5rem;
    border: 1px solid transparent;
    border-radius: 6px;
    background: transparent;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .expand:hover { background: var(--panel); }

  .add {
    width: 2rem;
    padding: 0.25rem 0;
    text-align: center;
  }

  .expanded {
    margin: 0.35rem 0 0.6rem;
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  aside {
    display: flex;
    flex-direction: column;
    gap: 1rem;
    width: 23rem;
    flex-shrink: 0;
    padding: 0.75rem 1rem;
    overflow-y: auto;
  }

  h3 {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0 0 0.5rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  .count {
    padding: 0 0.35rem;
    border-radius: 999px;
    background: var(--panel-2);
    color: var(--text);
  }

  .roster ul {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .roster li {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.4rem 0.5rem;
    margin-bottom: 0.3rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel);
  }

  .roster li.unresolved { border-color: var(--danger); }

  .who {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name {
    display: block;
    font-size: 0.85rem;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    font-size: 0.7rem;
    color: var(--muted);
  }

  .stepper {
    display: flex;
    align-items: center;
    gap: 0.15rem;
  }

  .stepper .n {
    min-width: 1.25rem;
    font-size: 0.85rem;
    text-align: center;
  }

  .env-card {
    margin-top: 0.6rem;
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .empty {
    padding: 0.75rem 0.5rem;
    border: none;
    background: none;
  }
</style>
