<script>
  import { Events } from '@wailsio/runtime'
  import {
    ClearTable,
    CloseTableWindow,
    DeleteHandout,
    HANDOUT_KINDS,
    ListCampaigns,
    ListHandouts,
    OpenTableWindow,
    SaveHandout,
    ShowHandout,
    TableNow,
    TableScreens,
    TableStatus,
    ToggleTableFullscreen,
    errorMessage,
    handoutKindLabel
  } from './api.js'
  import EmptyState from '../EmptyState.svelte'
  import HandoutForm from './HandoutForm.svelte'
  import HandoutThumb from './HandoutThumb.svelte'

  // The GM's side of the player window: open it on the table's screen, then pick
  // what it shows from the campaign's maps, scenes and text. Go owns the window and
  // what's on it, so leaving this section or switching campaigns changes nothing
  // the players see — only Show and Blank do.
  //
  // The campaign is the one Campaigns last opened, and picking one here moves that
  // too: it's the campaign being run, wherever it was chosen.
  const CAMPAIGN_KEY = 'gm.campaigns.last'
  const SCREEN_KEY = 'gm.table.screen'

  let status = $state({ open: false, fullscreen: false, placeable: true })
  let screens = $state([])
  let screenId = $state(localStorage.getItem(SCREEN_KEY) ?? '')
  let showing = $state({ handoutId: 0, kind: '', title: '', body: '', mediaUrl: '', mediaType: '' })

  let campaigns = $state([])
  let campaignId = $state(null)
  let handouts = $state([])
  let kindFilter = $state('')
  let editing = $state(null)
  let editSeq = 0
  let loading = $state(true)
  let busy = $state('')
  let error = $state('')

  // An unplugged screen falls back to the first one this window isn't on — the
  // table's, in the usual two-screen setup.
  const screenChoice = $derived(
    screens.some((s) => s.id === screenId) ? screenId : (screens.find((s) => !s.here)?.id ?? '')
  )
  const shown = $derived(kindFilter ? handouts.filter((h) => h.kind === kindFilter) : handouts)

  $effect(() =>
    Events.On('table:window', (e) => {
      status = e.data
      refreshScreens()
    })
  )
  $effect(() => Events.On('table:show', (e) => (showing = e.data)))

  async function load() {
    try {
      const [win, screenList, now, campaignList] = await Promise.all([
        TableStatus(),
        TableScreens(),
        TableNow(),
        ListCampaigns()
      ])
      status = win
      screens = screenList ?? []
      showing = now
      campaigns = campaignList ?? []
      const remembered = Number(localStorage.getItem(CAMPAIGN_KEY))
      campaignId = campaigns.some((c) => c.id === remembered) ? remembered : (campaigns[0]?.id ?? null)
      await loadHandouts()
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  load()

  async function loadHandouts() {
    handouts = campaignId === null ? [] : ((await ListHandouts(campaignId)) ?? [])
  }

  async function refreshScreens() {
    try {
      screens = (await TableScreens()) ?? []
    } catch {
      // A stale list only costs the picker an entry; Open still asks Go afresh.
    }
  }

  async function run(label, fn) {
    busy = label
    try {
      await fn()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = ''
    }
  }

  function pickScreen(id) {
    screenId = id
    localStorage.setItem(SCREEN_KEY, id)
  }

  function pickCampaign(id) {
    campaignId = id
    localStorage.setItem(CAMPAIGN_KEY, String(id))
    editing = null
    run('campaign', loadHandouts)
  }

  const openWindow = () => run('window', async () => (status = await OpenTableWindow(screenChoice)))
  const closeWindow = () => run('window', CloseTableWindow)
  const toggleFullscreen = () => run('window', ToggleTableFullscreen)
  const blank = () => run('blank', ClearTable)

  // Show opens the window if it's closed — pressing Show means "put it in front of
  // them", and a closed window would make that a silent no-op.
  const show = (h) =>
    run(`show-${h.id}`, async () => {
      showing = await ShowHandout(h.id)
      if (!status.open) status = await OpenTableWindow(screenChoice)
    })

  function startEdit(handout, kind = handout?.kind) {
    editing = { key: ++editSeq, handout, kind }
  }

  // Saving what's on the table puts the saved version up there too. Edits never
  // reach the players while they're being typed — only once Save is pressed.
  async function save(input) {
    const saved = await SaveHandout({ ...input, campaignId })
    editing = null
    await loadHandouts()
    if (showing.handoutId === saved.id) showing = await ShowHandout(saved.id)
  }

  function remove(h) {
    if (!confirm(`Delete “${h.title}”? Its file goes too, unless another handout uses it.`)) return
    run(`delete-${h.id}`, async () => {
      if (showing.handoutId === h.id) await ClearTable()
      await DeleteHandout(h.id)
      if (editing?.handout?.id === h.id) editing = null
      await loadHandouts()
    })
  }
</script>

<div class="player">
  <header>
    <h2>Player window</h2>
    <p class="blurb">
      A second window for the screen the players can see — maps, scenes and text. Nothing reaches it until you
      press Show.
    </p>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <section class="console">
    <div class="window">
      <h3>Window</h3>
      {#if status.open}
        <p class="state on">Open · {status.fullscreen ? 'fullscreen' : 'in a window'}</p>
        <div class="row">
          <button class="btn" onclick={toggleFullscreen} disabled={!!busy}>
            {status.fullscreen ? 'Leave fullscreen' : 'Fullscreen'}
          </button>
          <button class="btn ghost" onclick={closeWindow} disabled={!!busy}>Close</button>
        </div>
      {:else}
        <p class="state">Closed</p>
        <div class="row">
          {#if status.placeable && screens.length > 1}
            <select value={screenChoice} onchange={(e) => pickScreen(e.currentTarget.value)} aria-label="Screen">
              {#each screens as s (s.id)}
                <option value={s.id}>{s.label} · {s.width}×{s.height}{s.here ? ' · this one' : ''}</option>
              {/each}
            </select>
          {/if}
          <button class="btn primary" onclick={openWindow} disabled={!!busy}>Open player window</button>
        </div>
      {/if}
      <p class="hint">
        {#if !status.placeable}
          This desktop doesn't let apps place their own windows, so drag it onto the table's screen and
          double-click it to fill that screen.
        {:else if screens.length < 2}
          With one screen it opens as a window here. Connect the table's screen first and it opens there,
          fullscreen.
        {:else}
          It opens fullscreen on the screen picked here.
        {/if}
        Double-click it or press F11 to toggle fullscreen.
      </p>
    </div>

    <div class="now">
      <h3>On the table</h3>
      <div class="nowrow">
        <div class="nowthumb">
          <HandoutThumb kind={showing.kind} mediaUrl={showing.mediaUrl} mediaType={showing.mediaType} body={showing.body} />
        </div>
        <div class="nowinfo">
          {#if showing.kind}
            <p class="nowtitle">{showing.title}</p>
            <span class="chip">{handoutKindLabel(showing.kind)}</span>
          {:else}
            <p class="nowtitle muted">Nothing — the screen is black</p>
          {/if}
          <button class="btn" onclick={blank} disabled={!showing.kind || !!busy}>Blank the screen</button>
        </div>
      </div>
    </div>
  </section>

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if !campaigns.length}
    <EmptyState title="No campaigns yet">
      Maps, scenes and text belong to a campaign. Create one under Campaigns, then come back to fill it.
    </EmptyState>
  {:else}
    <div class="bar">
      <select value={campaignId} onchange={(e) => pickCampaign(Number(e.currentTarget.value))} aria-label="Campaign">
        {#each campaigns as c (c.id)}
          <option value={c.id}>{c.name}</option>
        {/each}
      </select>
      <div class="filters">
        <button class="chip" class:on={kindFilter === ''} onclick={() => (kindFilter = '')}>All</button>
        {#each HANDOUT_KINDS as k (k.value)}
          <button class="chip" class:on={kindFilter === k.value} onclick={() => (kindFilter = k.value)}>
            {k.plural}
          </button>
        {/each}
      </div>
      <div class="add">
        {#each HANDOUT_KINDS as k (k.value)}
          <button class="btn ghost" onclick={() => startEdit(null, k.value)}>+ {k.label}</button>
        {/each}
      </div>
    </div>

    {#if editing}
      {#key editing.key}
        <HandoutForm handout={editing.handout} kind={editing.kind} onsave={save} oncancel={() => (editing = null)} />
      {/key}
    {/if}

    {#if !shown.length}
      <p class="empty">
        {#if kindFilter}
          No {handoutKindLabel(kindFilter).toLowerCase()} handouts in this campaign yet.
        {:else}
          Nothing here yet. Add the campaign's maps, the scenes that set a mood — a still image or a looping
          video — and any text the players should read off the screen.
        {/if}
      </p>
    {:else}
      <ul class="grid">
        {#each shown as h (h.id)}
          <li class:live={showing.handoutId === h.id} class:editing={editing?.handout?.id === h.id}>
            <HandoutThumb kind={h.kind} mediaUrl={h.mediaUrl} mediaType={h.mediaType} body={h.body} missing={h.mediaMissing} />
            <div class="meta">
              <span class="name" title={h.title}>{h.title}</span>
              {#if showing.handoutId === h.id}
                <span class="chip live">On the table</span>
              {:else}
                <span class="chip">{handoutKindLabel(h.kind)}</span>
              {/if}
            </div>
            <div class="actions">
              <button class="btn primary" onclick={() => show(h)} disabled={!!busy || h.mediaMissing}>
                {busy === `show-${h.id}` ? 'Showing…' : 'Show'}
              </button>
              <button class="btn ghost" onclick={() => startEdit(h)}>Edit</button>
              <button class="btn danger" onclick={() => remove(h)} disabled={!!busy} title="Delete">✕</button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  {/if}
</div>

<style>
  .player {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.75rem;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  .player > :global(*) {
    flex-shrink: 0;
    width: 100%;
    max-width: 68rem;
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

  .console {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.75rem;
  }

  .window,
  .now {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  h3 {
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .state {
    margin: 0;
    font-size: 0.9rem;
    color: var(--muted);
  }

  .state.on { color: var(--ok); }

  .row {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    flex-wrap: wrap;
  }

  .hint {
    margin: 0;
    font-size: 0.75rem;
    line-height: 1.45;
    color: var(--muted);
  }

  .nowrow {
    display: flex;
    gap: 0.75rem;
  }

  .nowthumb { flex: 0 0 12rem; }

  .nowinfo {
    display: flex;
    flex: 1;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.4rem;
    min-width: 0;
  }

  .nowtitle {
    margin: 0;
    max-width: 100%;
    overflow: hidden;
    font-size: 0.95rem;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .nowtitle.muted { color: var(--muted); }

  .bar {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    flex-wrap: wrap;
  }

  .filters {
    display: flex;
    flex: 1;
    gap: 0.25rem;
    flex-wrap: wrap;
  }

  .filters .chip {
    background: transparent;
    font: inherit;
    font-size: 0.7rem;
    cursor: pointer;
  }

  .filters .chip.on {
    border-color: var(--fear);
    color: var(--fear);
  }

  .add {
    display: flex;
    gap: 0.25rem;
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
    gap: 0.6rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .grid li {
    display: flex;
    flex-direction: column;
    gap: 0.45rem;
    padding: 0.5rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .grid li.editing { border-color: var(--fear); }

  .grid li.live { border-color: var(--ok); }

  .meta {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .name {
    flex: 1;
    min-width: 0;
    overflow: hidden;
    font-size: 0.85rem;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .chip.live {
    border-color: var(--ok);
    color: var(--ok);
  }

  .actions {
    display: flex;
    gap: 0.35rem;
  }

  .actions .btn.primary { flex: 1; }
</style>
