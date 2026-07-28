<script>
  import { untrack } from 'svelte'
  import { TIERS, errorMessage } from './api.js'

  // Shared list/detail chrome for the adversary and environment browsers. Tier
  // and type are pushed down to the Go filter; the name search stays client-side
  // so typing doesn't re-hit the bridge on every keystroke.
  //
  // The parent swaps this component out for the homebrew form, so coming back
  // remounts it and the list reloads on its own — initialSlug is how the parent
  // restores the selection to the card that was just saved.
  // Bumping reloadToken re-runs the load — the parent needs it after a delete,
  // which leaves this component mounted with a stale list.
  // compact stacks the list above the detail instead of sitting beside it, for the
  // reference pane inside the homebrew forms where there is no room for two columns.
  let {
    types,
    load,
    emptyLabel,
    row,
    detail,
    onnew = null,
    newLabel = 'New',
    initialSlug = '',
    reloadToken = 0,
    compact = false
  } = $props()

  let tier = $state('')
  let type = $state('')
  let search = $state('')

  let items = $state([])
  let selectedSlug = $state(untrack(() => initialSlug))
  let loading = $state(true)
  let error = $state('')

  $effect(() => {
    const filter = { tier, type }
    reloadToken // tracked so a bump re-runs the load
    let stale = false
    loading = true
    load(filter)
      .then((rows) => {
        if (stale) return
        items = rows ?? []
        error = ''
      })
      .catch((e) => {
        if (stale) return
        items = []
        error = errorMessage(e)
      })
      .finally(() => {
        if (!stale) loading = false
      })
    return () => (stale = true)
  })

  const visible = $derived.by(() => {
    const needle = search.trim().toLowerCase()
    if (!needle) return items
    return items.filter((item) => item.name.toLowerCase().includes(needle))
  })

  const selected = $derived(visible.find((item) => item.slug === selectedSlug) ?? visible[0] ?? null)
</script>

<div class="browser" class:compact>
  <div class="filters">
    <input type="search" placeholder="Search by name…" bind:value={search} />
    <select bind:value={tier} aria-label="Tier">
      <option value="">All tiers</option>
      {#each TIERS as t (t)}
        <option value={t}>Tier {t}</option>
      {/each}
    </select>
    <select bind:value={type} aria-label="Type">
      <option value="">All types</option>
      {#each types as t (t)}
        <option value={t}>{t}</option>
      {/each}
    </select>
    <span class="count">{visible.length}</span>
    {#if onnew}
      <button class="btn primary" onclick={onnew}>{newLabel}</button>
    {/if}
  </div>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="split">
    <ul class="list">
      {#if loading}
        <li class="empty">Loading…</li>
      {:else if !visible.length}
        <li class="empty">{emptyLabel}</li>
      {:else}
        {#each visible as item (item.slug)}
          <li>
            <button
              class:selected={selected?.slug === item.slug}
              onclick={() => (selectedSlug = item.slug)}
            >
              {@render row(item)}
            </button>
          </li>
        {/each}
      {/if}
    </ul>

    <div class="detail">
      {#if selected}
        {#key selected.slug}
          {@render detail(selected)}
        {/key}
      {:else if !loading}
        <p class="placeholder">Nothing to show.</p>
      {/if}
    </div>
  </div>
</div>

<style>
  .browser {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .filters {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--line);
  }

  .filters input[type='search'] { flex: 1; }

  .count {
    font-size: 0.75rem;
    color: var(--muted);
    min-width: 2rem;
    text-align: right;
  }

  .error {
    margin: 0;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--danger);
    font-size: 0.8rem;
    color: var(--danger);
  }

  .split {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  .list {
    width: 18rem;
    flex-shrink: 0;
    margin: 0;
    padding: 0.5rem;
    border-right: 1px solid var(--line);
    list-style: none;
    overflow-y: auto;
  }

  .list button {
    display: block;
    width: 100%;
    padding: 0.45rem 0.55rem;
    border: 1px solid transparent;
    border-left: 2px solid transparent;
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .list button:hover {
    background: var(--panel);
    color: var(--text);
  }

  /* The accent edge carries the selection so the fill can stay quiet — a block
     bright enough to read against on its own would drown the row's own text. */
  .list button.selected {
    background: var(--panel-2);
    border-color: var(--line);
    border-left-color: var(--fear);
    color: var(--text);
  }

  .list button.selected :global(:first-child) { font-weight: 600; }

  .empty,
  .placeholder {
    padding: 1rem 0.5rem;
    font-size: 0.85rem;
    color: var(--muted);
    line-height: 1.5;
  }

  /* Capped for line length — at 1920 a card's features would otherwise run the
     full width of the pane. */
  .detail {
    flex: 1;
    min-width: 0;
    max-width: 60rem;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  .compact .detail { max-width: none; }

  .compact .filters {
    flex-wrap: wrap;
    gap: 0.35rem;
    padding: 0.5rem 0.6rem;
  }

  .compact .filters input[type='search'] { flex: 1 0 100%; }
  .compact .filters select { flex: 1; min-width: 0; }

  .compact .split { flex-direction: column; }

  .compact .list {
    width: auto;
    max-height: 11rem;
    border-right: none;
    border-bottom: 1px solid var(--line);
  }

  .compact .detail { padding: 0.75rem 0.85rem; }
</style>
