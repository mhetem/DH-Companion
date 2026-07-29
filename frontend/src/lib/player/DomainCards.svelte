<script>
  import EmptyState from '../EmptyState.svelte'
  import SrdText from '../SrdText.svelte'
  import { active } from './active.svelte.js'
  import {
    AddDomainCard,
    AvailableDomainCards,
    GetCharacter,
    GetLoadout,
    MoveDomainCard,
    RemoveDomainCard,
    errorMessage
  } from './api.js'

  let character = $state(null)
  let loadout = $state(null)
  let available = $state([])
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let search = $state('')
  let selected = $state(null)

  const id = $derived(active.id)

  $effect(() => {
    const target = id
    if (!target) {
      loadout = null
      character = null
      loading = false
      return
    }
    loading = true
    Promise.all([GetCharacter(target), GetLoadout(target), AvailableDomainCards(target)])
      .then(([c, l, a]) => {
        character = c
        loadout = l
        available = a ?? []
        error = ''
      })
      .catch((e) => {
        error = errorMessage(e)
        loadout = null
      })
      .finally(() => (loading = false))
  })

  // The name search stays client-side so typing doesn't re-cross the bridge, the
  // same split the GM card browser uses.
  const filtered = $derived(
    available.filter((card) => card.name.toLowerCase().includes(search.trim().toLowerCase()))
  )

  const full = $derived((loadout?.loadout.length ?? 0) >= (loadout?.loadoutMax ?? 5))

  // Two separate caps. `full` is the five-card loadout; `maxedOut` is how many cards
  // you own at all — two at level 1, one more per level and per extra-card advancement.
  const maxedOut = $derived((loadout?.held ?? 0) >= (loadout?.allowance ?? 0))

  async function run(fn) {
    if (busy) return
    busy = true
    try {
      loadout = await fn()
      available = (await AvailableDomainCards(id)) ?? []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  const add = (slug, location) => run(() => AddDomainCard(id, slug, location))
  const move = (slug, location) => run(() => MoveDomainCard(id, slug, location))

  function remove(entry) {
    if (!confirm(`Remove ${entry.card.name} from this character?`)) return
    return run(() => RemoveDomainCard(id, entry.cardSlug))
  }
</script>

{#if !id}
  <div class="pane">
    <EmptyState title="No character open">Pick one in Characters first.</EmptyState>
  </div>
{:else if loading}
  <div class="pane"><p class="empty">Loading cards…</p></div>
{:else if !loadout}
  <div class="pane"><p class="error">{error}</p></div>
{:else}
  <div class="pane">
    <header>
      <div class="intro">
        <h2>Domain cards</h2>
        <p class="blurb">
          {character?.name} draws on {character?.domains?.join(' & ') || '—'}.
          Your loadout holds {loadout.loadoutMax}; everything else sits in the vault.
        </p>
      </div>
      <div class="tallies">
        <span class="tally" class:full>{loadout.loadout.length}/{loadout.loadoutMax} in loadout</span>
        <span class="tally" class:full={maxedOut}>{loadout.held}/{loadout.allowance} cards known</span>
      </div>
    </header>

    {#if error}<p class="error">{error}</p>{/if}

    <div class="columns">
      <section>
        <h3>Loadout</h3>
        <ul class="cards">
          {#each loadout.loadout as entry (entry.cardSlug)}
            <li class:unresolved={entry.unresolved}>
              <button class="face" onclick={() => (selected = entry)}>
                <span class="name">{entry.card.name}</span>
                <span class="meta">{entry.card.domain} · Level {entry.card.level} · {entry.card.type}</span>
                {#if entry.card.recallCost}<span class="recall">Recall {entry.card.recallCost}</span>{/if}
              </button>
              <div class="ops">
                <button class="btn ghost" disabled={busy} onclick={() => move(entry.cardSlug, 'vault')}>Vault</button>
                <button class="btn danger" disabled={busy} onclick={() => remove(entry)}>✕</button>
              </div>
            </li>
          {/each}
          {#if !loadout.loadout.length}
            <li class="empty">Nothing in your loadout.</li>
          {/if}
        </ul>
      </section>

      <section>
        <h3>Vault</h3>
        <ul class="cards">
          {#each loadout.vault as entry (entry.cardSlug)}
            <li class:unresolved={entry.unresolved}>
              <button class="face" onclick={() => (selected = entry)}>
                <span class="name">{entry.card.name}</span>
                <span class="meta">{entry.card.domain} · Level {entry.card.level} · {entry.card.type}</span>
              </button>
              <div class="ops">
                <button class="btn ghost" disabled={busy || full} onclick={() => move(entry.cardSlug, 'loadout')}>
                  Equip
                </button>
                <button class="btn danger" disabled={busy} onclick={() => remove(entry)}>✕</button>
              </div>
            </li>
          {/each}
          {#if !loadout.vault.length}
            <li class="empty">Vault is empty.</li>
          {/if}
        </ul>
      </section>

      <section>
        <h3>Add a card</h3>
        {#if maxedOut}
          <p class="capped">
            You know all {loadout.allowance} cards your level allows. Levelling up is what earns another.
          </p>
        {/if}
        <input bind:value={search} placeholder="Search your domains…" />
        <ul class="cards scroll">
          {#each filtered as card (card.slug)}
            <li>
              <button class="face" onclick={() => (selected = { card, cardSlug: card.slug })}>
                <span class="name">{card.name}</span>
                <span class="meta">{card.domain} · Level {card.level} · {card.type}</span>
              </button>
              <div class="ops">
                <button class="btn ghost" disabled={busy || full || maxedOut} onclick={() => add(card.slug, 'loadout')}>
                  Loadout
                </button>
                <button class="btn ghost" disabled={busy || maxedOut} onclick={() => add(card.slug, 'vault')}>
                  Vault
                </button>
              </div>
            </li>
          {/each}
          {#if !filtered.length}
            <li class="empty">Nothing left at your level in your domains.</li>
          {/if}
        </ul>
      </section>
    </div>

    {#if selected}
      <aside class="detail">
        <div class="detail-head">
          <div>
            <h3>{selected.card.name}</h3>
            <p class="meta">
              {selected.card.domain} · Level {selected.card.level} · {selected.card.type}
              {#if selected.card.recallCost}· Recall {selected.card.recallCost}{/if}
            </p>
          </div>
          <button class="btn ghost" onclick={() => (selected = null)}>✕</button>
        </div>
        <SrdText text={selected.card.description} />
      </aside>
    {/if}
  </div>
{/if}

<style>
  .pane {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  header .intro { flex: 1; }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  h3 {
    margin: 0 0 0.5rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .blurb {
    margin: 0.25rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .tallies {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
  }

  .capped {
    margin: 0;
    padding: 0.4rem 0.6rem;
    border: 1px solid var(--gold-deep);
    border-radius: 6px;
    font-size: 0.75rem;
    line-height: 1.45;
    color: var(--muted);
  }

  .tally {
    padding: 0.3rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    font-size: 0.75rem;
    color: var(--hope);
  }

  .tally.full {
    border-color: var(--gold);
    color: var(--gold);
  }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .columns {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
    gap: 0.9rem;
    align-items: start;
  }

  section {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    min-width: 0;
  }

  .cards {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .cards.scroll {
    max-height: 26rem;
    overflow-y: auto;
  }

  .cards li {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    padding: 0.1rem 0.4rem 0.1rem 0.1rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .cards li.unresolved { border-color: var(--danger); }

  .cards li.empty {
    border: none;
    background: none;
    padding: 0.3rem 0;
  }

  .face {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.1rem;
    min-width: 0;
    padding: 0.4rem 0.5rem;
    border: none;
    background: none;
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .name { font-size: 0.85rem; }

  .meta {
    margin: 0;
    font-size: 0.72rem;
    color: var(--muted);
  }

  .recall {
    font-size: 0.68rem;
    color: var(--gold);
  }

  .ops {
    display: flex;
    gap: 0.3rem;
  }

  .detail {
    max-width: 44rem;
    margin-top: 1rem;
    padding: 0.85rem 0.95rem;
    border: 1px solid var(--gold-deep);
    border-radius: 10px;
    background: var(--panel);
  }

  .detail-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .detail h3 {
    margin: 0;
    font-size: 1rem;
    text-transform: none;
    letter-spacing: 0;
    color: var(--text);
  }

</style>
