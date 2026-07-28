<script>
  import CardBrowser from './CardBrowser.svelte'
  import AdversaryDetail from './AdversaryDetail.svelte'
  import ShareExport from './ShareExport.svelte'
  import ShareImport from './ShareImport.svelte'
  import AdversaryForm from './AdversaryForm.svelte'
  import {
    ADVERSARY_TYPES,
    BrowseAdversaries,
    DeleteCustomAdversary,
    costLabel,
    errorMessage
  } from './api.js'

  // null = browsing, 'new' = creating, otherwise the card being edited. Showing
  // the form unmounts the browser, so returning reloads the list on its own —
  // only delete needs an explicit reload, since the browser stays mounted.
  let editing = $state(null)
  let selectedSlug = $state('')
  let reloadToken = $state(0)
  let error = $state('')

  function onsaved(saved) {
    selectedSlug = saved.slug
    editing = null
  }

  function onimported(card) {
    selectedSlug = card.slug
    reloadToken += 1
  }

  async function remove(card) {
    if (
      !confirm(`Delete homebrew adversary “${card.name}”? Encounters using it will show it as missing.`)
    )
      return
    try {
      await DeleteCustomAdversary(card.slug)
      error = ''
      selectedSlug = ''
      reloadToken += 1
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

{#if editing !== null}
  <AdversaryForm
    card={editing === 'new' ? null : editing}
    {onsaved}
    oncancel={() => (editing = null)}
  />
{:else}
  <div class="wrap">
    {#if error}
      <p class="error">{error}</p>
    {/if}
    <div class="tools">
      <ShareImport expect="adversary" {onimported} />
    </div>
    <CardBrowser
      types={ADVERSARY_TYPES}
      load={BrowseAdversaries}
      emptyLabel="No adversaries match these filters."
      onnew={() => (editing = 'new')}
      newLabel="New homebrew"
      initialSlug={selectedSlug}
      {reloadToken}
    >
      {#snippet row(item)}
        <span class="name">{item.name}</span>
        <span class="meta">
          Tier {item.tier} · {item.type} · {costLabel(item.type)}
          {#if item.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
        </span>
      {/snippet}

      {#snippet detail(item)}
        <AdversaryDetail card={item}>
          {#snippet actions()}
            {#if item.source === 'custom'}
              <button class="btn" onclick={() => (editing = item)}>Edit</button>
              <ShareExport kind="adversary" slug={item.slug} name={item.name} />
              <button class="btn danger" onclick={() => remove(item)}>Delete</button>
            {/if}
          {/snippet}
        </AdversaryDetail>
      {/snippet}
    </CardBrowser>
  </div>
{/if}

<style>
  .tools {
    display: flex;
    justify-content: flex-end;
    gap: 0.5rem;
    padding: 0.6rem 1rem 0;
  }

  .wrap {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
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

  .error {
    margin: 0;
    padding: 0.5rem 1rem;
    border-bottom: 1px solid var(--danger);
    font-size: 0.8rem;
    color: var(--danger);
  }
</style>
