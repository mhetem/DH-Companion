<script>
  let { sections } = $props()

  let activeId = $state('')
  const active = $derived(sections.find((s) => s.id === activeId) ?? sections[0])
</script>

<div class="shell">
  <nav>
    {#each sections as section (section.id)}
      <button class:active={section.id === active.id} onclick={() => (activeId = section.id)}>
        {section.label}
      </button>
    {/each}
  </nav>

  <main>
    <h2>{active.label}</h2>
    <p class="blurb">{active.blurb}</p>
    <p class="phase">Coming in phase {active.phase}</p>
  </main>
</div>

<style>
  .shell {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  nav {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    width: 12rem;
    padding: 0.75rem;
    border-right: 1px solid var(--line);
    background: var(--panel);
  }

  nav button {
    padding: 0.45rem 0.6rem;
    border: none;
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.9rem;
    text-align: left;
    cursor: pointer;
  }

  nav button:hover { color: var(--text); }

  nav button.active {
    background: var(--panel-2);
    color: var(--text);
  }

  main {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: center;
    justify-content: center;
    gap: 0.5rem;
    padding: 2rem;
    text-align: center;
  }

  h2 {
    margin: 0;
    font-size: 1.4rem;
  }

  .blurb {
    margin: 0;
    max-width: 30rem;
    color: var(--muted);
  }

  .phase {
    margin: 0.5rem 0 0;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
    opacity: 0.7;
  }
</style>
