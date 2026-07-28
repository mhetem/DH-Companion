<script>
  import lockup from '../assets/images/hilt-app-lockup.svg'

  let { onpick } = $props()
</script>

<div class="picker">
  <h1><img src={lockup} alt="Hilt" /></h1>
  <p class="tagline">Pick a side of the table. You can switch at any time — nothing is locked to a role.</p>

  <div class="cards">
    <button class="card gm" onclick={() => onpick('gm')}>
      <!-- Drawn rather than set as emoji: this machine has no emoji font, and a
           GTK build can't assume one — 🎲/🛡️ came out as tofu. -->
      <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M12 1.5 21.5 7v10L12 22.5 2.5 17V7Z" />
        <path d="M12 6.5 18.5 16.5H5.5Z" />
      </svg>
      <span class="title">Game Master</span>
      <span class="blurb">Build encounters, run combat, keep your campaign notes.</span>
    </button>

    <button class="card player" onclick={() => onpick('player')}>
      <svg class="icon" viewBox="0 0 24 24" aria-hidden="true">
        <path d="M12 1.8 20 5.4v5.8c0 5.2-3.3 9-8 11-4.7-2-8-5.8-8-11V5.4Z" />
        <path d="M12 7v9" />
      </svg>
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
    /* Lifts the lockup off the flat field without introducing another surface. */
    background: radial-gradient(ellipse at 50% 32%, var(--panel) 0%, var(--bg) 60%);
  }

  h1 {
    margin: 0;
    line-height: 0;
  }

  h1 img {
    display: block;
    height: 7rem;
  }

  .tagline {
    /* The lockup's own dead space already reads as a gap, so this sits tighter. */
    margin: 0.25rem 0 2.5rem;
    max-width: 40rem;
    color: var(--muted);
    text-align: center;
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
    transition: transform 120ms ease, border-color 120ms ease, box-shadow 120ms ease;
  }

  .card:hover {
    transform: translateY(-3px);
  }

  .card.gm:hover {
    border-color: var(--fear);
    box-shadow: 0 0.5rem 1.5rem -0.5rem var(--fear);
  }

  .card.player:hover {
    border-color: var(--hope);
    box-shadow: 0 0.5rem 1.5rem -0.5rem var(--hope);
  }

  @media (prefers-reduced-motion: reduce) {
    .card { transition: border-color 120ms ease; }
    .card:hover { transform: none; }
  }

  .icon {
    width: 2.5rem;
    height: 2.5rem;
    fill: none;
    stroke: currentColor;
    stroke-width: 1.5;
    stroke-linejoin: round;
  }

  .card.gm .icon { color: var(--fear); }
  .card.player .icon { color: var(--hope); }

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
