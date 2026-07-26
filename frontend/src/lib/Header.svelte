<script>
  let { role, onswitch } = $props()

  const other = $derived(role === 'gm' ? 'player' : 'gm')
  const otherLabel = $derived(other === 'gm' ? 'Game Master' : 'Player')
</script>

<header>
  <span class="brand">DH Companion</span>
  <span class="badge" class:gm={role === 'gm'} class:player={role === 'player'}>
    {role === 'gm' ? 'Game Master' : 'Player'}
  </span>
  <button class="switch" onclick={() => onswitch(other)}>
    Switch to {otherLabel}
  </button>
</header>

<style>
  header {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.6rem 1rem;
    border-bottom: 1px solid var(--line);
    background: var(--panel);
  }

  .brand {
    font-weight: 600;
    letter-spacing: 0.02em;
  }

  .badge {
    padding: 0.15rem 0.6rem;
    border-radius: 999px;
    border: 1px solid var(--line);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--muted);
  }

  .badge.gm { border-color: var(--fear); color: var(--fear); }
  .badge.player { border-color: var(--hope); color: var(--hope); }

  .switch {
    margin-left: auto;
    padding: 0.35rem 0.8rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .switch:hover {
    color: var(--text);
    border-color: var(--muted);
  }
</style>
