<script>
  let { title, children } = $props()
</script>

<div class="state">
  <!-- The hilt on its own, lifted out of the logo without its field and grip: the
       full mark is a dark tile with a gold frame and turns to mush at this size. -->
  <svg class="mark" viewBox="42 29 46 46" aria-hidden="true">
    <rect x="61" y="40" width="8" height="32" rx="3" fill="var(--gold)" />
    <path d="M46 47 Q65 39 84 47 Q65 54 46 47 Z" fill="var(--gold)" />
    <circle cx="65" cy="37" r="5.5" fill="var(--danger)" stroke="var(--hope)" stroke-width="1.3" />
  </svg>
  <p class="title">{title}</p>
  {#if children}
    <p class="hint">{@render children()}</p>
  {/if}
</div>

<style>
  /* The full-pane "nothing here yet" case, as opposed to the one-line .empty used
     inside a list that already has a heading above it. Deliberately not flex:1 —
     panes lay themselves out differently, and a centred block under the pane
     header reads well in all of them without any of them having to cooperate. */
  .state {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.35rem;
    padding: 3.5rem 1rem;
    text-align: center;
  }

  .mark {
    width: 2.75rem;
    height: 2.75rem;
    margin-bottom: 0.5rem;
    opacity: 0.6;
  }

  .title {
    margin: 0;
    font-size: 0.95rem;
    color: var(--text);
  }

  .hint {
    margin: 0;
    max-width: 26rem;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }
</style>
