<script>
  import { untrack } from 'svelte'
  import HomebrewForm from './HomebrewForm.svelte'
  import LootDetail from './LootDetail.svelte'
  import {
    CreateCustomConsumable,
    CreateCustomItem,
    LOOT_RARITIES,
    UpdateCustomConsumable,
    UpdateCustomItem,
    errorMessage
  } from './api.js'

  // kind is 'items' or 'consumables' — the two share this form because the rows
  // are the same shape; only which service call to make differs.
  let { card = null, kind = 'items', onsaved, oncancel } = $props()

  const original = untrack(() => card)
  const editing = original !== null
  const isItem = $derived(kind === 'items')
  const noun = $derived(isItem ? 'item' : 'consumable')

  // The blank row is built from the kind this instance mounted with, the same way
  // the card is snapshotted — the parent hides the tabs while the form is up.
  const startedAsItem = untrack(() => kind === 'items')

  let form = $state(
    original
      ? { ...original }
      : {
          kind: startedAsItem ? 'item' : 'consumable',
          slug: '',
          name: '',
          type: startedAsItem ? 'Item' : 'Consumable',
          tier: '',
          description: '',
          rarity: 'Common',
          roll: 0,
          table: 'Homebrew'
        }
  )

  let saving = $state(false)
  let error = $state('')

  const previewCard = $derived({ ...form, source: 'custom' })

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      let saved
      if (editing) {
        saved = isItem ? await UpdateCustomItem(form) : await UpdateCustomConsumable(form)
      } else {
        saved = isItem ? await CreateCustomItem(form) : await CreateCustomConsumable(form)
      }
      error = ''
      onsaved?.(saved)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }
</script>

<HomebrewForm
  title={editing ? `Edit ${original.name}` : `New homebrew ${noun}`}
  formId="loot-form"
  {saving}
  {error}
  canSave={form.name.trim().length > 0}
  onsubmit={submit}
  {oncancel}
>
  {#snippet fields()}
    <section>
      <div class="row">
        <label class="grow">
          <span>Name</span>
          <input bind:value={form.name} placeholder="Thornwood Draught" required />
          {#if editing}
            <small>The slug stays <code>{original.slug}</code>.</small>
          {/if}
        </label>
        <label class="narrow">
          <span>Rarity</span>
          <select bind:value={form.rarity}>
            {#each LOOT_RARITIES as r (r.name)}
              <option value={r.name}>{r.name}</option>
            {/each}
          </select>
        </label>
      </div>

      <label>
        <span>Description</span>
        <textarea
          rows="5"
          bind:value={form.description}
          placeholder="What it does. Inline <strong> and <em> are rendered."
        ></textarea>
      </label>

      <!-- A roll number indexes a full 60-row table, which homebrew isn't part of,
           so there's no roll field here — the roller draws these by rarity. -->
      <small>
        Homebrew {noun}s have no roll number. The randomizer draws them by rarity when
        “Include homebrew” is ticked.
      </small>
    </section>
  {/snippet}

  {#snippet preview()}
    <LootDetail card={previewCard} />
  {/snippet}
</HomebrewForm>
