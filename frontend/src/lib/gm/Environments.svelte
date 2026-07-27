<script>
  import CardBrowser from './CardBrowser.svelte'
  import EnvironmentDetail from './EnvironmentDetail.svelte'
  import EnvironmentForm from './EnvironmentForm.svelte'
  import {
    BrowseEnvironments,
    DeleteCustomEnvironment,
    ENVIRONMENT_TYPES,
    errorMessage
  } from './api.js'

  // Same shape as the adversary browser — see Adversaries.svelte for why only
  // delete needs an explicit reload.
  let editing = $state(null)
  let selectedSlug = $state('')
  let reloadToken = $state(0)
  let error = $state('')

  function onsaved(saved) {
    selectedSlug = saved.slug
    editing = null
  }

  async function remove(card) {
    if (
      !confirm(`Delete homebrew environment “${card.name}”? Encounters using it will lose it.`)
    )
      return
    try {
      await DeleteCustomEnvironment(card.slug)
      error = ''
      selectedSlug = ''
      reloadToken += 1
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

{#if editing !== null}
  <EnvironmentForm
    card={editing === 'new' ? null : editing}
    {onsaved}
    oncancel={() => (editing = null)}
  />
{:else}
  <div class="wrap">
    {#if error}
      <p class="error">{error}</p>
    {/if}
    <CardBrowser
      types={ENVIRONMENT_TYPES}
      load={BrowseEnvironments}
      emptyLabel="No environments match these filters."
      onnew={() => (editing = 'new')}
      newLabel="New homebrew"
      initialSlug={selectedSlug}
      {reloadToken}
    >
      {#snippet row(item)}
        <span class="name">{item.name}</span>
        <span class="meta">
          Tier {item.tier} · {item.type}
          {#if item.difficulty}· Difficulty {item.difficulty}{/if}
          {#if item.source === 'custom'}<span class="chip custom">Homebrew</span>{/if}
        </span>
      {/snippet}

      {#snippet detail(item)}
        <EnvironmentDetail card={item}>
          {#snippet actions()}
            {#if item.source === 'custom'}
              <button class="btn" onclick={() => (editing = item)}>Edit</button>
              <button class="btn danger" onclick={() => remove(item)}>Delete</button>
            {/if}
          {/snippet}
        </EnvironmentDetail>
      {/snippet}
    </CardBrowser>
  </div>
{/if}

<style>
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
