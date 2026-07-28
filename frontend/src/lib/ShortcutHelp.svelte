<script>
  import Modal from './Modal.svelte'
  import { SHORTCUTS, onKeys } from './keys.js'

  let open = $state(false)

  // '?' is bound app-wide rather than per pane, so the list is reachable from
  // anywhere — including the panes that have no shortcuts of their own. Modal
  // itself owns Escape and the capture-phase gate.
  $effect(() => onKeys(() => ({ '?': () => (open = !open) }), { whenModal: true }))
</script>

<button class="iconbtn" onclick={() => (open = true)} title="Keyboard shortcuts (?)" aria-label="Keyboard shortcuts">
  ⌨
</button>

<Modal bind:open label="Keyboard shortcuts" width="46rem">
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
</Modal>

<style>
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
