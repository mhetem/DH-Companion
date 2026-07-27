<script>
  import FeatureList from './FeatureList.svelte'

  let { combatant, busy = false, onmarkhp, onmarkstress, onspotlight, onremove } = $props()

  let open = $state(false)

  const card = $derived(combatant.adversary)
  const attack = $derived(card?.standardAttack)
  const down = $derived(combatant.hpMax > 0 && combatant.hpMarked >= combatant.hpMax)
</script>

<li class:spotlit={combatant.spotlight} class:down class:unresolved={combatant.unresolved}>
  <div class="row">
    <button
      class="spot"
      class:on={combatant.spotlight}
      onclick={() => onspotlight?.(!combatant.spotlight)}
      disabled={busy}
      title={combatant.spotlight ? 'Remove from spotlight' : 'Put in the spotlight'}
      aria-pressed={combatant.spotlight}
    >◆</button>

    <button class="who" onclick={() => (open = !open)} title="Show card details">
      <span class="name">{combatant.displayName}</span>
      <span class="meta">
        {#if combatant.unresolved}
          <span class="chip missing">missing card</span>
        {:else if card}
          Tier {card.tier} · {card.type}
          {#if combatant.source === 'custom'}<span class="chip custom">custom</span>{/if}
        {:else}
          ad-hoc
        {/if}
      </span>
    </button>

    <div class="track">
      <span class="tlabel">HP</span>
      <button class="btn ghost step" onclick={() => onmarkhp?.(-1)} disabled={busy || combatant.hpMarked <= 0}>−</button>
      <span class="value" class:critical={down}>{combatant.hpMarked}<span class="of">/{combatant.hpMax}</span></span>
      <button class="btn ghost step" onclick={() => onmarkhp?.(1)} disabled={busy || combatant.hpMarked >= combatant.hpMax}>+</button>
    </div>

    <div class="track">
      <span class="tlabel">Stress</span>
      <button class="btn ghost step" onclick={() => onmarkstress?.(-1)} disabled={busy || combatant.stressMarked <= 0}>−</button>
      <span class="value">{combatant.stressMarked}<span class="of">/{combatant.stressMax}</span></span>
      <button class="btn ghost step" onclick={() => onmarkstress?.(1)} disabled={busy || combatant.stressMarked >= combatant.stressMax}>+</button>
    </div>

    <button class="btn danger" onclick={() => onremove?.()} disabled={busy} title="Remove from the fight">✕</button>
  </div>

  {#if open}
    <div class="detail">
      {#if combatant.unresolved}
        <p class="note">
          The card <code>{combatant.adversarySlug}</code> no longer exists — its stats were
          spawned from a fallback. Edit the combatant to set HP and Stress by hand.
        </p>
      {:else if !card}
        <p class="note">An ad-hoc combatant with no source card.</p>
      {:else}
        {#if card.description}<p class="desc">{card.description}</p>{/if}
        <dl class="stats">
          {#if card.difficulty}<div><dt>Difficulty</dt><dd>{card.difficulty}</dd></div>{/if}
          {#if card.thresholdMinor}<div><dt>Thresholds</dt><dd>{card.thresholdMinor} / {card.thresholdMajor}</dd></div>{/if}
          {#if card.motives}<div><dt>Motives</dt><dd>{card.motives}</dd></div>{/if}
        </dl>
        {#if attack?.name}
          <p class="attack">
            <strong>{attack.name}</strong>
            <span class="muted">{attack.range} · {attack.modifier} · {attack.damage} {attack.damageType}</span>
          </p>
        {/if}
        <FeatureList features={card.features ?? []} />
      {/if}
    </div>
  {/if}
</li>

<style>
  li {
    border: 1px solid var(--line);
    border-radius: 8px;
    margin-bottom: 0.4rem;
    background: var(--panel);
  }

  li.spotlit { border-color: var(--fear); }
  li.unresolved { border-color: var(--danger); }
  li.down { opacity: 0.6; }

  .row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    padding: 0.5rem 0.6rem;
  }

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

  .name { font-size: 0.9rem; }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .chip.missing {
    border-color: var(--danger);
    color: var(--danger);
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

  .value.critical { color: var(--danger); }

  .of {
    font-size: 0.7rem;
    color: var(--muted);
  }

  .detail {
    padding: 0 0.75rem 0.75rem;
    border-top: 1px solid var(--line);
    margin-top: 0.1rem;
    padding-top: 0.6rem;
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

  .attack {
    display: flex;
    gap: 0.5rem;
    flex-wrap: wrap;
    margin: 0 0 0.75rem;
    font-size: 0.82rem;
  }

  .muted { color: var(--muted); }
</style>
