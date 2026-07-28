<script>
  import { fade } from 'svelte/transition'
  import { modal, onKeys } from './keys.js'

  const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false

  let { open = $bindable(false), label, width = '40rem', children } = $props()

  let sheet = $state(null)

  // While a sheet is up it owns the keyboard — modal owners listen in the capture
  // phase, so the panes underneath never see the keys. See keys.js.
  $effect(() => {
    if (!open) return
    modal.open()
    return () => modal.close()
  })

  $effect(() => {
    if (open) sheet?.focus()
  })

  $effect(() => onKeys(() => (open ? { Escape: () => (open = false) } : {}), { whenModal: true }))
</script>

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
      style="--sheet-width: {width}"
      role="dialog"
      aria-modal="true"
      aria-label={label}
      tabindex="-1"
      onclick={(e) => e.stopPropagation()}
    >
      {@render children?.()}
    </div>
  </div>
{/if}

<style>
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
    gap: 1rem;
    width: min(var(--sheet-width), 100%);
    max-height: 100%;
    overflow-y: auto;
    padding: 1.25rem 1.4rem;
    border: 1px solid var(--line);
    border-radius: 12px;
    background: var(--panel);
    box-shadow: 0 18px 48px rgb(0 0 0 / 45%);
    text-align: left;
    cursor: auto;
  }
</style>
