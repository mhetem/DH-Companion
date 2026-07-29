<script>
  import { SetGold, errorMessage } from './api.js'

  let { character, busy = false, onupdate } = $props()

  const PURSES = [
    ['handfuls', 'Handfuls'],
    ['bags', 'Bags'],
    ['chests', 'Chests']
  ]

  let saving = $state(false)
  let error = $state('')

  async function adjust(field, delta) {
    if (saving || busy) return
    saving = true
    try {
      const next = { ...character.gold, [field]: character.gold[field] + delta }
      onupdate?.(await SetGold(character.id, next.handfuls, next.bags, next.chests))
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }
</script>

<div class="gold">
  {#each PURSES as [field, label] (field)}
    <div class="purse">
      <span class="l">{label}</span>
      <div class="row">
        <button class="step" disabled={saving || busy} onclick={() => adjust(field, -1)} aria-label="One fewer {label}">−</button>
        <span class="n">{character.gold[field]}</span>
        <button class="step" disabled={saving || busy} onclick={() => adjust(field, 1)} aria-label="One more {label}">+</button>
      </div>
    </div>
  {/each}
</div>

{#if error}<p class="error">{error}</p>{/if}

<style>
  /* auto-fit rather than three fixed columns, so the row can wrap instead of pushing
     the last stepper outside a narrow card. */
  .gold {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(5.5rem, 1fr));
    gap: 0.5rem 0.35rem;
  }

  .purse {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.3rem;
    min-width: 0;
  }

  .l {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.04em;
    color: var(--muted);
  }

  .row {
    display: flex;
    align-items: center;
    gap: 0.3rem;
  }

  /* The compact square stepper ResourceTrack uses — .btn's padding is what made the
     row wider than its column. */
  .step {
    flex: none;
    width: 1.4rem;
    height: 1.4rem;
    padding: 0;
    border: 1px solid var(--line);
    border-radius: 5px;
    background: var(--panel-2);
    color: var(--text);
    font: inherit;
    line-height: 1;
    cursor: pointer;
  }

  .step:hover:not(:disabled) { border-color: var(--gold); }

  .step:disabled {
    cursor: default;
    opacity: 0.4;
  }

  .n {
    min-width: 1.4rem;
    font-size: 1.05rem;
    text-align: center;
    color: var(--gold);
  }

  .error {
    margin: 0.5rem 0 0;
    font-size: 0.78rem;
    color: var(--danger);
  }
</style>
