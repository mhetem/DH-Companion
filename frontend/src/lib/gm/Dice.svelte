<script>
  import { Damage, GM, Sizes, errorMessage } from './api.js'

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

  function record(entry) {
    log = [{ ...entry, at: new Date() }, ...log].slice(0, 12)
  }

  async function rollD20() {
    try {
      const roll = await GM(advantage, disadvantage, Number(modifier))
      error = ''
      record({
        kind: 'd20',
        headline: String(roll.Result),
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
        headline: String(roll.total),
        note: '',
        detail: `${roll.count}d${roll.sides}${signed(roll.modifier)}`
      })
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

<div class="dice">
  <header>
    <h2>Dice</h2>
    <p class="blurb">
      A d20 for GM rolls and a damage roller. Duality dice stay on the player side — the
      combat runner will wire this to Fear in phase 3.
    </p>
  </header>

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
        </label>
        <label class="check">
          <input type="checkbox" bind:checked={disadvantage} />
          <span>Disadvantage</span>
        </label>
        <label class="narrow">
          <span>Modifier</span>
          <input type="number" bind:value={modifier} />
        </label>
      </div>
      {#if advantage && disadvantage}
        <p class="hint">Advantage and disadvantage cancel — this rolls flat.</p>
      {/if}
      <button class="btn primary" onclick={rollD20}>Roll {describeD20()}</button>
    </section>

    <section>
      <h3>Damage</h3>
      <div class="controls">
        <label class="narrow">
          <span>Dice</span>
          <input type="number" min="1" max="20" bind:value={count} />
        </label>
        <label class="narrow">
          <span>Size</span>
          <select bind:value={sides}>
            {#each dieSizes as size (size)}
              <option value={size}>d{size}</option>
            {/each}
          </select>
        </label>
        <label class="narrow">
          <span>Modifier</span>
          <input type="number" bind:value={damageModifier} />
        </label>
      </div>
      <p class="hint">The modifier lands once on the total, not on each die.</p>
      <button class="btn primary" onclick={rollDamage} disabled={!dieSizes.length}>
        Roll {count}d{sides}{signed(damageModifier)}
      </button>
    </section>
  </div>

  <section class="log">
    <h3>Recent rolls</h3>
    {#if !log.length}
      <p class="empty">Nothing rolled yet.</p>
    {:else}
      <ul>
        {#each log as entry (entry.at.getTime() + entry.headline)}
          <li>
            <span class="result" class:crit={Boolean(entry.note)}>{entry.headline}</span>
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

  .log { margin-top: 1.25rem; }

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

  .result {
    min-width: 2.5rem;
    font-size: 1.15rem;
    font-weight: 600;
    text-align: right;
  }

  .result.crit { color: var(--fear); }

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

  .empty {
    margin: 0.5rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }
</style>
