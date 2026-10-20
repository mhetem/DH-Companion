<script>
  import FeatureList from './FeatureList.svelte'

  let { segment, selected = false, busy = false, onmarkhp, onspotlight, onbroken } = $props()

  let open = $state(false)

  // seg is this segment's block on the colossus card, resolved by name. It goes
  // missing if the homebrew card was edited mid-fight and the segment renamed.
  const seg = $derived(segment.segment)
  const attack = $derived(seg?.standardAttack)
  const hasAttack = $derived(Boolean(attack?.name || attack?.damage))
  // A segment printed with no HP spawns with no track: it can't be Destroyed.
  const tracked = $derived(segment.hpMax > 0)
  const destroyed = $derived(tracked && segment.hpMarked >= segment.hpMax)

  let el = $state(null)
  $effect(() => {
    if (selected) el?.scrollIntoView({ block: 'nearest' })
  })
</script>

<li
  bind:this={el}
  class:spotlit={segment.spotlight}
  class:destroyed
  class:broken={segment.broken && !destroyed}
  class:selected
>
  <div class="row">
    <button
      class="spot"
      class:on={segment.spotlight}
      onclick={() => onspotlight?.(!segment.spotlight)}
      disabled={busy}
      title={segment.spotlight ? 'Remove from spotlight' : 'Put this segment in the spotlight'}
      aria-pressed={segment.spotlight}
    >◆</button>

    <button class="who" onclick={() => (open = !open)} title="Show segment details">
      <span class="name">
        {segment.displayName}
        {#if destroyed}<span class="chip gone">Destroyed</span>{/if}
      </span>
      <span class="meta">
        {#if !seg}
          <span class="chip missing">segment missing from the card</span>
        {:else}
          {#if seg.difficulty}Difficulty {seg.difficulty}{/if}
          {#if seg.difficulty && seg.adjacent}·{/if}
          {#if seg.adjacent}next to {seg.adjacent}{/if}
        {/if}
      </span>
    </button>

    <button
      class="brk"
      class:on={segment.broken}
      onclick={() => onbroken?.(!segment.broken)}
      disabled={busy}
      title={segment.broken
        ? 'Clear Broken'
        : 'Mark Broken — it can’t use actions or reactions until the condition is cleared'}
      aria-pressed={segment.broken}
    >{segment.broken ? 'Broken' : 'Break'}</button>

    {#if tracked}
      <div class="track">
        <span class="tlabel">HP</span>
        <button class="btn ghost step" onclick={() => onmarkhp?.(-1)} disabled={busy || segment.hpMarked <= 0}>−</button>
        <span class="value" class:critical={destroyed}>{segment.hpMarked}<span class="of">/{segment.hpMax}</span></span>
        <button class="btn ghost step" onclick={() => onmarkhp?.(1)} disabled={busy || destroyed}>+</button>
      </div>
    {:else}
      <span class="notrack" title="Printed with no HP — it can't be damaged or Destroyed">no HP</span>
    {/if}
  </div>

  {#if open}
    <div class="detail">
      {#if !seg}
        <p class="note">
          The card no longer has a segment called <code>{segment.segmentName}</code>. HP still
          tracks, but there is no stat block to show.
        </p>
      {:else}
        {#if hasAttack}
          <p class="attack">
            <strong>{attack.name}</strong>
            <span class="muted">{attack.range} · {attack.modifier} · {attack.damage} {attack.damageType}</span>
          </p>
        {:else}
          <p class="note">No standard attack.</p>
        {/if}
        <FeatureList features={seg.features ?? []} />
      {/if}
    </div>
  {/if}
</li>

<style>
  li {
    border: 1px solid var(--line);
    border-radius: 6px;
    margin-bottom: 0.3rem;
    background: var(--bg);
  }

  li.spotlit { border-color: var(--fear); }
  li.broken { border-style: dashed; }
  li.destroyed { opacity: 0.55; }

  li.selected {
    border-color: var(--gold);
    box-shadow: inset 0.2rem 0 0 var(--gold);
  }

  li.selected.destroyed { opacity: 1; }

  .row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.4rem 0.6rem;
  }

  .spot {
    padding: 0.15rem 0.35rem;
    border: 1px solid transparent;
    border-radius: 6px;
    background: transparent;
    color: var(--line);
    font-size: 0.8rem;
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
    font-size: 0.85rem;
  }

  .meta {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .chip.gone {
    border-color: var(--danger);
    color: var(--danger);
  }

  .chip.missing {
    border-color: var(--danger);
    color: var(--danger);
  }

  .brk {
    padding: 0.1rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.7rem;
    cursor: pointer;
  }

  .brk:hover:not(:disabled) { color: var(--text); }

  .brk.on {
    border-color: var(--gold);
    color: var(--gold);
  }

  .brk:disabled { cursor: default; }

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

  .value.critical { color: var(--danger); }

  .of {
    font-size: 0.7rem;
    color: var(--muted);
  }

  .notrack {
    min-width: 7.6rem;
    font-size: 0.7rem;
    text-align: center;
    color: var(--muted);
  }

  .detail {
    padding: 0.6rem 0.75rem 0.75rem;
    border-top: 1px solid var(--line);
  }

  .note {
    margin: 0 0 0.6rem;
    font-size: 0.8rem;
    line-height: 1.5;
    color: var(--muted);
  }

  code {
    padding: 0.05rem 0.3rem;
    border-radius: 4px;
    background: var(--panel);
    font-size: 0.75rem;
  }

  .attack {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin: 0 0 0.75rem;
    font-size: 0.82rem;
  }

  .muted { color: var(--muted); }
</style>
