<script>
  import SrdText from '../SrdText.svelte'
  import { LOOT_RARITIES, LOOT_TABLES, LookupLoot, RollLoot, TIERS, errorMessage } from './api.js'

  // One roller serves both sections because the backend call is the same: a haul
  // is up to four counts rolled at once. The two axes are genuinely different, so
  // mode decides which half is on show — equipment is drawn from a tier's tables,
  // while items and consumables have no tier and are rolled against rarity.
  let { mode = 'loot' } = $props()

  const isEquipment = $derived(mode === 'equipment')

  let tier = $state('1')
  let rarity = $state('Common')
  let dice = $state(1)
  let table = $state('')
  let includeHomebrew = $state(false)

  let weapons = $state(1)
  let armor = $state(0)
  let items = $state(1)
  let consumables = $state(1)

  let haul = $state(null)
  let error = $state('')
  let rolling = $state(false)

  // Hand-entry, for tables that roll their own d12s. It keeps its own table picker
  // rather than borrowing the one above: that one offers "Both books", and the same
  // total names a different entry in each, so there'd be nothing to resolve.
  let entryKind = $state('items')
  let entryTable = $state(LOOT_TABLES[0])
  let entryTotal = $state(1)

  const band = $derived(LOOT_RARITIES.find((r) => r.name === rarity) ?? LOOT_RARITIES[0])

  // The two dice counts belong to the rarity, so switching bands has to re-pick:
  // 1d12 is not on offer for a Legendary roll.
  function pickRarity(name) {
    rarity = name
    dice = LOOT_RARITIES.find((r) => r.name === name)?.minDice ?? 1
  }

  const total = $derived(
    isEquipment ? Number(weapons) + Number(armor) : Number(items) + Number(consumables)
  )

  // A total outside the band the chosen dice can reach is usually a typo or the
  // wrong rarity selected, so it's worth saying — but not worth refusing, since
  // the GM reading the dice is the authority on what was rolled.
  const outOfBand = $derived.by(() => {
    const n = Number(entryTotal)
    const d = Number(dice)
    if (!Number.isFinite(n) || n < d || n > d * 12) return `outside ${d}d12’s range (${d}–${d * 12})`
    return ''
  })

  function blankHaul() {
    return { tier: '', rarity, dice: Number(dice), weapons: [], armor: [], items: [], consumables: [] }
  }

  async function insert() {
    try {
      const roll = await LookupLoot({
        kind: entryKind === 'items' ? 'item' : 'consumable',
        table: entryTable,
        total: Number(entryTotal)
      })
      // Appends rather than replacing, so a run of hand-entered rolls builds up;
      // pressing Roll starts a fresh haul.
      const next = haul ?? blankHaul()
      next[entryKind] = [...next[entryKind], roll]
      haul = next
      error = ''
    } catch (e) {
      error = errorMessage(e)
    }
  }

  async function roll() {
    rolling = true
    try {
      haul = await RollLoot({
        tier,
        rarity,
        dice: Number(dice),
        table,
        includeHomebrew,
        weapons: isEquipment ? Number(weapons) : 0,
        armor: isEquipment ? Number(armor) : 0,
        items: isEquipment ? 0 : Number(items),
        consumables: isEquipment ? 0 : Number(consumables)
      })
      error = ''
    } catch (e) {
      haul = null
      error = errorMessage(e)
    } finally {
      rolling = false
    }
  }
</script>

