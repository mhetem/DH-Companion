<script>
  import { untrack } from 'svelte'
  import HomebrewForm from './HomebrewForm.svelte'
  import ArmorDetail from './ArmorDetail.svelte'
  import { CreateCustomArmor, MAX_ARMOR_SCORE, TIERS, UpdateCustomArmor, errorMessage } from './api.js'

  let { card = null, onsaved, oncancel } = $props()

  const original = untrack(() => card)
  const editing = original !== null

  let form = $state(
    original
      ? { ...original, feature: original.feature ? { ...original.feature } : null }
      : {
          kind: 'armor',
          slug: '',
          name: '',
          tier: '1',
          type: 'Armor',
          description: '',
          thresholdMajor: 6,
          thresholdSevere: 13,
          baseScore: 3,
          feature: null
        }
  )

  let saving = $state(false)
  let error = $state('')

  const previewCard = $derived({ ...form, source: 'custom' })

  function toggleFeature() {
    form.feature = form.feature ? null : { title: '', description: '' }
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      // The number inputs hand back strings; the Go side wants ints.
      const payload = {
        ...form,
        thresholdMajor: Number(form.thresholdMajor),
        thresholdSevere: Number(form.thresholdSevere),
        baseScore: Number(form.baseScore)
      }
      const saved = editing ? await UpdateCustomArmor(payload) : await CreateCustomArmor(payload)
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
  title={editing ? `Edit ${original.name}` : 'New homebrew armor'}
  formId="armor-form"
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
          <input bind:value={form.name} placeholder="Thornwood Cuirass" required />
          {#if editing}
            <small>The slug stays <code>{original.slug}</code>.</small>
          {/if}
        </label>
        <label class="narrow">
          <span>Tier</span>
          <select bind:value={form.tier}>
            {#each TIERS as tier (tier)}
              <option value={tier}>Tier {tier}</option>
            {/each}
          </select>
        </label>
      </div>

      <div class="row">
        <label>
          <span>Major threshold</span>
          <input type="number" min="0" bind:value={form.thresholdMajor} />
        </label>
        <label>
          <span>Severe threshold</span>
          <input type="number" min="0" bind:value={form.thresholdSevere} />
        </label>
        <label>
          <span>Base Armor Score</span>
          <input type="number" min="0" max={MAX_ARMOR_SCORE} bind:value={form.baseScore} />
        </label>
      </div>
      <small>Thresholds are the base pair, before the wearer’s level is added.</small>

      <label>
        <span>Description</span>
        <textarea rows="2" bind:value={form.description} placeholder="Optional flavour text."></textarea>
      </label>
    </section>

    <section>
      <h3>Feature</h3>
      {#if form.feature}
        <label>
          <span>Title</span>
          <input bind:value={form.feature.title} placeholder="Flexible" />
        </label>
        <label>
          <span>Description</span>
          <textarea rows="3" bind:value={form.feature.description} placeholder="+1 to Evasion"></textarea>
        </label>
        <button class="btn danger start" type="button" onclick={toggleFeature}>Remove feature</button>
      {:else}
        <button class="btn ghost start" type="button" onclick={toggleFeature}>+ Add a feature</button>
      {/if}
    </section>
  {/snippet}

  {#snippet preview()}
    <ArmorDetail card={previewCard} />
  {/snippet}
</HomebrewForm>

<style>
  .start { align-self: flex-start; }
</style>
