<script>
  import { fade, fly, scale } from 'svelte/transition'
  import { Damage, GM, Sizes, errorMessage } from './api.js'
  import { onKeys } from '../keys.js'
  import RollResult from './RollResult.svelte'

  const reduced = window.matchMedia?.('(prefers-reduced-motion: reduce)').matches ?? false

  // compact drops the page chrome and stacks the rollers so the same component can
  // sit in the combat runner's rail. The roll logic lives here only.
  let { compact = false } = $props()

  // GMDice and DualityDice have no json tags on the Go side, so their fields
  // arrive capitalised — Result and Msg, not result and msg.
  let advantage = $state(false)
  let disadvantage = $state(false)
  let modifier = $state(0)

  let count = $state(1)
  let sides = $state(6)
  let damageModifier = $state(0)

  let error = $state('')
  let log = $state([])

  // The rollable dice come from internal/dice rather than a list duplicated
  // here, so the picker can't offer a die the backend would reject.
  let dieSizes = $state([])
  Sizes()
    .then((sizes) => (dieSizes = sizes ?? []))
    .catch((e) => (error = errorMessage(e)))

  let nextId = 0

  // The result flashes centre-screen long enough to read, then clears itself. Miss it
  // and it's still in the log below.
  const FLASH_MS = 2200
  let flash = $state(null)
  let flashTimer

  function record(entry) {
    const item = { ...entry, id: nextId++, at: new Date() }
    log = [item, ...log].slice(0, 12)
    flash = item
    clearTimeout(flashTimer)
    flashTimer = setTimeout(() => (flash = null), FLASH_MS)
  }

  function dismiss() {
    clearTimeout(flashTimer)
    flash = null
  }

  $effect(() => () => clearTimeout(flashTimer))

  async function rollD20() {
    try {
      const roll = await GM(advantage, disadvantage, Number(modifier))
      error = ''
      record({
        kind: 'd20',
        value: Number(roll.Result),
        max: 20 + Number(modifier),
        note: roll.Msg,
        detail: describeD20()
      })
    } catch (e) {
      error = errorMessage(e)
    }
  }

  function signed(n) {
    const value = Number(n)
    if (!value) return ''
    return value > 0 ? `+${value}` : String(value)
  }

  function describeD20() {
    const parts = ['d20']
    if (advantage && !disadvantage) parts.push('advantage')
    if (disadvantage && !advantage) parts.push('disadvantage')
    if (advantage && disadvantage) parts.push('adv + dis cancel')
    const mod = signed(modifier)
    if (mod) parts.push(mod)
    return parts.join(' · ')
  }

  async function rollDamage() {
    try {
      const roll = await Damage(Number(count), Number(sides), Number(damageModifier))
      error = ''
      record({
        kind: 'damage',
        value: Number(roll.total),
        max: Number(roll.count) * Number(roll.sides) + Number(roll.modifier),
        note: '',
        detail: `${roll.count}d${roll.sides}${signed(roll.modifier)}`
      })
    } catch (e) {
      error = errorMessage(e)
    }
  }

  // Escape is only claimed while a flash is up, so it stays free for the runner's
  // "drop the selection" when both are mounted and nothing has been rolled.
  $effect(() =>
    onKeys(() => ({
      r: rollD20,
      d: rollDamage,
      a: () => (advantage = !advantage),
      z: () => (disadvantage = !disadvantage),
      ...Object.fromEntries(dieSizes.map((size, i) => [String(i + 1), () => (sides = size)])),
      ...(flash ? { Escape: dismiss } : {})
    }))
  )

  const sizeKey = (size) => {
    const i = dieSizes.indexOf(size)
    return i >= 0 && i < 9 ? String(i + 1) : ''
  }
</script>