<div class="roller">
  <div class="controls">
    {#if isEquipment}
      <label>
        <span>Party tier</span>
        <select bind:value={tier}>
          {#each TIERS as t (t)}
            <option value={t}>Tier {t}</option>
          {/each}
        </select>
      </label>
      <label>
        <span>Weapons</span>
        <input type="number" min="0" max="20" bind:value={weapons} />
      </label>
      <label>
        <span>Armor</span>
        <input type="number" min="0" max="20" bind:value={armor} />
      </label>
    {:else}
      <label>
        <span>Rarity</span>
        <select value={rarity} onchange={(e) => pickRarity(e.currentTarget.value)}>
          {#each LOOT_RARITIES as r (r.name)}
            <option value={r.name}>{r.name}</option>
          {/each}
        </select>
      </label>
      <label>
        <span>Dice</span>
        <select bind:value={dice}>
          <option value={band.minDice}>{band.minDice}d12</option>
          <option value={band.maxDice}>{band.maxDice}d12</option>
        </select>
      </label>
      <label>
        <span>Table</span>
        <select bind:value={table}>
          <option value="">Both books</option>
          {#each LOOT_TABLES as t (t)}
            <option value={t}>{t}</option>
          {/each}
        </select>
      </label>
      <label>
        <span>Items</span>
        <input type="number" min="0" max="20" bind:value={items} />
      </label>
      <label>
        <span>Consumables</span>
        <input type="number" min="0" max="20" bind:value={consumables} />
      </label>
    {/if}

    <!-- Off by default so a haul is reproducible from the SRD alone unless the GM
         asks for their own cards to be in the mix. -->
    <label class="check">
      <input type="checkbox" bind:checked={includeHomebrew} />
      <span>Include homebrew</span>
    </label>

    <button class="btn primary" onclick={roll} disabled={rolling || total < 1}>
      {rolling ? 'Rolling…' : 'Roll'}
    </button>
  </div>

  {#if !isEquipment}
    <div class="controls entry">
      <span class="lead">Rolled at the table?</span>
      <label>
        <span>Kind</span>
        <select bind:value={entryKind}>
          <option value="items">Item</option>
          <option value="consumables">Consumable</option>
        </select>
      </label>
      <label>
        <span>Table</span>
        <select bind:value={entryTable}>
          {#each LOOT_TABLES as t (t)}
            <option value={t}>{t}</option>
          {/each}
        </select>
      </label>
      <label>
        <span>Total</span>
        <input type="number" min="1" max="60" bind:value={entryTotal} />
      </label>
      <button class="btn" onclick={insert}>Add result</button>
      {#if outOfBand}
        <span class="hint">{outOfBand}</span>
      {/if}
    </div>
  {/if}

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if haul}
    <div class="results">
      <div class="resulthead">
        <button class="btn ghost" onclick={() => (haul = null)}>Clear</button>
      </div>
      {#if isEquipment}
        {#if haul.weapons?.length}
          <h4>Weapons</h4>
          <ul class="cards">
            {#each haul.weapons as weapon (weapon.slug)}
              <li>
                <span class="name">{weapon.name}</span>
                <span class="meta">
                  {weapon.category} · {weapon.trait} · {weapon.range} · {weapon.damage}
                  {weapon.damageType} · {weapon.burden}
                  {#if weapon.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
                </span>
                {#if weapon.feature}
                  <SrdText text={`<strong>${weapon.feature.title}:</strong> ${weapon.feature.description}`} />
                {/if}
              </li>
            {/each}
          </ul>
        {/if}

        {#if haul.armor?.length}
          <h4>Armor</h4>
          <ul class="cards">
            {#each haul.armor as piece (piece.slug)}
              <li>
                <span class="name">{piece.name}</span>
                <span class="meta">
                  Thresholds {piece.thresholdMajor} / {piece.thresholdSevere} · Score {piece.baseScore}
                  {#if piece.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
                </span>
                {#if piece.feature}
                  <SrdText text={`<strong>${piece.feature.title}:</strong> ${piece.feature.description}`} />
                {/if}
              </li>
            {/each}
          </ul>
        {/if}
      {:else}
        {#each [['Items', haul.items], ['Consumables', haul.consumables]] as [label, rolls] (label)}
          {#if rolls?.length}
            <h4>{label}</h4>
            <ul class="cards">
              <!-- Rolling the same total twice is a real result on these tables, so
                   the index is part of the key rather than the entry's slug. -->
              {#each rolls as entry, i (label + i)}
                <li>
                  <span class="name">{entry.entry.name}</span>
                  <span class="meta">
                    {#if entry.source === 'custom'}
                      {entry.entry.rarity}
                      <span class="chip custom">Homebrew</span>
                    {:else if entry.dice?.length}
                      {entry.dice.join(' + ')} = {entry.total} · {entry.entry.table} · {entry.entry.rarity}
                    {:else}
                      Rolled {entry.total} · {entry.entry.table} · {entry.entry.rarity}
                      <span class="chip">Entered</span>
                    {/if}
                  </span>
                  <SrdText text={entry.entry.description} />
                </li>
              {/each}
            </ul>
          {/if}
        {/each}
      {/if}
    </div>
  {/if}
</div>

<style>
  .roller {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .controls {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 0.6rem;
    padding: 0.75rem 1rem;
    border-bottom: 1px solid var(--line);
  }

  .controls label {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
  }

  .controls span {
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .controls input[type='number'] { width: 4.5rem; }

  .check {
    flex-direction: row;
    align-items: center;
    gap: 0.35rem;
    padding-bottom: 0.4rem;
  }

  .check span { text-transform: none; letter-spacing: 0; font-size: 0.78rem; }

  .entry {
    background: var(--panel);
  }

  /* Both sit directly inside .controls, which styles its spans as field labels —
     these are prose, so they opt back out. */
  .lead {
    padding-bottom: 0.4rem;
    font-size: 0.78rem;
    text-transform: none;
    letter-spacing: 0;
    color: var(--muted);
  }

  .hint {
    padding-bottom: 0.45rem;
    font-size: 0.72rem;
    font-style: italic;
    text-transform: none;
    letter-spacing: 0;
    color: var(--muted);
  }

  .resulthead {
    display: flex;
    justify-content: flex-end;
    margin-bottom: 0.5rem;
  }

  .error {
    margin: 0;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--danger);
    font-size: 0.8rem;
    color: var(--danger);
  }

  .results {
    flex: 1;
    max-width: 60rem;
    padding: 0.75rem 1rem 1.5rem;
    overflow-y: auto;
  }

  h4 {
    margin: 0.75rem 0 0.4rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  h4:first-child { margin-top: 0; }

  .cards {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .cards li {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    padding: 0.55rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--panel);
  }

  .name {
    font-size: 0.9rem;
    font-weight: 600;
  }

  .meta {
    display: flex;
    align-items: center;
    flex-wrap: wrap;
    gap: 0.3rem;
    font-size: 0.72rem;
    color: var(--muted);
  }
</style>
