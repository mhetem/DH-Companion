<script>
  let { budget } = $props()

  // The bar tracks spend against the adjusted budget; anything past 100% is
  // clamped and the meter flips to the over-budget colour instead of overflowing.
  const percent = $derived(
    budget && budget.budget > 0 ? Math.min(100, Math.round((budget.spent / budget.budget) * 100)) : 0
  )
</script>

{#if budget}
  <div class="meter" class:over={budget.over}>
    <div class="numbers">
      <span class="spent">{budget.spent}</span>
      <span class="of">/ {budget.budget} battle points</span>
      <span class="remaining">
        {budget.over ? `${-budget.remaining} over` : `${budget.remaining} left`}
      </span>
    </div>

    <div class="track" role="presentation">
      <div class="fill" style:width="{percent}%"></div>
    </div>

    {#if budget.adjustments?.length}
      <ul class="adjustments">
        {#each budget.adjustments as adjustment (adjustment)}
          <li>{adjustment}</li>
        {/each}
      </ul>
    {/if}
  </div>
{:else}
  <p class="none">Attach a party to see the battle-point budget.</p>
{/if}

<style>
  .meter {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .meter.over { border-color: var(--danger); }

  .numbers {
    display: flex;
    align-items: baseline;
    gap: 0.4rem;
  }

  .spent {
    font-size: 1.5rem;
    font-weight: 600;
    line-height: 1;
  }

  .of {
    font-size: 0.8rem;
    color: var(--muted);
  }

  .remaining {
    margin-left: auto;
    font-size: 0.75rem;
    color: var(--ok);
  }

  .meter.over .remaining { color: var(--danger); }

  .track {
    height: 6px;
    border-radius: 999px;
    background: var(--bg);
    overflow: hidden;
  }

  .fill {
    height: 100%;
    background: var(--ok);
    transition: width 140ms ease;
  }

  .meter.over .fill { background: var(--danger); }

  .adjustments {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .adjustments li {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .none {
    margin: 0;
    padding: 0.75rem;
    border: 1px dashed var(--line);
    border-radius: 8px;
    font-size: 0.8rem;
    color: var(--muted);
  }
</style>
