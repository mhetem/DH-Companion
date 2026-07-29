<script>
  // A pip track. Marked counts up, matching the GM runner's HP/Stress display —
  // the database has always stored marked-ascending on both sides.
  let { label, marked = 0, max = 0, tone = 'plain', busy = false, onchange } = $props()

  const pips = $derived(Array.from({ length: Math.max(0, max) }, (_, i) => i < marked))

  function set(index) {
    // Clicking the pip you're already at clears down to it instead of being a no-op.
    onchange?.(index + 1 === marked ? index : index + 1)
  }
</script>

<div class="track {tone}">
  <div class="head">
    <span class="label">{label}</span>
    <span class="count">{marked}/{max}</span>
  </div>
  <div class="pips">
    <button class="step" disabled={busy || marked <= 0} onclick={() => onchange?.(marked - 1)} aria-label="Clear one {label}">−</button>
    <div class="row">
      {#each pips as filled, i (i)}
        <button
          class="pip"
          class:filled
          disabled={busy}
          onclick={() => set(i)}
          aria-label="{label} {i + 1}"
        ></button>
      {/each}
      {#if !pips.length}
        <span class="none">none</span>
      {/if}
    </div>
    <button class="step" disabled={busy || marked >= max} onclick={() => onchange?.(marked + 1)} aria-label="Mark one {label}">+</button>
  </div>
</div>

<style>
  .track {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    --tone: var(--muted);
  }

  .track.hope { --tone: var(--hope); }
  .track.danger { --tone: var(--danger); }
  .track.stress { --tone: var(--fear); }
  .track.armor { --tone: var(--gold); }

  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .count {
    font-size: 0.78rem;
    color: var(--tone);
  }

  .pips {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .row {
    display: flex;
    flex: 1;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .pip {
    width: 1.1rem;
    height: 1.1rem;
    padding: 0;
    border: 1px solid var(--line);
    border-radius: 4px;
    background: var(--panel-2);
    cursor: pointer;
  }

  .pip:hover:not(:disabled) { border-color: var(--tone); }

  .pip.filled {
    border-color: var(--tone);
    background: var(--tone);
  }

  .pip:disabled { cursor: default; }

  .step {
    width: 1.4rem;
    height: 1.4rem;
    padding: 0;
    border: 1px solid var(--line);
    border-radius: 5px;
    background: var(--panel-2);
    color: var(--text);
    font: inherit;
    line-height: 1;
    cursor: pointer;
  }

  .step:hover:not(:disabled) { border-color: var(--tone); }

  .step:disabled {
    cursor: default;
    opacity: 0.4;
  }

  .none {
    font-size: 0.75rem;
    color: var(--muted);
  }
</style>
