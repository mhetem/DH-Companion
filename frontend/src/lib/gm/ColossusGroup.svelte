<script>
  import FeatureList from './FeatureList.svelte'
  import SegmentRow from './SegmentRow.svelte'

  // colossus is the framework row — it holds the shared Stress and is what gets
  // spotlit for the colossus's own actions. segments are the rows that point at it.
  let {
    colossus,
    segments = [],
    selectedId = null,
    busy = false,
    onmarkhp,
    onmarkstress,
    onspotlight,
    onbroken,
    onremove
  } = $props()

  let open = $state(false)

  const card = $derived(colossus.adversary)

  // By default a colossus falls when every segment is Destroyed. A segment printed
  // with no HP can't be Destroyed, so it doesn't count either way. Fatal and Link
  // features end the fight sooner, but they're the GM's call to read.
  const tracked = $derived(segments.filter((s) => s.hpMax > 0))
  const destroyed = $derived(tracked.filter((s) => s.hpMarked >= s.hpMax).length)
  const defeated = $derived(tracked.length > 0 && destroyed === tracked.length)

  let el = $state(null)
  $effect(() => {
    if (selectedId === colossus.id) el?.scrollIntoView({ block: 'nearest' })
  })
</script>

<li class="colossus" class:spotlit={colossus.spotlight} class:defeated class:unresolved={colossus.unresolved}>
  <div class="head" class:selected={selectedId === colossus.id} bind:this={el}>
    <button
      class="spot"
      class:on={colossus.spotlight}
      onclick={() => onspotlight?.(colossus, !colossus.spotlight)}
      disabled={busy}
      title={colossus.spotlight ? 'Remove from spotlight' : 'Spotlight the colossus as a whole, for its own actions'}
      aria-pressed={colossus.spotlight}
    >◆</button>

    <button class="who" onclick={() => (open = !open)} title="Show the colossus card">
      <span class="name">
        {colossus.displayName}
        {#if defeated}<span class="chip gone">Defeated</span>{/if}
      </span>
      <span class="meta">
        {#if colossus.unresolved}
          <span class="chip missing">missing card</span>
        {:else if card}
          Colossus · Tier {card.tier}
          {#if card.thresholdMinor || card.thresholdMajor}
            · Thresholds {card.thresholdMinor || '—'} / {card.thresholdMajor || '—'}
          {/if}
          {#if colossus.source === 'custom'}<span class="chip custom">custom</span>{/if}
        {/if}
      </span>
    </button>

    {#if tracked.length}
      <span class="tally" title="Segments Destroyed">{destroyed}/{tracked.length} destroyed</span>
    {/if}

    <div class="track">
      <span class="tlabel">Stress</span>
      <button class="btn ghost step" onclick={() => onmarkstress?.(colossus, -1)} disabled={busy || colossus.stressMarked <= 0}>−</button>
      <span class="value">{colossus.stressMarked}<span class="of">/{colossus.stressMax}</span></span>
      <button class="btn ghost step" onclick={() => onmarkstress?.(colossus, 1)} disabled={busy || colossus.stressMarked >= colossus.stressMax}>+</button>
    </div>

    <button class="btn danger" onclick={() => onremove?.(colossus)} disabled={busy} title="Remove the colossus and its segments from the fight">✕</button>
  </div>

  {#if open}
    <div class="detail">
      {#if colossus.unresolved}
        <p class="note">
          The card <code>{colossus.adversarySlug}</code> no longer exists. The segments keep
          their HP, but there is no stat block to show.
        </p>
      {:else if card}
        {#if card.description}<p class="desc">{card.description}</p>{/if}
        <dl class="stats">
          {#if card.size}<div><dt>Size</dt><dd>{card.size}</dd></div>{/if}
          {#if card.motives}<div><dt>Motives</dt><dd>{card.motives}</dd></div>{/if}
          {#if card.experiences}<div><dt>Experience</dt><dd>{card.experiences}</dd></div>{/if}
        </dl>
        <FeatureList features={card.features ?? []} />
      {/if}

      <ul class="rules">
        <li>Spotlight each segment as its own adversary; spotlight the colossus itself for its own actions.</li>
        <li>A PC on a segment has advantage attacking it, and that segment's standard attack can't target them.</li>
        <li>Leaping to an adjacent segment takes an action roll — usually Agility or Strength.</li>
        <li>Push and pull effects move the PC, not the colossus. PCs can start any number of Tag Team Rolls.</li>
        <li>An effect that targets the whole colossus uses its highest segment Difficulty.</li>
      </ul>
    </div>
  {/if}

  <ul class="segments">
    {#each segments as s (s.id)}
      <SegmentRow
        segment={s}
        selected={s.id === selectedId}
        {busy}
        onmarkhp={(d) => onmarkhp?.(s, d)}
        onspotlight={(on) => onspotlight?.(s, on)}
        onbroken={(on) => onbroken?.(s, on)}
      />
    {/each}
  </ul>
</li>

<style>
  .colossus {
    border: 1px solid var(--line);
    border-left: 0.2rem solid var(--fear);
    border-radius: 8px;
    margin-bottom: 0.4rem;
    background: var(--panel);
  }

  .colossus.spotlit { border-color: var(--fear); }
  .colossus.unresolved { border-color: var(--danger); }
  .colossus.defeated .head { opacity: 0.6; }

  .head {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.5rem 0.6rem;
    border-radius: 8px 8px 0 0;
  }

  .head.selected {
    box-shadow: inset 0 0 0 1px var(--gold), inset 0.2rem 0 0 var(--gold);
  }

  .colossus.defeated .head.selected { opacity: 1; }

  .spot {
    padding: 0.15rem 0.35rem;
    border: 1px solid transparent;
    border-radius: 6px;
    background: transparent;
    color: var(--line);
    font-size: 0.9rem;
    cursor: pointer;
  }

  .spot:hover:not(:disabled) { color: var(--fear); }
  .spot.on { color: var(--fear); }
  .spot:disabled { cursor: default; }

  .who {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.1rem;
    min-width: 0;
    padding: 0;
    border: none;
    background: none;
    color: inherit;
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .name {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    font-size: 0.95rem;
    font-weight: 600;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .chip.gone,
  .chip.missing {
    border-color: var(--danger);
    color: var(--danger);
  }

  .tally {
    font-size: 0.72rem;
    color: var(--muted);
    white-space: nowrap;
    font-variant-numeric: tabular-nums;
  }

  .track {
    display: flex;
    align-items: center;
    gap: 0.3rem;
  }

  .tlabel {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--muted);
  }

  .step {
    padding: 0.05rem 0.4rem;
    font-size: 0.95rem;
    line-height: 1.2;
  }

  .value {
    min-width: 2.6rem;
    font-size: 0.85rem;
    text-align: center;
    font-variant-numeric: tabular-nums;
  }

  .of {
    font-size: 0.7rem;
    color: var(--muted);
  }

  .detail {
    padding: 0.6rem 0.75rem 0.75rem;
    border-top: 1px solid var(--line);
  }

  .note,
  .desc {
    margin: 0 0 0.6rem;
    font-size: 0.8rem;
    line-height: 1.5;
    color: var(--muted);
  }

  code {
    padding: 0.05rem 0.3rem;
    border-radius: 4px;
    background: var(--bg);
    font-size: 0.75rem;
  }

  .stats {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem 1.25rem;
    margin: 0 0 0.6rem;
  }

  .stats div {
    display: flex;
    gap: 0.35rem;
  }

  dt {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--muted);
  }

  dd {
    margin: 0;
    font-size: 0.8rem;
  }

  .rules {
    margin: 0.6rem 0 0;
    padding: 0.5rem 0.75rem 0.5rem 1.6rem;
    border: 1px dashed var(--line);
    border-radius: 6px;
    font-size: 0.75rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .segments {
    margin: 0;
    padding: 0 0.5rem 0.3rem 1.5rem;
    list-style: none;
  }
</style>
