<script>
  import {
    AdoptUnassignedCountdowns,
    AdvanceCountdown,
    DeleteCountdown,
    ListCountdownsForCampaign,
    ListUnassignedCountdowns,
    SaveCountdown,
    errorMessage
  } from './api.js'

  // With a campaignId the panel shows that campaign's clocks and creates new ones
  // against it. Without one it falls back to the clocks no campaign has claimed —
  // anything made before Phase 4, or during a fight run outside a campaign.
  // compact is the runner's rail: each clock shows only the beat for the number it
  // stands on, where the campaign tab lays out the whole trail.
  let { campaignId = null, compact = false } = $props()

  let clocks = $state([])
  let loose = $state([])
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let adding = $state(false)
  let form = $state(blank())
  let editing = $state(null)
  let draft = $state(null)

  function blank() {
    return { name: '', max: 6, kind: '' }
  }

  function remaining(clock) {
    return clock.max - clock.value
  }

  function beatAt(clock) {
    return (clock.beats ?? []).find((b) => b.at === remaining(clock))
  }

  function steps(max) {
    const n = Math.max(0, Math.trunc(Number(max)) || 0)
    return Array.from({ length: n + 1 }, (_, i) => n - i)
  }

  async function refresh() {
    try {
      clocks = (await (campaignId ? ListCountdownsForCampaign(campaignId) : ListUnassignedCountdowns())) ?? []
      loose = campaignId && !compact ? ((await ListUnassignedCountdowns()) ?? []) : []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

  // Clocks from before campaigns existed, or from a fight run without one, are
  // only reachable from a campaign-less fight. The campaign tab offers to take
  // them in, all at once — they were one pile before campaigns split them up.
  async function adopt() {
    busy = true
    try {
      await AdoptUnassignedCountdowns(campaignId)
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  async function add(event) {
    event.preventDefault()
    busy = true
    try {
      await SaveCountdown({
        id: null,
        campaignId,
        name: form.name,
        value: 0,
        max: Number(form.max),
        kind: form.kind,
        beats: []
      })
      form = blank()
      adding = false
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function edit(clock) {
    editing = clock.id
    draft = {
      name: clock.name,
      max: clock.max,
      kind: clock.kind,
      beats: Object.fromEntries((clock.beats ?? []).map((b) => [b.at, b.text]))
    }
  }

  function stopEditing() {
    editing = null
    draft = null
  }

  async function save(event, clock) {
    event.preventDefault()
    busy = true
    try {
      const updated = await SaveCountdown({
        id: clock.id,
        campaignId: clock.campaignId ?? null,
        name: draft.name,
        value: clock.value,
        max: Number(draft.max),
        kind: draft.kind,
        beats: Object.entries(draft.beats).map(([at, text]) => ({ at: Number(at), text: text ?? '' }))
      })
      clocks = clocks.map((c) => (c.id === updated.id ? updated : c))
      stopEditing()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  async function tick(clock, delta) {
    busy = true
    try {
      const updated = await AdvanceCountdown(clock.id, delta)
      clocks = clocks.map((c) => (c.id === updated.id ? updated : c))
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  async function remove(clock) {
    if (!confirm(`Delete countdown “${clock.name}”?`)) return
    busy = true
    try {
      await DeleteCountdown(clock.id)
      clocks = clocks.filter((c) => c.id !== clock.id)
      if (editing === clock.id) stopEditing()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }
</script>

<section class="clocks">
  <header>
    <h3>Countdowns</h3>
    <button class="btn ghost" onclick={() => (adding = !adding)}>{adding ? 'Cancel' : '+ Add'}</button>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if loose.length}
    <div class="loose">
      <p>
        {loose.length === 1 ? 'One countdown isn’t' : `${loose.length} countdowns aren’t`} in any campaign
        ({loose.map((c) => c.name).join(', ')}) — made before campaigns existed, or in a fight run without one.
      </p>
      <button class="btn" onclick={adopt} disabled={busy}>Move {loose.length === 1 ? 'it' : 'them'} here</button>
    </div>
  {/if}

  {#if adding}
    <form class="row" onsubmit={add}>
      <input bind:value={form.name} placeholder="Ritual completes" required />
      <input type="number" min="1" max="24" bind:value={form.max} title="Segments" />
      <input bind:value={form.kind} placeholder="kind (optional)" />
      <button class="btn primary" type="submit" disabled={busy || !form.name.trim()}>Add</button>
    </form>
  {/if}

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if !clocks.length}
    <p class="empty">No countdowns. Add one to track a ritual, a collapse, or reinforcements.</p>
  {:else}
    <ul>
      {#each clocks as clock (clock.id)}
        <li class:complete={clock.value >= clock.max}>
          <div class="top">
            <span class="name">{clock.name}</span>
            {#if clock.kind}<span class="chip">{clock.kind}</span>{/if}
            <span class="count">{clock.value}/{clock.max}</span>
            <button
              class="btn ghost"
              onclick={() => (editing === clock.id ? stopEditing() : edit(clock))}
              disabled={busy}
              title="Edit countdown and trail">✎</button>
            <button class="btn danger" onclick={() => remove(clock)} disabled={busy} title="Delete">✕</button>
          </div>
          <div class="bar">
            <button class="btn ghost step" onclick={() => tick(clock, -1)} disabled={busy || clock.value <= 0}>−</button>
            <div class="segs">
              {#each Array.from({ length: clock.max }, (_, i) => i < clock.value) as filled, i (i)}
                <span class="seg" class:filled></span>
              {/each}
            </div>
            <button class="btn ghost step" onclick={() => tick(clock, 1)} disabled={busy || clock.value >= clock.max}>+</button>
          </div>

          {#if editing === clock.id && draft}
            <form class="edit" onsubmit={(e) => save(e, clock)}>
              <div class="row">
                <input bind:value={draft.name} placeholder="Name" />
                <input type="number" min="1" max="24" bind:value={draft.max} title="Segments" />
                <input bind:value={draft.kind} placeholder="kind (optional)" />
              </div>
              <ol class="trail">
                {#each steps(draft.max) as at (at)}
                  <li class:now={at === Math.max(0, Number(draft.max) - clock.value)}>
                    <span class="at">{at}</span>
                    <textarea rows="2" bind:value={draft.beats[at]} placeholder="What happens at {at}?"></textarea>
                  </li>
                {/each}
              </ol>
              <div class="actions">
                <button class="btn ghost" type="button" onclick={stopEditing} disabled={busy}>Cancel</button>
                <button class="btn primary" type="submit" disabled={busy || !draft.name.trim()}>Save</button>
              </div>
            </form>
          {:else if compact}
            {@const beat = beatAt(clock)}
            {#if beat}
              <p class="beat"><span class="at">{beat.at}</span><span>{beat.text}</span></p>
            {/if}
          {:else if clock.beats?.length}
            <ol class="trail">
              {#each clock.beats as beat (beat.at)}
                <li class:now={beat.at === remaining(clock)} class:past={beat.at > remaining(clock)}>
                  <span class="at">{beat.at}</span>
                  <span>{beat.text}</span>
                </li>
              {/each}
            </ol>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .clocks {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.7rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  h3 {
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .error {
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }

  .loose {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.55rem 0.75rem;
    border: 1px dashed var(--gold-deep);
    border-radius: 8px;
  }

  .loose p {
    flex: 1;
    margin: 0;
    font-size: 0.8rem;
    color: var(--muted);
  }

  .row {
    display: flex;
    gap: 0.35rem;
  }

  .row input:first-child { flex: 1; min-width: 0; }
  .row input[type='number'] { width: 3.5rem; }
  .row input:nth-child(3) { width: 6rem; }

  ul {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li.complete .name { color: var(--fear); }

  .top {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    margin-bottom: 0.25rem;
  }

  .name {
    flex: 1;
    min-width: 0;
    font-size: 0.85rem;
  }

  .count {
    font-size: 0.75rem;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }

  .bar {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .step {
    padding: 0.05rem 0.4rem;
    font-size: 0.9rem;
    line-height: 1.2;
  }

  .segs {
    display: flex;
    flex: 1;
    gap: 0.15rem;
  }

  .seg {
    flex: 1;
    height: 0.6rem;
    border: 1px solid var(--line);
    border-radius: 2px;
    background: var(--bg);
  }

  .seg.filled {
    border-color: var(--fear);
    background: var(--fear);
  }

  .at {
    flex: none;
    min-width: 1.3rem;
    font-weight: 600;
    text-align: right;
    font-variant-numeric: tabular-nums;
    color: var(--fear);
  }

  .trail {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    margin: 0.45rem 0 0;
    padding: 0;
    list-style: none;
  }

  .trail li {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    padding: 0.15rem 0.35rem;
    border-radius: 4px;
    font-size: 0.8rem;
    line-height: 1.35;
  }

  .trail li.past { color: var(--muted); }
  .trail li.past .at { color: var(--muted); }

  .trail li.now {
    background: var(--bg);
    box-shadow: inset 2px 0 0 var(--fear);
  }

  .beat {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin: 0.45rem 0 0;
    padding: 0.3rem 0.45rem;
    border-radius: 4px;
    background: var(--bg);
    box-shadow: inset 2px 0 0 var(--fear);
    font-size: 0.8rem;
    line-height: 1.35;
  }

  .edit {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin-top: 0.45rem;
  }

  .edit .trail { margin: 0; }

  .edit .trail li {
    align-items: flex-start;
    padding: 0;
  }

  .edit .trail .at { padding-top: 0.4rem; }

  .edit textarea {
    flex: 1;
    min-width: 0;
    resize: vertical;
    font-size: 0.8rem;
  }

  .actions {
    display: flex;
    justify-content: flex-end;
    gap: 0.35rem;
  }
</style>
