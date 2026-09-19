<script>
  import { untrack } from 'svelte'
  import HomebrewForm from './HomebrewForm.svelte'
  import WeaponDetail from './WeaponDetail.svelte'
  import {
    CreateCustomWeapon,
    DAMAGE_TYPES,
    TIERS,
    UpdateCustomWeapon,
    WEAPON_BURDENS,
    WEAPON_CATEGORIES,
    WEAPON_RANGES,
    WEAPON_TRAITS,
    WEAPON_TYPES,
    errorMessage
  } from './api.js'

  let { card = null, onsaved, oncancel } = $props()

  // Snapshot once, like the environment form: the parent unmounts this to leave
  // it, so an instance always edits the card it mounted with.
  const original = untrack(() => card)
  const editing = original !== null

  let form = $state(
    original
      ? { ...original, feature: original.feature ? { ...original.feature } : null }
      : {
          kind: 'weapon',
          slug: '',
          name: '',
          tier: '1',
          type: 'Physical',
          category: 'Primary',
          description: '',
          trait: 'Agility',
          range: 'Melee',
          damage: 'd8',
          damageType: 'phy',
          burden: 'One-Handed',
          feature: null
        }
  )

  let saving = $state(false)
  let error = $state('')

  const previewCard = $derived({ ...form, source: 'custom' })

  // A weapon has at most one feature, so this is a toggle rather than the
  // repeater the adversary and environment forms use.
  function toggleFeature() {
    form.feature = form.feature ? null : { title: '', description: '' }
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      const saved = editing ? await UpdateCustomWeapon(form) : await CreateCustomWeapon(form)
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
  title={editing ? `Edit ${original.name}` : 'New homebrew weapon'}
  formId="weapon-form"
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
          <input bind:value={form.name} placeholder="Thornwood Glaive" required />
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
          <span>Category</span>
          <select bind:value={form.category}>
            {#each WEAPON_CATEGORIES as c (c)}
              <option value={c}>{c}</option>
            {/each}
          </select>
        </label>
        <label>
          <span>Damage</span>
          <select bind:value={form.type}>
            {#each WEAPON_TYPES as t (t)}
              <option value={t}>{t}</option>
            {/each}
          </select>
        </label>
        <label>
          <span>Burden</span>
          <select bind:value={form.burden}>
            {#each WEAPON_BURDENS as b (b)}
              <option value={b}>{b}</option>
            {/each}
          </select>
        </label>
      </div>

      <div class="row">
        <label>
          <span>Trait</span>
          <select bind:value={form.trait}>
            {#each WEAPON_TRAITS as t (t)}
              <option value={t}>{t}</option>
            {/each}
          </select>
        </label>
        <label>
          <span>Range</span>
          <select bind:value={form.range}>
            {#each WEAPON_RANGES as r (r)}
              <option value={r}>{r}</option>
            {/each}
          </select>
        </label>
        <label>
          <span>Damage dice</span>
          <input bind:value={form.damage} placeholder="d10+3" />
        </label>
        <label class="narrow">
          <span>Type</span>
          <select bind:value={form.damageType}>
            {#each DAMAGE_TYPES as d (d.value)}
              <option value={d.value}>{d.label}</option>
            {/each}
          </select>
        </label>
      </div>

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
          <input bind:value={form.feature.title} placeholder="Reliable" />
        </label>
        <label>
          <span>Description</span>
          <textarea rows="3" bind:value={form.feature.description} placeholder="+1 to attack rolls"></textarea>
        </label>
        <button class="btn danger start" type="button" onclick={toggleFeature}>Remove feature</button>
      {:else}
        <button class="btn ghost start" type="button" onclick={toggleFeature}>+ Add a feature</button>
      {/if}
    </section>
  {/snippet}

  {#snippet preview()}
    <WeaponDetail card={previewCard} />
  {/snippet}
</HomebrewForm>

<style>
  .start { align-self: flex-start; }
</style>
