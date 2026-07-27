<script>
  // The roll is already resolved by the time this mounts — the tumble is theatre over
  // a known outcome, never the source of the number. max only sets the flicker range.
  let { value = 0, max = 20, crit = false, compact = false, big = false, animate = true } = $props()

  const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false
  const still = !animate || reduced

  let shown = $state(still ? value : 1)
  let settled = $state(still)

  $effect(() => {
    if (still) return
    const duration = compact ? 260 : 420
    const ceiling = Math.max(2, Math.round(max))
    const started = performance.now()
    const id = setInterval(() => {
      if (performance.now() - started >= duration) {
        shown = value
        settled = true
        clearInterval(id)
        return
      }
      shown = 1 + Math.floor(Math.random() * ceiling)
    }, 45)
    return () => clearInterval(id)
  })
</script>

<span class="result" class:crit={crit && settled} class:rolling={!settled} class:compact class:big>{shown}</span>

<style>
  .result {
    display: inline-block;
    min-width: 2.5rem;
    font-size: 1.15rem;
    font-weight: 600;
    text-align: right;
    /* Fixed-width digits so the number doesn't jitter while it tumbles. */
    font-variant-numeric: tabular-nums;
  }

  .result.compact {
    min-width: 2rem;
    font-size: 1rem;
  }

  .result.big {
    min-width: 0;
    font-size: 4.5rem;
    line-height: 1.05;
    text-align: center;
  }

  .result.rolling {
    color: var(--muted);
    opacity: 0.75;
  }

  .result.crit {
    color: var(--fear);
    animation: pop 420ms ease-out;
  }

  @keyframes pop {
    0% { transform: scale(1); }
    35% { transform: scale(1.28); }
    100% { transform: scale(1); }
  }

  @media (prefers-reduced-motion: reduce) {
    .result.crit { animation: none; }
  }
</style>
