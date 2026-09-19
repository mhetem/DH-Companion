<script>
  import FeatureList from './FeatureList.svelte'
  import SrdText from '../SrdText.svelte'

  // actions is an optional snippet of buttons rendered beside the title,
  // matching how EnvironmentDetail takes its homebrew controls.
  let { card, actions = null } = $props()

  // The SRD prints damage type as an abbreviation in the tables; spell it out in
  // the detail pane, where there's room and no column to keep narrow.
  const DAMAGE_TYPES = {
    phy: 'Physical',
    mag: 'Magic',
    'phy/mag': 'Physical or magic'
  }

  const damageType = $derived(DAMAGE_TYPES[card.damageType] ?? card.damageType)
</script>

<article class="detail">
  <header>
    <div class="titles">
      <h3>{card.name}</h3>
      <div class="chips">
        <span class="chip">Tier {card.tier}</span>
        <span class="chip">{card.category}</span>
        <span class="chip">{card.type}</span>
        {#if card.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
      </div>
    </div>
    {#if actions}
      <div class="actions">{@render actions()}</div>
    {/if}
  </header>

  <SrdText text={card.description} />

  <dl class="stats">
    <div><dt>Trait</dt><dd>{card.trait}</dd></div>
    <div><dt>Range</dt><dd>{card.range}</dd></div>
    <div><dt>Damage</dt><dd>{card.damage}</dd></div>
    <div><dt>Damage type</dt><dd>{damageType}</dd></div>
    <div><dt>Burden</dt><dd>{card.burden}</dd></div>
  </dl>

  {#if card.feature}
    <div class="features">
      <FeatureList features={[card.feature]} />
    </div>
  {:else}
    <p class="none">No feature.</p>
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

  /* A grid rather than a row of chips: these are five values a GM reads off one
     at a time, and the labels have to stay attached to them. */
  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
    gap: 0.6rem;
    margin: 0;
  }

  .stats div {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
  }

  dt {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  dd {
    margin: 0;
    font-size: 0.9rem;
  }

  .features {
    padding-top: 0.5rem;
    border-top: 1px solid var(--line);
  }

  .none {
    margin: 0;
    padding-top: 0.5rem;
    border-top: 1px solid var(--line);
    font-size: 0.8rem;
    color: var(--muted);
  }
</style>
