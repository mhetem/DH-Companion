<script>
  let { onpick } = $props()
</script>

<div class="picker">
  <h1>DH Companion</h1>
  <p class="tagline">Pick a side of the table. You can switch at any time — nothing is locked to a role.</p>

  <div class="cards">
    <button class="card gm" onclick={() => onpick('gm')}>
      <span class="icon">🎲</span>
      <span class="title">Game Master</span>
      <span class="blurb">Build encounters, run combat, keep your campaign notes.</span>
    </button>

    <button class="card player" onclick={() => onpick('player')}>
      <span class="icon">🛡️</span>
      <span class="title">Player</span>
      <span class="blurb">Your character sheet, domain cards, and dice.</span>
    </button>
  </div>
</div>

<style>
  .picker {
    display: flex;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    height: 100%;
    padding: 2rem;
  }

  h1 {
    margin: 0;
    font-size: 2.5rem;
    letter-spacing: 0.02em;
  }

  .tagline {
    margin: 0.75rem 0 2.5rem;
    max-width: 34rem;
    color: var(--muted);
  }

  .cards {
    display: flex;
    gap: 1.5rem;
    flex-wrap: wrap;
    justify-content: center;
  }

  .card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.5rem;
    width: 15rem;
    padding: 2rem 1.5rem;
    border: 1px solid var(--line);
    border-radius: 12px;
    background: var(--panel);
    color: inherit;
    font: inherit;
    cursor: pointer;
    transition: transform 120ms ease, border-color 120ms ease;
  }

  .card:hover {
    transform: translateY(-3px);
  }

  .card.gm:hover { border-color: var(--fear); }
  .card.player:hover { border-color: var(--hope); }

  .icon { font-size: 2rem; }

  .title {
    font-size: 1.15rem;
    font-weight: 600;
  }

  .blurb {
    font-size: 0.85rem;
    color: var(--muted);
    line-height: 1.4;
  }
</style>
