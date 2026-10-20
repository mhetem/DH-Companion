<script>
  import FeatureList from './FeatureList.svelte'

  // actions is an optional snippet of buttons rendered beside the title.
  let { card, actions = null } = $props()

  // A colossus keeps Difficulty, HP and the standard attack on its segments; the
  // card itself is the framework — thresholds, Stress, and features for the whole.
  // The card-level fields are ignored rather than trusted to be blank, since the
  // homebrew form keeps whatever was typed before the type was switched.
  const colossus = $derived(card.type === 'Colossus')
  const segments = $derived(colossus ? (card.segments ?? []) : [])
  const roster = $derived(segments.map((s) => (s.count > 1 ? `${s.name} ×${s.count}` : s.name)).join(', '))

  const attack = $derived(card.standardAttack ?? {})
  const hasAttack = $derived(!colossus && Boolean(attack.name || attack.damage))

  function hasSegmentAttack(s) {
    return Boolean(s.standardAttack?.name || s.standardAttack?.damage)
  }

  const stats = $derived(
    [
      colossus ? null : { label: 'Difficulty', value: card.difficulty },
      { label: 'Thresholds', value: thresholds(card) },
      colossus ? null : { label: 'HP', value: card.hp },
      { label: 'Stress', value: card.stress },
      // Only Hordes print a unit count on the card.
      card.type === 'Horde' ? { label: 'Horde', value: card.hordeNumber } : null
    ].filter((s) => s && s.value)
  )

  function thresholds(a) {
    if (!a.thresholdMinor && !a.thresholdMajor) return ''
    return `${a.thresholdMinor || '—'} / ${a.thresholdMajor || '—'}`
  }
</script>

<article class="detail">
  <header>
    <div class="titles">
      <h3>{card.name}</h3>
      <div class="chips">
        <span class="chip">Tier {card.tier}</span>
        <span class="chip">{card.type}</span>
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

  {#if stats.length}
    <dl class="stats">
      {#each stats as stat (stat.label)}
        <div>
          <dt>{stat.label}</dt>
          <dd>{stat.value}</dd>
        </div>
      {/each}
    </dl>
  {/if}

  {#if colossus && card.size}
    <p class="line"><span class="key">Size</span> {card.size}</p>
  {/if}
  {#if roster}
    <p class="line"><span class="key">Segments</span> {roster}</p>
  {/if}
  {#if card.motives}
    <p class="line"><span class="key">Motives &amp; Tactics</span> {card.motives}</p>
  {/if}
  {#if card.experiences}
    <p class="line"><span class="key">Experience</span> {card.experiences}</p>
  {/if}

  {#if hasAttack}
    <p class="line">
      <span class="key">Attack</span>
      {attack.name}{attack.modifier ? ` ${attack.modifier}` : ''}
      {#if attack.range}<span class="dim">· {attack.range}</span>{/if}
      {#if attack.damage}<span class="dim">· {attack.damage} {attack.damageType}</span>{/if}
    </p>
  {/if}

  {#if card.features?.length}
    <div class="features">
      {#if colossus}<h4 class="group">Every segment</h4>{/if}
      <FeatureList features={card.features} />
    </div>
  {/if}

  {#each segments as s, i (s.name + i)}
    <section class="segment">
      <h4>
        {s.name}
        {#if s.count > 1}<span class="dim">×{s.count}</span>{/if}
      </h4>
      {#if s.difficulty || s.hp}
        <p class="line">
          {#if s.difficulty}<span class="key">Difficulty</span> {s.difficulty}{/if}
          {#if s.hp}<span class="key" class:spaced={s.difficulty}>HP</span> {s.hp}{/if}
        </p>
      {/if}
      {#if s.adjacent}
        <p class="line"><span class="key">Adjacent</span> {s.adjacent}</p>
      {/if}
      {#if hasSegmentAttack(s)}
        <p class="line">
          <span class="key">Attack</span>
          {s.standardAttack.name}{s.standardAttack.modifier ? ` ${s.standardAttack.modifier}` : ''}
          {#if s.standardAttack.range}<span class="dim">· {s.standardAttack.range}</span>{/if}
          {#if s.standardAttack.damage}<span class="dim">· {s.standardAttack.damage} {s.standardAttack.damageType}</span>{/if}
        </p>
      {/if}
      <FeatureList features={s.features ?? []} />
    </section>
  {/each}
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

  .stats {
    display: flex;
    flex-wrap: wrap;
    gap: 0.4rem;
    margin: 0;
  }

  .stats div {
    flex: 1 1 5rem;
    padding: 0.4rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--bg);
    text-align: center;
  }

  dt {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  dd {
    margin: 0.15rem 0 0;
    font-size: 0.95rem;
    font-weight: 600;
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

  .dim { color: var(--muted); }

  .features {
    padding-top: 0.5rem;
    border-top: 1px solid var(--line);
  }

  .key.spaced { margin-left: 0.6rem; }

  h4 {
    margin: 0;
    font-size: 0.95rem;
  }

  h4.group {
    margin-bottom: 0.4rem;
    font-size: 0.65rem;
    font-weight: normal;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .segment {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.6rem 0.7rem;
    border: 1px solid var(--line);
    border-left: 0.2rem solid var(--fear);
    border-radius: 6px;
    background: var(--bg);
  }
</style>
