<script>
  import CardBrowser from './CardBrowser.svelte'
  import WeaponDetail from './WeaponDetail.svelte'
  import ArmorDetail from './ArmorDetail.svelte'
  import WeaponForm from './WeaponForm.svelte'
  import ArmorForm from './ArmorForm.svelte'
  import LootRoller from './LootRoller.svelte'
  import ShareExport from './ShareExport.svelte'
  import ShareImport from './ShareImport.svelte'
  import {
    BrowseArmor,
    BrowseWeapons,
    DeleteCustomArmor,
    DeleteCustomWeapon,
    TIERS,
    WEAPON_CATEGORIES,
    WEAPON_TYPES,
    errorMessage
  } from './api.js'

  const tierOptions = TIERS.map((t) => ({ value: t, label: `Tier ${t}` }))

  // Weapons browse on three axes rather than the usual two; armor has only a tier.
  const WEAPON_FILTERS = [
    { key: 'tier', label: 'Tier', all: 'All tiers', options: tierOptions },
    {
      key: 'category',
      label: 'Category',
      all: 'All categories',
      options: WEAPON_CATEGORIES.map((c) => ({ value: c, label: c }))
    },
    {
      key: 'type',
      label: 'Damage',
      all: 'Physical & magic',
      options: WEAPON_TYPES.map((t) => ({ value: t, label: t }))
    }
  ]

  const ARMOR_FILTERS = [{ key: 'tier', label: 'Tier', all: 'All tiers', options: tierOptions }]

  const TABS = [
    { id: 'weapons', label: 'Weapons' },
    { id: 'armor', label: 'Armor' },
    { id: 'roll', label: 'Randomize' }
  ]

  let tab = $state('weapons')

  // Same shape as the adversary and environment browsers: only delete needs an
  // explicit reload, because it leaves this component mounted with a stale list.
  let editing = $state(null)
  let selectedSlug = $state('')
  let reloadToken = $state(0)
  let error = $state('')

  const isWeapons = $derived(tab === 'weapons')

  function onsaved(saved) {
    selectedSlug = saved.slug
    editing = null
    reloadToken += 1
  }

  function onimported(card) {
    selectedSlug = card.slug
    reloadToken += 1
  }

  async function remove(card) {
    const noun = isWeapons ? 'weapon' : 'armor'
    if (!confirm(`Delete homebrew ${noun} “${card.name}”?`)) return
    try {
      await (isWeapons ? DeleteCustomWeapon(card.slug) : DeleteCustomArmor(card.slug))
      error = ''
      selectedSlug = ''
      reloadToken += 1
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

{#if editing !== null}
  {#if isWeapons}
    <WeaponForm
      card={editing === 'new' ? null : editing}
      {onsaved}
      oncancel={() => (editing = null)}
    />
  {:else}
    <ArmorForm
      card={editing === 'new' ? null : editing}
      {onsaved}
      oncancel={() => (editing = null)}
    />
  {/if}
{:else}
  <div class="wrap">
    <div class="tabs">
      {#each TABS as t (t.id)}
        <button class:active={tab === t.id} onclick={() => (tab = t.id)}>{t.label}</button>
      {/each}
      {#if tab !== 'roll'}
        <div class="tools">
          <ShareImport expect={isWeapons ? 'weapon' : 'armor'} {onimported} />
        </div>
      {/if}
    </div>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    <!-- Keyed so switching tabs remounts the browser; CardBrowser reads its filter
         set once at mount, and the tabs do not share one. -->
    {#key tab}
      {#if tab === 'weapons'}
        <CardBrowser
          filters={WEAPON_FILTERS}
          load={BrowseWeapons}
          emptyLabel="No weapons match these filters."
          onnew={() => (editing = 'new')}
          newLabel="New homebrew"
          initialSlug={selectedSlug}
          {reloadToken}
        >
          {#snippet row(item)}
            <span class="name">{item.name}</span>
            <span class="meta">
              Tier {item.tier} · {item.category} · {item.damage}
              {item.damageType}
              {#if item.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
            </span>
          {/snippet}

          {#snippet detail(item)}
            <WeaponDetail card={item}>
              {#snippet actions()}
                {#if item.source === 'custom'}
                  <button class="btn" onclick={() => (editing = item)}>Edit</button>
                  <ShareExport kind="weapon" slug={item.slug} name={item.name} />
                  <button class="btn danger" onclick={() => remove(item)}>Delete</button>
                {/if}
              {/snippet}
            </WeaponDetail>
          {/snippet}
        </CardBrowser>
      {:else if tab === 'armor'}
        <CardBrowser
          filters={ARMOR_FILTERS}
          load={BrowseArmor}
          emptyLabel="No armor matches these filters."
          onnew={() => (editing = 'new')}
          newLabel="New homebrew"
          initialSlug={selectedSlug}
          {reloadToken}
        >
          {#snippet row(item)}
            <span class="name">{item.name}</span>
            <span class="meta">
              Tier {item.tier} · {item.thresholdMajor} / {item.thresholdSevere} · Score {item.baseScore}
              {#if item.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
            </span>
          {/snippet}

          {#snippet detail(item)}
            <ArmorDetail card={item}>
              {#snippet actions()}
                {#if item.source === 'custom'}
                  <button class="btn" onclick={() => (editing = item)}>Edit</button>
                  <ShareExport kind="armor" slug={item.slug} name={item.name} />
                  <button class="btn danger" onclick={() => remove(item)}>Delete</button>
                {/if}
              {/snippet}
            </ArmorDetail>
          {/snippet}
        </CardBrowser>
      {:else}
        <LootRoller mode="equipment" />
      {/if}
    {/key}
  </div>
{/if}

<style>
  .wrap {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
  }

  .tabs {
    display: flex;
    align-items: center;
    gap: 0.25rem;
    padding: 0.6rem 1rem 0;
  }

  .tools { margin-left: auto; }

  .tabs button {
    padding: 0.35rem 0.7rem;
    border: 1px solid transparent;
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .tabs button:hover { color: var(--text); }

  .tabs button.active {
    background: var(--panel-2);
    border-color: var(--line);
    color: var(--text);
  }

  .error {
    margin: 0;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--danger);
    font-size: 0.8rem;
    color: var(--danger);
  }

  .name {
    display: block;
    font-size: 0.9rem;
  }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.35rem;
    font-size: 0.72rem;
    color: var(--muted);
  }
</style>
