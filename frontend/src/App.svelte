<script>
  import { GetRole, GetUIScale, SetRole } from '../wailsjs/go/main/App.js'
  import { applyScale } from './lib/display.js'
  import RolePicker from './lib/RolePicker.svelte'
  import Header from './lib/Header.svelte'
  import GmShell from './lib/GmShell.svelte'
  import PlayerShell from './lib/PlayerShell.svelte'

  let role = $state('')      // '' until a role is chosen — first run shows the picker
  let loading = $state(true)
  let error = $state('')

  let scale = $state(100)

  // Load the last-used role before painting anything, so a returning user never
  // sees the picker flash — and the saved scale alongside it, so the first frame is
  // already at the right size rather than visibly resettling.
  Promise.all([GetRole(), GetUIScale()])
    .then(([savedRole, savedScale]) => {
      role = savedRole
      scale = savedScale
      applyScale(savedScale)
    })
    .catch((e) => (error = String(e)))
    .finally(() => (loading = false))

  async function pick(next) {
    const previous = role
    role = next
    try {
      await SetRole(next)
    } catch (e) {
      error = String(e)
      role = previous
    }
  }
</script>

{#if loading}
  <div class="splash">Loading…</div>
{:else if !role}
  <RolePicker onpick={pick} />
{:else}
  <div class="app">
    <Header
      {role}
      {scale}
      onswitch={pick}
      onscale={(next) => (scale = next)}
      onerror={(message) => (error = message)}
    />
    {#if role === 'gm'}
      <GmShell />
    {:else}
      <PlayerShell />
    {/if}
  </div>
{/if}

{#if error}
  <div class="error" role="alert">
    <span>{error}</span>
    <button onclick={() => (error = '')} aria-label="Dismiss">×</button>
  </div>
{/if}

<style>
  .app {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .splash {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: var(--muted);
  }

  .error {
    position: fixed;
    left: 50%;
    bottom: 1.25rem;
    transform: translateX(-50%);
    display: flex;
    align-items: center;
    gap: 0.75rem;
    max-width: 40rem;
    padding: 0.6rem 0.9rem;
    border: 1px solid var(--danger);
    border-radius: 8px;
    background: var(--panel-2);
    font-size: 0.85rem;
    text-align: left;
  }

  .error button {
    border: none;
    background: transparent;
    color: var(--muted);
    font-size: 1.1rem;
    line-height: 1;
    cursor: pointer;
  }
</style>