<div class="dice" class:compact>
  {#if !compact}
    <header>
      <h2>Dice</h2>
      <p class="blurb">
        A d20 for GM rolls and a damage roller. Duality dice stay on the player side —
        only players roll Hope and Fear.
      </p>
    </header>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <div class="rollers">
    <section>
      <h3>GM roll (d20)</h3>
      <div class="controls">
        <label class="check">
          <input type="checkbox" bind:checked={advantage} />
          <span>Advantage</span>
          <kbd>a</kbd>
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={disadvantage} />
          <span>Disadvantage</span>
          <kbd>z</kbd>
        </label>
        <label class="narrow">
          <span>Modifier</span>
          <input type="number" bind:value={modifier} />
        </label>
      </div>
      {#if advantage && disadvantage}
        <p class="hint">Advantage and disadvantage cancel — this rolls flat.</p>
      {/if}
      <button class="btn primary" onclick={rollD20}>Roll {describeD20()} <kbd>r</kbd></button>
    </section>

    <section>
      <h3>Damage</h3>
      <div class="controls">
        <label class="narrow">
          <span>Dice</span>
          <input type="number" min="1" max="20" bind:value={count} />
        </label>
        <label class="narrow">
          <span>Modifier</span>
          <input type="number" bind:value={damageModifier} />
        </label>
      </div>

      <div class="sizes" role="group" aria-label="Die size">
        <span class="slabel">Size</span>
        <div class="chips">
          {#each dieSizes as size (size)}
            <button
              class="die"
              class:on={sides === size}
              onclick={() => (sides = size)}
              aria-pressed={sides === size}
              title={sizeKey(size) ? `d${size} — press ${sizeKey(size)}` : `d${size}`}
            >
              d{size}
            </button>
          {/each}
        </div>
      </div>
      <p class="hint">The modifier lands once on the total, not on each die.</p>
      <button class="btn primary" onclick={rollDamage} disabled={!dieSizes.length}>
        Roll {count}d{sides}{signed(damageModifier)} <kbd>d</kbd>
      </button>
    </section>
  </div>

  <section class="log">
    <h3>Recent rolls</h3>
    {#if !log.length}
      <p class="empty">Nothing rolled yet.</p>
    {:else}
      <ul>
        {#each log as entry (entry.id)}
          <li in:fly={{ y: -8, duration: reduced ? 0 : 180 }}>
            <RollResult
              value={entry.value}
              max={entry.max}
              crit={Boolean(entry.note)}
              {compact}
            />
            <span class="detail">
              {entry.detail}
              {#if entry.note}<strong class="note">{entry.note}</strong>{/if}
            </span>
            <span class="time">{entry.at.toLocaleTimeString()}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </section>

  {#if flash}
    <div class="flash" role="status" aria-live="polite">
      {#key flash.id}
        <button
          class="card"
          class:crit={Boolean(flash.note)}
          onclick={dismiss}
          title="Dismiss"
          in:scale={{ start: 0.88, duration: reduced ? 0 : 170 }}
          out:fade={{ duration: reduced ? 0 : 220 }}
        >
          <span class="fdetail">{flash.detail}</span>
          <RollResult value={flash.value} max={flash.max} crit={Boolean(flash.note)} big />
          {#if flash.note}<span class="fnote">{flash.note}</span>{/if}
        </button>
      {/key}
    </div>
  {/if}
</div>

<style>
  .dice {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header { margin-bottom: 1rem; }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .blurb {
    margin: 0.25rem 0 0;
    max-width: 38rem;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .rollers {
    display: flex;
    flex-wrap: wrap;
    gap: 0.75rem;
  }

  .rollers section {
    display: flex;
    flex: 1 1 18rem;
    flex-direction: column;
    gap: 0.6rem;
    align-items: flex-start;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  h3 {
    margin: 0;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  .controls {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 0.6rem;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.narrow { flex: 0 0 5rem; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .check {
    flex-direction: row;
    align-items: center;
    gap: 0.35rem;
    padding-bottom: 0.35rem;
    cursor: pointer;
  }

  .check span { text-transform: none; letter-spacing: 0; }

  .hint {
    margin: 0;
    font-size: 0.72rem;
    color: var(--muted);
    opacity: 0.85;
  }

  .sizes {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    width: 100%;
  }

  .slabel {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .die {
    padding: 0.2rem 0.55rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--bg);
    color: var(--muted);
    font: inherit;
    font-size: 0.8rem;
    font-variant-numeric: tabular-nums;
    cursor: pointer;
  }

  .die:hover:not(.on) {
    border-color: var(--muted);
    color: var(--text);
  }

  .die.on {
    border-color: var(--fear);
    background: var(--fear);
    color: var(--bg);
  }

  .log { margin-top: 1.25rem; }

  /* Rail-sized: no page padding or scroll of its own, rollers stacked, log trimmed. */
  .dice.compact {
    flex: none;
    padding: 0;
    overflow: visible;
  }

  .dice.compact .rollers { flex-direction: column; }

  /* Nested inside the rail's panel, so the inner cards step up a shade. */
  .dice.compact .rollers section {
    flex: none;
    padding: 0.7rem 0.75rem;
    background: var(--panel-2);
  }

  .dice.compact .log li { background: var(--panel-2); }

  .dice.compact .log { margin-top: 0.75rem; }

  .dice.compact .log li:nth-child(n + 5) { display: none; }

  .dice.compact .time { display: none; }

  .log ul {
    margin: 0.5rem 0 0;
    padding: 0;
    list-style: none;
  }

  .log li {
    display: flex;
    align-items: baseline;
    gap: 0.6rem;
    padding: 0.4rem 0.6rem;
    margin-bottom: 0.3rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel);
  }

  .detail {
    flex: 1;
    font-size: 0.78rem;
    color: var(--muted);
  }

  .note {
    margin-left: 0.35rem;
    color: var(--fear);
  }

  .time {
    font-size: 0.7rem;
    color: var(--muted);
    opacity: 0.7;
  }

  .empty { margin-top: 0.5rem; }

  /* Fixed to the viewport so it centres over the whole window, and transparent to
     clicks so a stray roll never blocks the runner underneath. */
  .flash {
    position: fixed;
    inset: 0;
    display: grid;
    place-items: center;
    z-index: 20;
    pointer-events: none;
  }

  .card {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.35rem;
    padding: 1.25rem 2.25rem;
    border: 1px solid var(--line);
    border-radius: 14px;
    background: var(--panel);
    color: var(--text);
    font: inherit;
    box-shadow: 0 18px 48px rgb(0 0 0 / 45%);
    pointer-events: auto;
    cursor: pointer;
  }

  .card.crit { border-color: var(--fear); }

  .fdetail {
    font-size: 0.78rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  .fnote {
    font-size: 0.85rem;
    font-weight: 600;
    color: var(--fear);
  }
</style>
