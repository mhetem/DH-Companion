<script>
  import SrdText from '../SrdText.svelte'
  import { rarityDice } from './api.js'

  // actions is an optional snippet of buttons rendered beside the title,
  // matching how EnvironmentDetail takes its homebrew controls.
  let { card, actions = null } = $props()

  const dice = $derived(rarityDice(card.rarity))
</script>

<article class="detail">
  <header>
    <div class="titles">
      <h3>{card.name}</h3>
      <div class="chips">
        <span class="chip">{card.rarity}</span>
        {#if card.source === 'custom'}
          <span class="chip custom">Homebrew</span>
        {:else}
          <span class="chip">{card.table}</span>
          <span class="chip">Roll {card.roll}</span>
        {/if}
      </div>
    </div>
    {#if actions}
      <div class="actions">{@render actions()}</div>
    {/if}
  </header>

  <SrdText text={card.description} tone="plain" />

  <!-- Rarity is a property of the roll, not the entry: it says how many d12s to
       throw. Showing the dice makes clear why this row sits where it does. -->
  {#if card.source === 'custom'}
    <p class="note">
      Homebrew has no roll number — the randomizer draws it by rarity instead.
    </p>
  {:else if dice}
    <p class="note">Reachable from {dice} on the {card.table} table.</p>
  {/if}
</article>

<style>
  .detail {
    display: flex;
    flex-direction: column;
    gap: 0.75rem;
  }


  header {
    display: flex;
    align-items: flex-start;
    gap: 0.75rem;
  }

  .titles {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    min-width: 0;
  }

  .actions {
    display: flex;
    gap: 0.4rem;
    margin-left: auto;
  }

  h3 {
    margin: 0;
    font-size: 1.15rem;
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .note {
    margin: 0;
    font-size: 0.78rem;
    font-style: italic;
    color: var(--muted);
  }
</style>
