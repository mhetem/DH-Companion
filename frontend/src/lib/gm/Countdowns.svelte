<script>
  import {
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
  let { campaignId = null } = $props()

  let clocks = $state([])
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let adding = $state(false)
  let form = $state(blank())

  function blank() {
    return { name: '', max: 6, kind: '' }
  }

  async function refresh() {
    try {
      clocks = (await (campaignId ? ListCountdownsForCampaign(campaignId) : ListUnassignedCountdowns())) ?? []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

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
        kind: form.kind
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

  {#if adding}
    <form onsubmit={add}>
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

  form {
    display: flex;
    gap: 0.35rem;
  }

  form input:first-child { flex: 1; min-width: 0; }
  form input[type='number'] { width: 3.5rem; }
  form input:nth-child(3) { width: 6rem; }

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

  .empty {
    margin: 0;
    font-size: 0.8rem;
    color: var(--muted);
  }
</style>
