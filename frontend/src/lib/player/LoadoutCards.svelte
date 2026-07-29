<script>
  import SrdText from '../SrdText.svelte'
  import { player } from '../../../wailsjs/go/models'
  import { SwapDomainCard, errorMessage, recallCost } from './api.js'

  // The loadout is the five cards whose effects you can actually use, so the sheet
  // shows them in full. Domain Cards stays the place to manage the collection; this
  // is the place to read it mid-session and change your mind.
  //
  // `resting` is the whole rules difference: recalling a card from the vault costs
  // Stress equal to its Recall Cost, and that cost is waived during a rest. The
  // backend decides which it is — this only has to say so before you press.
  let {
    characterId,
    loadout = [],
    vault = [],
    loadoutMax = 5,
    stressRoom = null,
    compact = false,
    resting = false,
    onswap
  } = $props()

  // Stands in for "the empty slot" as a picking target. Card slugs are camelCase
  // identifiers, so this can't collide with one.
  const SPARE = '+'

  let picking = $state('')
  let busy = $state(false)
  let error = $state('')

  const room = $derived(loadout.length < loadoutMax)
  const spare = $derived(loadoutMax - loadout.length)

  function cost(entry) {
    return resting ? 0 : recallCost(entry.card)
  }

  // Without a character loaded there's nothing to check against, so the backend's
  // error is the gate instead of a greyed-out button.
  function affordable(entry) {
    return stressRoom === null || cost(entry) <= stressRoom
  }

  function toggle(slug) {
    error = ''
    picking = picking === slug ? '' : slug
  }

  async function swap(option, vaultSlug) {
    if (busy) return
    busy = true
    error = ''
    try {
      const result = await SwapDomainCard(
        player.SwapInput.createFrom({
          characterId,
          recall: option.cardSlug,
          vault: vaultSlug,
          resting
        })
      )
      picking = ''
      onswap?.(result)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }
</script>

{#snippet vaultPicks(vaultSlug)}
  <div class="picks">
    {#each vault as option (option.cardSlug)}
      <button class="pick" disabled={busy || !affordable(option)} onclick={() => swap(option, vaultSlug)}>
        <span class="name">{option.card.name}</span>
        <span class="meta">{option.card.domain} · Level {option.card.level} · {option.card.type}</span>
        <span class="price" class:free={cost(option) === 0}>
          {cost(option) === 0 ? 'free' : `${cost(option)} Stress`}
        </span>
      </button>
    {/each}
  </div>
{/snippet}

<div class="loadout" class:compact>
  {#if error}<p class="error">{error}</p>{/if}

  {#if !loadout.length}
    <p class="empty">Nothing in your loadout — equip a card in Domain Cards.</p>
  {/if}

  <ul class="cards">
    {#each loadout as entry (entry.cardSlug)}
      <li class:unresolved={entry.unresolved}>
        <div class="head">
          <div class="what">
            <span class="name">{entry.card.name}</span>
            <span class="meta">
              {entry.card.domain} · Level {entry.card.level} · {entry.card.type}
              {#if recallCost(entry.card)}· Recall {recallCost(entry.card)}{/if}
            </span>
          </div>
          <button class="btn ghost" disabled={busy || !vault.length} onclick={() => toggle(entry.cardSlug)}>
            {picking === entry.cardSlug ? 'Cancel' : 'Swap'}
          </button>
        </div>

        {#if !compact && entry.card.description}
          <SrdText text={entry.card.description} />
        {/if}

        {#if picking === entry.cardSlug}
          <div class="swap">
            <p class="need">
              Vault <strong>{entry.card.name}</strong> and recall which card?
              {#if resting}
                Swapping is free on a rest.
              {:else}
                Outside a rest it costs Stress equal to the incoming card's Recall Cost.
              {/if}
            </p>
            {@render vaultPicks(entry.cardSlug)}
          </div>
        {/if}
      </li>
    {/each}
  </ul>

  {#if room && vault.length}
    <div class="spare">
      <button class="btn ghost" disabled={busy} onclick={() => toggle(SPARE)}>
        {picking === SPARE ? 'Cancel' : `Recall a card — ${spare} slot${spare === 1 ? '' : 's'} free`}
      </button>
      {#if picking === SPARE}
        {@render vaultPicks('')}
      {/if}
    </div>
  {/if}

  {#if !vault.length}
    <p class="hint">Your vault is empty, so there's nothing to swap in.</p>
  {/if}
</div>

<style>
  .loadout {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .cards {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .cards li {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0.6rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
  }

  .compact .cards li {
    gap: 0.3rem;
    padding: 0.4rem 0.5rem;
  }

  .cards li.unresolved { border-color: var(--danger); }

  .head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .what {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name { font-size: 0.9rem; }

  .compact .name { font-size: 0.82rem; }

  .meta {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .swap {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding-top: 0.4rem;
    border-top: 1px solid var(--line);
  }

  .need {
    margin: 0;
    font-size: 0.74rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .need strong {
    font-weight: 600;
    color: var(--text);
  }

  .picks {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    max-height: 16rem;
    overflow-y: auto;
  }

  .pick {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    width: 100%;
    padding: 0.35rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .pick:hover:not(:disabled) { border-color: var(--hope); }

  .pick:disabled {
    cursor: default;
    opacity: 0.45;
  }

  .pick .name { flex: 0 0 auto; }

  .pick .meta { flex: 1; }

  .price {
    flex: 0 0 auto;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--gold);
  }

  .price.free { color: var(--hope); }

  .spare {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
  }

  .spare > button { align-self: flex-start; }

  .hint {
    margin: 0;
    font-size: 0.72rem;
    line-height: 1.5;
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
</style>
