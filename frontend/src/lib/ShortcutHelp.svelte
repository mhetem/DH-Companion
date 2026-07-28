<script>
  import { fade } from 'svelte/transition'
  import { SHORTCUTS, modal, onKeys } from './keys.js'

  const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false

  let open = $state(false)
  let sheet = $state(null)

  // Focus follows the sheet, so the opener doesn't keep it and re-fire on space.
  $effect(() => {
    if (open) sheet?.focus()
  })

  // While the sheet is up it owns the keyboard — see the modal counter in keys.js.
  $effect(() => {
    if (!open) return
    modal.open()
    return () => modal.close()
  })

  // '?' is bound app-wide rather than per pane, so the list is reachable from
  // anywhere — including the panes that have no shortcuts of their own.
  $effect(() =>
    onKeys(
      () => ({
        '?': () => (open = !open),
        ...(open ? { Escape: () => (open = false) } : {})
      }),
      { whenModal: true }
    )
  )
</script>

<button class="opener" onclick={() => (open = true)} title="Keyboard shortcuts (?)" aria-label="Keyboard shortcuts">
  ⌨
</button>

{#if open}
  <!-- svelte-ignore a11y_click_events_have_key_events -->
  <div
    class="scrim"
    role="button"
    tabindex="-1"
    aria-label="Close"
    onclick={() => (open = false)}
    transition:fade={{ duration: reduced ? 0 : 120 }}
  >
    <div
      bind:this={sheet}
      class="sheet"
      role="dialog"
      aria-modal="true"
      aria-label="Keyboard shortcuts"
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
    >
      <header>
        <h2>Keyboard shortcuts</h2>
        <button class="btn ghost" onclick={() => (open = false)}>Close</button>
      </header>

      <div class="groups">
        {#each SHORTCUTS as group (group.scope)}
          <section>
            <h3>{group.scope}</h3>
            {#if group.note}<p class="note">{group.note}</p>{/if}
            <dl>
              {#each group.keys as row (row.key)}
                <div>
                  <dt><kbd>{row.key}</kbd></dt>
                  <dd>{row.label}</dd>
                </div>
              {/each}
            </dl>
          </section>
        {/each}
      </div>

      <p class="foot">Shortcuts stay out of the way while you are typing in a field.</p>
    </div>
  </div>
{/if}

<style>
  .opener {
    padding: 0.35rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.9rem;
    line-height: 1;
    cursor: pointer;
  }

  .opener:hover {
    border-color: var(--muted);
    color: var(--text);
  }

  .scrim {
    position: fixed;
    inset: 0;
    z-index: 40;
    display: grid;
    place-items: center;
    padding: 2rem 1rem;
    border: none;
    background: rgb(0 0 0 / 55%);
    cursor: default;
  }

  .sheet {
    display: flex;
    flex-direction: column;
    gap: 0.9rem;
    width: min(46rem, 100%);
    max-height: 100%;
    overflow-y: auto;
    padding: 1.25rem 1.4rem;
    border: 1px solid var(--line);
    border-radius: 12px;
    background: var(--panel);
    box-shadow: 0 18px 48px rgb(0 0 0 / 45%);
    text-align: left;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  h2 {
    margin: 0;
    font-size: 1.05rem;
  }

  h3 {
    margin: 0;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--gold);
  }

  .groups {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(16rem, 1fr));
    gap: 1.1rem;
    align-items: start;
  }

  .note {
    margin: 0.2rem 0 0;
    font-size: 0.72rem;
    color: var(--muted);
  }

  dl {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0.5rem 0 0;
  }

  dl div {
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
  }

  dt {
    flex: 0 0 4.5rem;
    text-align: right;
  }

  dd {
    margin: 0;
    font-size: 0.82rem;
    color: var(--muted);
  }

  .foot {
    margin: 0;
    font-size: 0.75rem;
    color: var(--muted);
    opacity: 0.8;
  }
</style>
