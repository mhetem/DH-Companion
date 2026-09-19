<script>
  import FeatureList from './FeatureList.svelte'
  import SrdText from '../SrdText.svelte'

  // actions is an optional snippet of buttons rendered beside the title,
  // matching how EnvironmentDetail takes its homebrew controls.
  let { card, actions = null } = $props()
</script>

<article class="detail">
  <header>
    <div class="titles">
      <h3>{card.name}</h3>
      <div class="chips">
        <span class="chip">Tier {card.tier}</span>
        {#if card.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
      </div>
    </div>
    {#if actions}
      <div class="actions">{@render actions()}</div>
    {/if}
  </header>

  <SrdText text={card.description} />

  <dl class="stats">
    <div><dt>Major threshold</dt><dd>{card.thresholdMajor}</dd></div>
    <div><dt>Severe threshold</dt><dd>{card.thresholdSevere}</dd></div>
    <div><dt>Base Armor Score</dt><dd>{card.baseScore}</dd></div>
  </dl>

  <!-- The printed thresholds are pre-level; the character sheet adds the wearer's
       level to both. Saying so here saves the GM looking the rule back up. -->
  <p class="note">Add the wearer’s level to both thresholds.</p>

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

  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(8rem, 1fr));
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

  .note {
    margin: 0;
    font-size: 0.78rem;
    font-style: italic;
    color: var(--muted);
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
