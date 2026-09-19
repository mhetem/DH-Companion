<script>
  import CardBrowser from './CardBrowser.svelte'
  import LootDetail from './LootDetail.svelte'
  import LootForm from './LootForm.svelte'
  import LootRoller from './LootRoller.svelte'
  import ShareExport from './ShareExport.svelte'
  import ShareImport from './ShareImport.svelte'
  import {
    BrowseConsumables,
    BrowseItems,
    DeleteCustomConsumable,
    DeleteCustomItem,
    HOMEBREW_TABLE,
    LOOT_RARITIES,
    LOOT_TABLES,
    errorMessage
  } from './api.js'

  // Loot has no tier, so the browser filters on rarity and source instead of the
  // usual tier/type pair. Homebrew is a source alongside the two books.
  const LOOT_FILTERS = [
    {
      key: 'rarity',
      label: 'Rarity',
      all: 'All rarities',
      options: LOOT_RARITIES.map((r) => ({ value: r.name, label: r.name }))
    },
    {
      key: 'table',
      label: 'Source',
      all: 'Everything',
      options: [...LOOT_TABLES, HOMEBREW_TABLE].map((t) => ({ value: t, label: t }))
    }
  ]

  const TABS = [
    { id: 'items', label: 'Items' },
    { id: 'consumables', label: 'Consumables' },
    { id: 'roll', label: 'Randomize' }
  ]

  let tab = $state('items')

  let editing = $state(null)
  let selectedSlug = $state('')
  let reloadToken = $state(0)
  let error = $state('')

  const isItems = $derived(tab === 'items')
  const noun = $derived(isItems ? 'item' : 'consumable')

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
    if (!confirm(`Delete homebrew ${noun} “${card.name}”?`)) return
    try {
      await (isItems ? DeleteCustomItem(card.slug) : DeleteCustomConsumable(card.slug))
      error = ''
      selectedSlug = ''
      reloadToken += 1
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

{#if editing !== null}
  <LootForm
    kind={tab}
    card={editing === 'new' ? null : editing}
    {onsaved}
    oncancel={() => (editing = null)}
  />
{:else}
  <div class="wrap">
    <div class="tabs">
      {#each TABS as t (t.id)}
        <button class:active={tab === t.id} onclick={() => (tab = t.id)}>{t.label}</button>
      {/each}
      {#if tab !== 'roll'}
        <div class="tools">
          <ShareImport expect={noun} {onimported} />
        </div>
      {/if}
    </div>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    {#key tab}
      {#if tab === 'roll'}
        <LootRoller mode="loot" />
      {:else}
        <CardBrowser
          filters={LOOT_FILTERS}
          load={isItems ? BrowseItems : BrowseConsumables}
          emptyLabel={`No ${tab} match these filters.`}
          onnew={() => (editing = 'new')}
          newLabel="New homebrew"
          initialSlug={selectedSlug}
          {reloadToken}
        >
          {#snippet row(item)}
            <span class="name">{item.name}</span>
            <span class="meta">
              {item.rarity}
              {#if item.source === 'custom'}
                <span class="chip custom">Homebrew</span>
              {:else}
                · {item.table} · Roll {item.roll}
              {/if}
            </span>
          {/snippet}

          {#snippet detail(item)}
            <LootDetail card={item}>
              {#snippet actions()}
                {#if item.source === 'custom'}
                  <button class="btn" onclick={() => (editing = item)}>Edit</button>
                  <ShareExport kind={noun} slug={item.slug} name={item.name} />
                  <button class="btn danger" onclick={() => remove(item)}>Delete</button>
                {/if}
              {/snippet}
            </LootDetail>
          {/snippet}
        </CardBrowser>
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
