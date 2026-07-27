<script>
  // Fear is GM-only and moves by hand — players roll the duality dice, not the GM,
  // so nothing in this app bumps the pool automatically.
  let { fear = 0, max = 12, busy = false, onadjust, onset } = $props()

  const pips = $derived(Array.from({ length: max }, (_, i) => i < fear))
</script>

<div class="fear">
  <div class="head">
    <span class="label">Fear</span>
    <span class="count">{fear} / {max}</span>
  </div>

  <div class="controls">
    <button
      class="btn ghost"
      onclick={() => onadjust?.(-1)}
      disabled={busy || fear <= 0}
      title="Spend a Fear"
    >−</button>

    <div class="pips" role="group" aria-label="Fear pool">
      {#each pips as filled, i (i)}
        <button
          class="pip"
          class:filled
          onclick={() => onset?.(filled && i + 1 === fear ? i : i + 1)}
          disabled={busy}
          title="Set Fear to {filled && i + 1 === fear ? i : i + 1}"
          aria-label="Set Fear to {filled && i + 1 === fear ? i : i + 1}"
        ></button>
      {/each}
    </div>

    <button
      class="btn ghost"
      onclick={() => onadjust?.(1)}
      disabled={busy || fear >= max}
      title="Gain a Fear"
    >+</button>
  </div>
</div>

<style>
  .fear {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    padding: 0.6rem 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--fear);
  }

  .count {
    font-size: 0.75rem;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }

  .controls {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .controls .btn {
    padding: 0.1rem 0.5rem;
    font-size: 1rem;
    line-height: 1.2;
  }

  .pips {
    display: flex;
    flex: 1;
    gap: 0.2rem;
  }

  .pip {
    flex: 1;
    height: 1.15rem;
    min-width: 0.5rem;
    padding: 0;
    border: 1px solid var(--line);
    border-radius: 3px;
    background: var(--bg);
    cursor: pointer;
  }

  .pip:hover:not(:disabled) { border-color: var(--fear); }

  .pip.filled {
    border-color: var(--fear);
    background: var(--fear);
  }

  .pip:disabled { cursor: default; }
</style>
