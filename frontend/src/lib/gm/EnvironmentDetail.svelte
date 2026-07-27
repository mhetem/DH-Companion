<script>
  import FeatureList from './FeatureList.svelte'

  // actions is an optional snippet of buttons rendered beside the title.
  let { card, actions = null } = $props()
</script>

<article class="detail">
  <header>
    <div class="titles">
      <h3>{card.name}</h3>
      <div class="chips">
        <span class="chip">Tier {card.tier}</span>
        <span class="chip">{card.type}</span>
        {#if card.difficulty}<span class="chip">Difficulty {card.difficulty}</span>{/if}
        {#if card.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
      </div>
    </div>
    {#if actions}
      <div class="actions">{@render actions()}</div>
    {/if}
  </header>

  {#if card.description}
    <p class="description">{card.description}</p>
  {/if}

  {#if card.impulses}
    <p class="line"><span class="key">Impulses</span> {card.impulses}</p>
  {/if}

  {#if card.potentialAdversaries?.length}
    <div>
      <span class="key">Potential Adversaries</span>
      <ul class="potential">
        {#each card.potentialAdversaries as entry (entry)}
          <li>{entry}</li>
        {/each}
      </ul>
    </div>
  {/if}

  {#if card.features?.length}
    <div class="features">
      <FeatureList features={card.features} />
    </div>
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

  h3 {
    margin: 0;
    font-size: 1.15rem;
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .actions {
    display: flex;
    gap: 0.4rem;
    margin-left: auto;
  }

  .description {
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .line {
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.5;
  }

  .key {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
    margin-right: 0.35rem;
  }

  .potential {
    margin: 0.3rem 0 0;
    padding-left: 1.1rem;
  }

  .potential li {
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .features {
    padding-top: 0.5rem;
    border-top: 1px solid var(--line);
  }
</style>
