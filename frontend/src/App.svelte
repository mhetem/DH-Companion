<script>
  import { GetUIScale } from '../wailsjs/go/main/App.js'
  import { applyScale } from './lib/display.js'
  import RolePicker from './lib/RolePicker.svelte'
  import Header from './lib/Header.svelte'
  import GmShell from './lib/GmShell.svelte'
  import PlayerShell from './lib/PlayerShell.svelte'

  // The role is never remembered — every launch opens on the picker. It's a view,
  // not an account, and starting from the choice keeps a shared machine honest.
  let role = $state('')
  let loading = $state(true)
  let error = $state('')

  let scale = $state(100)
  let reloadToken = $state(0)

  // The saved scale still loads before the first paint, so the picker itself is
  // already at the right size rather than visibly resettling.
  GetUIScale()
    .then((savedScale) => {
      scale = savedScale
      applyScale(savedScale)
    })
    .catch((e) => (error = String(e)))
    .finally(() => (loading = false))

  function pick(next) {
    role = next
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
      onreload={() => (reloadToken += 1)}
    />
    {#key reloadToken}
      {#if role === 'gm'}
        <GmShell />
      {:else}
        <PlayerShell />
      {/if}
    {/key}
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
