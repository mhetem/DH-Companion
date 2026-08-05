<script>
  import { ListCampaigns, Search, errorMessage } from './api.js'
  import { renderExcerpt } from './markdown.js'

  const LIMIT = 50
  const ENTITY_LABELS = {
    note: 'Note',
    master: 'Master note',
    adversary: 'Adversary',
    environment: 'Environment'
  }

  // The two campaign-scoped entities — the scope select narrows these and leaves
  // the cards alone, matching what gm.Search does on the Go side.
  const CAMPAIGN_ENTITIES = ['note', 'master']

  let query = $state('')
  let scope = $state(0)
  let campaigns = $state([])
  let hits = $state([])
  let searched = $state(false)
  let busy = $state(false)
  let error = $state('')
  let openKey = $state(null)

  // The backend quotes each token before it reaches FTS5, so punctuation typed
  // mid-search can't reach the MATCH grammar — no debounce guard needed for it.
  let timer = null

  ListCampaigns()
    .then((rows) => (campaigns = rows ?? []))
    .catch((e) => (error = errorMessage(e)))

  async function run() {
    if (!query.trim()) {
      hits = []
      searched = false
      return
    }
    busy = true
    try {
      hits = (await Search(query, Number(scope), LIMIT)) ?? []
      searched = true
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function schedule() {
    clearTimeout(timer)
    timer = setTimeout(run, 150)
  }

  function submit(event) {
    event.preventDefault()
    clearTimeout(timer)
    run()
  }

  const keyOf = (hit) => `${hit.entity}:${hit.entityId}:${hit.slug}`
  const toggle = (hit) => (openKey = openKey === keyOf(hit) ? null : keyOf(hit))
</script>

<div class="search">
  <header>
    <h2>Search</h2>
    <p class="blurb">One index over your notes, your campaigns' master notes, and every adversary and environment card, SRD or homebrew.</p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <form onsubmit={submit}>
    <input
      bind:value={query}
      oninput={schedule}
      placeholder="mire, ritual, a name you half remember…"
      autocomplete="off"
    />
    <select bind:value={scope} onchange={run}>
      <option value={0}>Notes from every campaign</option>
      {#each campaigns as campaign (campaign.id)}
        <option value={campaign.id}>Notes from {campaign.name}</option>
      {/each}
    </select>
  </form>

  {#if busy && !hits.length}
    <p class="empty">Searching…</p>
  {:else if searched && !hits.length}
    <p class="empty">Nothing matches “{query}”.</p>
  {:else if !searched}
    <p class="empty">Cards are matched on their features and motives too, not just their names.</p>
  {:else}
    <p class="count">{hits.length}{hits.length === LIMIT ? '+' : ''} {hits.length === 1 ? 'match' : 'matches'}</p>
    <ul class="list">
      {#each hits as hit (keyOf(hit))}
        <li>
          <button class="row" onclick={() => toggle(hit)}>
            <span class="chip" class:note={CAMPAIGN_ENTITIES.includes(hit.entity)}>{ENTITY_LABELS[hit.entity] ?? hit.entity}</span>
            <span class="name">{hit.title}</span>
            <span class="excerpt">{@html renderExcerpt(hit.excerpt)}</span>
          </button>
          {#if openKey === keyOf(hit)}
            <div class="body">
              <p class="full">{@html renderExcerpt(hit.excerpt)}</p>
              {#if hit.entity === 'master'}
                <p class="hint">The master note of “{hit.title}” — open that campaign to read or edit it in full.</p>
              {:else if hit.entity === 'note'}
                <p class="hint">Open this campaign's Notes tab to read or edit it in full.</p>
              {:else}
                <p class="hint">Card slug <code>{hit.slug}</code> — find it in the {ENTITY_LABELS[hit.entity]?.toLowerCase()} browser.</p>
              {/if}
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .search {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.75rem;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .blurb {
    margin: 0.25rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .error {
    margin: 0;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  form {
    display: flex;
    gap: 0.5rem;
  }

  form input {
    flex: 1;
    min-width: 0;
    padding: 0.5rem 0.7rem;
    font-size: 0.95rem;
  }

  form select { flex: 0 0 15rem; }

  .count {
    margin: 0;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .list li {
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .row {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    width: 100%;
    padding: 0.5rem 0.7rem;
    border: none;
    background: none;
    color: var(--text);
    font: inherit;
    font-size: 0.85rem;
    text-align: left;
    cursor: pointer;
  }

  .chip.note {
    border-color: var(--fear);
    color: var(--fear);
  }

  .name {
    flex: 0 0 auto;
    max-width: 14rem;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .excerpt {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    font-size: 0.8rem;
    color: var(--muted);
  }

  .excerpt :global(mark) {
    background: none;
    color: var(--hope);
  }

  .body {
    padding: 0 0.7rem 0.6rem;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .full {
    margin: 0 0 0.4rem;
    white-space: pre-wrap;
  }

  .full :global(mark) {
    background: none;
    color: var(--hope);
  }

  .hint {
    margin: 0;
    font-size: 0.75rem;
    opacity: 0.8;
  }

  code {
    padding: 0.05rem 0.25rem;
    border-radius: 4px;
    background: var(--panel-2);
    font-size: 0.8rem;
  }
</style>
