<script>
  import {
    DeleteWorldNote,
    ListWorldNotes,
    MoveWorldNote,
    SaveWorldNote,
    ShiftWorldNote,
    WORLD_CHILD_KIND,
    WORLD_KINDS,
    errorMessage,
    worldKindLabel
  } from './api.js'
  import WorldPage from './WorldPage.svelte'

  // The campaign's worldbuilding: a tree of pages — regions holding cities holding
  // places, timelines holding their events. Campaigns keys this on the campaign id,
  // so campaignId is fixed for the component's life, and so is the key the open
  // page is remembered under.
  let { campaignId } = $props()

  const LAST_KEY = `gm.world.last.${campaignId}`

  let notes = $state([])
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let selectedId = $state(null)
  let expanded = $state({})
  let filter = $state('')
  let creating = $state(null)
  let moveTo = $state('')

  const byId = $derived(new Map(notes.map((n) => [n.id, n])))

  // The list arrives in sibling order, so grouping it by parent keeps that order.
  const childrenOf = $derived.by(() => {
    const out = new Map()
    for (const n of notes) {
      const key = n.parentId ?? 0
      if (!out.has(key)) out.set(key, [])
      out.get(key).push(n)
    }
    return out
  })

  const selected = $derived(byId.get(selectedId) ?? null)
  const trail = $derived(ancestors(selected))
  const children = $derived(kids(selected?.id ?? null))

  // Filtering keeps every match plus the pages above it, so a hit is always shown
  // where it lives rather than floating loose.
  const needle = $derived(filter.trim().toLowerCase())
  const matches = $derived.by(() => {
    if (!needle) return null
    const keep = new Set()
    for (const n of notes) {
      if (!n.title.toLowerCase().includes(needle)) continue
      for (let at = n; at && !keep.has(at.id); at = byId.get(at.parentId)) keep.add(at.id)
    }
    return keep
  })

  const rows = $derived.by(() => {
    const out = []
    const walk = (parentId, depth) => {
      for (const n of kids(parentId)) {
        if (matches && !matches.has(n.id)) continue
        const count = kids(n.id).length
        out.push({ note: n, depth, count })
        if (count && (matches || expanded[n.id])) walk(n.id, depth + 1)
      }
    }
    walk(null, 0)
    return out
  })

  // A page can't move inside itself or anything under it — the backend refuses
  // too, but offering the option only to reject it would be a trap.
  const moveTargets = $derived.by(() => {
    if (!selected) return []
    const banned = new Set([selected.id])
    const stack = [selected.id]
    while (stack.length) {
      for (const c of kids(stack.pop())) {
        banned.add(c.id)
        stack.push(c.id)
      }
    }
    const out = []
    const walk = (parentId, depth) => {
      for (const n of kids(parentId)) {
        if (banned.has(n.id)) continue
        if (n.id !== selected.parentId) out.push({ note: n, depth })
        walk(n.id, depth + 1)
      }
    }
    walk(null, 0)
    return out
  })

  function kids(id) {
    return childrenOf.get(id ?? 0) ?? []
  }

  function ancestors(note) {
    const out = []
    for (let at = note && byId.get(note.parentId); at; at = byId.get(at.parentId)) out.unshift(at)
    return out
  }

  function descendants(id) {
    let n = 0
    for (const c of kids(id)) n += 1 + descendants(c.id)
    return n
  }

  async function refresh() {
    try {
      notes = (await ListWorldNotes(campaignId)) ?? []
      if (selectedId !== null && !notes.some((n) => n.id === selectedId)) select(null)
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh().then(() => {
    const remembered = Number(localStorage.getItem(LAST_KEY))
    if (notes.some((n) => n.id === remembered)) select(remembered)
  })

  function select(id) {
    selectedId = id
    creating = null
    moveTo = ''
    localStorage.setItem(LAST_KEY, id === null ? '' : String(id))
    for (const a of ancestors(byId.get(id))) expanded[a.id] = true
  }

  const toggle = (note) => (expanded[note.id] = !expanded[note.id])

  function startCreate() {
    const kind = selected ? (WORLD_CHILD_KIND[selected.kind] ?? 'page') : 'region'
    creating = { kind, title: '' }
  }

  function newTopLevel() {
    select(null)
    startCreate()
  }

  async function create(event) {
    event.preventDefault()
    busy = true
    try {
      const note = await SaveWorldNote({
        id: null,
        campaignId,
        parentId: selected?.id ?? null,
        kind: creating.kind,
        title: creating.title,
        worldDate: '',
        body: ''
      })
      notes = [...notes, note]
      if (note.parentId !== null) expanded[note.parentId] = true
      select(note.id)
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  // An autosave landing updates the tree in place; nothing else about the list moved.
  function saved(note) {
    notes = notes.map((n) => (n.id === note.id ? note : n))
  }

  async function structural(fn) {
    busy = true
    try {
      await fn()
      await refresh()
      if (selectedId !== null) select(selectedId)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  const shift = (delta) => structural(() => ShiftWorldNote(selected.id, delta))

  function move() {
    if (moveTo === '') return
    const parentId = moveTo === 'top' ? null : Number(moveTo)
    structural(() => MoveWorldNote(selected.id, parentId))
  }

  // Moving the selection off the page unmounts its editor, and the editor's unmount
  // flush can reach the backend either side of the delete. Landing after it, the
  // update matches no row and changes nothing — it can't bring the page back.
  async function remove() {
    const note = selected
    const nested = descendants(note.id)
    const question = nested
      ? `Delete “${note.title}” and the ${nested} ${nested === 1 ? 'page' : 'pages'} inside it?`
      : `Delete “${note.title}”?`
    if (!confirm(question)) return
    select(note.parentId)
    await structural(() => DeleteWorldNote(note.id))
  }

  const siblings = $derived(selected ? kids(selected.parentId) : [])
  const first = $derived(siblings[0]?.id === selected?.id)
  const last = $derived(siblings[siblings.length - 1]?.id === selected?.id)

  function excerpt(body) {
    for (const line of body.split('\n')) {
      const text = line.replace(/^[\s#>*+-]+/, '').replace(/[*_`]/g, '').trim()
      if (text) return text
    }
    return ''
  }
</script>

<section class="world">
  <aside class="tree">
    <div class="treehead">
      <input bind:value={filter} placeholder="Filter pages" aria-label="Filter pages" />
      <button class="btn ghost" onclick={newTopLevel} disabled={busy}>+ New</button>
    </div>

    {#if loading}
      <p class="empty">Loading…</p>
    {:else if !notes.length}
      <p class="empty">No pages yet.</p>
    {:else if !rows.length}
      <p class="empty">Nothing titled “{filter.trim()}”.</p>
    {:else}
      <button class="rootrow" class:on={selectedId === null} onclick={() => select(null)}>The world</button>
      <ul>
        {#each rows as row (row.note.id)}
          <li class:on={row.note.id === selectedId} style:padding-left="{row.depth * 0.9}rem">
            {#if row.count}
              <button
                class="caret"
                class:open={matches || expanded[row.note.id]}
                onclick={() => toggle(row.note)}
                disabled={!!matches}
                aria-label={expanded[row.note.id] ? 'Collapse' : 'Expand'}
              >›</button>
            {:else}
              <span class="caret spacer"></span>
            {/if}
            <button class="name" onclick={() => select(row.note.id)} title={row.note.title}>
              {row.note.title}
            </button>
            <span class="kind">{worldKindLabel(row.note.kind)}</span>
          </li>
        {/each}
      </ul>
    {/if}
  </aside>

  <div class="page">
    <nav class="trail" aria-label="Where this page sits">
      <button class="crumb" onclick={() => select(null)}>The world</button>
      {#each trail as a (a.id)}
        <span class="sep">›</span>
        <button class="crumb" onclick={() => select(a.id)}>{a.title}</button>
      {/each}
      {#if selected}
        <span class="sep">›</span>
        <span class="crumb current">{selected.title}</span>
      {/if}
    </nav>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    {#if selected}
      {#key selected.id}
        <WorldPage note={selected} onsaved={saved} />
      {/key}

      <div class="structure">
        <button class="btn ghost" onclick={() => shift(-1)} disabled={busy || first} title="Move up among its siblings">↑</button>
        <button class="btn ghost" onclick={() => shift(1)} disabled={busy || last} title="Move down among its siblings">↓</button>
        <select bind:value={moveTo} onchange={move} disabled={busy} aria-label="Move this page">
          <option value="">Move to…</option>
          {#if selected.parentId !== null}
            <option value="top">The world (top level)</option>
          {/if}
          {#each moveTargets as t (t.note.id)}
            <option value={String(t.note.id)}>{'  '.repeat(t.depth)}{t.note.title}</option>
          {/each}
        </select>
        <span class="grow"></span>
        <button class="btn danger" onclick={remove} disabled={busy}>Delete</button>
      </div>
    {:else if !loading}
      <p class="intro">
        Your campaign's world, as nested pages. Put cities inside regions and places inside
        cities; a timeline lays the events inside it out in order. Everything here is
        markdown, saved as you type, and found by Search.
      </p>
    {/if}

    <div class="children">
      <header>
        <h3>
          {#if selected}Inside {selected.title}{:else}Top level{/if}
          {#if children.length}<span class="count">{children.length}</span>{/if}
        </h3>
        {#if !creating}
          <button class="btn ghost" onclick={startCreate} disabled={busy}>
            + Add {selected ? 'a page inside' : 'a page'}
          </button>
        {/if}
      </header>

      {#if creating}
        <form onsubmit={create}>
          <label class="narrow">
            <span>Kind</span>
            <select bind:value={creating.kind}>
              {#each WORLD_KINDS as kind (kind.value)}
                <option value={kind.value} title={kind.hint}>{kind.label}</option>
              {/each}
            </select>
          </label>
          <label>
            <span>Title</span>
            <!-- svelte-ignore a11y_autofocus -->
            <input bind:value={creating.title} placeholder={WORLD_KINDS.find((k) => k.value === creating.kind)?.hint} autofocus required />
          </label>
          <button class="btn primary" type="submit" disabled={busy || !creating.title.trim()}>Create</button>
          <button class="btn ghost" type="button" onclick={() => (creating = null)}>Cancel</button>
        </form>
      {/if}

      {#if !children.length}
        {#if !creating && !loading}
          <p class="empty">
            {#if selected?.kind === 'timeline'}No events yet — add one and give it a “when”.{:else if selected}Nothing inside this page yet.{:else}Start with a region, a city or a timeline.{/if}
          </p>
        {/if}
      {:else if selected?.kind === 'timeline'}
        <ol class="timeline">
          {#each children as child (child.id)}
            <li>
              <button onclick={() => select(child.id)}>
                <span class="when">{child.worldDate || '—'}</span>
                <span class="what">
                  <span class="ctitle">{child.title}</span>
                  {#if excerpt(child.body)}<span class="excerpt">{excerpt(child.body)}</span>{/if}
                </span>
              </button>
            </li>
          {/each}
        </ol>
      {:else}
        <ul class="cards">
          {#each children as child (child.id)}
            <li>
              <button onclick={() => select(child.id)}>
                <span class="top">
                  <span class="ctitle">{child.title}</span>
                  <span class="chip">{worldKindLabel(child.kind)}</span>
                </span>
                {#if child.worldDate}<span class="date">{child.worldDate}</span>{/if}
                {#if excerpt(child.body)}<span class="excerpt">{excerpt(child.body)}</span>{/if}
                {#if kids(child.id).length}
                  <span class="nested">{kids(child.id).length} inside</span>
                {/if}
              </button>
            </li>
          {/each}
        </ul>
      {/if}
    </div>
  </div>
</section>

<style>
  .world {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
  }

  .tree {
    display: flex;
    flex: 0 0 16rem;
    flex-direction: column;
    gap: 0.4rem;
    min-width: 0;
    padding: 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .treehead {
    display: flex;
    gap: 0.35rem;
  }

  .treehead input {
    flex: 1;
    min-width: 0;
    font-size: 0.8rem;
  }

  .tree ul {
    display: flex;
    flex-direction: column;
    gap: 0.05rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .tree li {
    display: flex;
    align-items: center;
    gap: 0.2rem;
    border-radius: 6px;
  }

  .tree li:hover { background: var(--panel-2); }

  .tree li.on { background: var(--panel-2); }

  .tree li.on .name { color: var(--fear); }

  .rootrow,
  .name {
    flex: 1;
    min-width: 0;
    padding: 0.25rem 0.3rem;
    border: none;
    background: none;
    color: var(--text);
    font: inherit;
    font-size: 0.82rem;
    text-align: left;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
    cursor: pointer;
  }

  .rootrow {
    flex: 0 0 auto;
    border-radius: 6px;
    font-size: 0.72rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .rootrow:hover { color: var(--text); }
  .rootrow.on { color: var(--fear); }

  .caret {
    flex: 0 0 1rem;
    padding: 0;
    border: none;
    background: none;
    color: var(--muted);
    font: inherit;
    font-size: 0.85rem;
    line-height: 1;
    cursor: pointer;
    transition: transform 0.12s ease;
  }

  .caret:disabled { cursor: default; }
  .caret.open { transform: rotate(90deg); }
  .caret.spacer { display: inline-block; }

  .kind {
    flex: 0 0 auto;
    padding-right: 0.3rem;
    font-size: 0.65rem;
    color: var(--muted);
    opacity: 0.8;
  }

  .page {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.75rem;
    min-width: 0;
  }

  .trail {
    display: flex;
    align-items: center;
    gap: 0.3rem;
    flex-wrap: wrap;
    font-size: 0.78rem;
  }

  .crumb {
    padding: 0;
    border: none;
    background: none;
    color: var(--muted);
    font: inherit;
    cursor: pointer;
  }

  .crumb:hover { color: var(--text); }

  .crumb.current {
    color: var(--text);
    cursor: default;
  }

  .sep { color: var(--muted); opacity: 0.6; }

  .error {
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }

  .intro {
    margin: 0;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .structure {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .structure select {
    flex: 0 1 16rem;
    min-width: 0;
    font-size: 0.8rem;
  }

  .grow { flex: 1; }

  .children {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding-top: 0.6rem;
    border-top: 1px solid var(--line);
  }

  .children header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  h3 {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .count {
    font-size: 0.7rem;
    opacity: 0.8;
  }

  form {
    display: flex;
    align-items: flex-end;
    gap: 0.5rem;
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  form label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
  }

  form label.narrow { flex: 0 0 9rem; }

  form label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .cards {
    display: grid;
    grid-template-columns: repeat(auto-fill, minmax(14rem, 1fr));
    gap: 0.45rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .cards button,
  .timeline button {
    width: 100%;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .cards button:hover,
  .timeline button:hover { border-color: var(--muted); }

  .cards button {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
    height: 100%;
    padding: 0.55rem 0.7rem;
  }

  .top {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.4rem;
  }

  .ctitle {
    min-width: 0;
    overflow: hidden;
    font-size: 0.88rem;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .date,
  .nested {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .date { color: var(--gold); }

  .excerpt {
    display: -webkit-box;
    overflow: hidden;
    -webkit-box-orient: vertical;
    -webkit-line-clamp: 2;
    line-clamp: 2;
    font-size: 0.78rem;
    line-height: 1.4;
    color: var(--muted);
  }

  .timeline {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    margin: 0;
    padding: 0 0 0 0.6rem;
    border-left: 2px solid var(--line);
    list-style: none;
  }

  .timeline li { position: relative; }

  .timeline li::before {
    content: '';
    position: absolute;
    top: 0.85rem;
    left: calc(-0.6rem - 5px);
    width: 8px;
    height: 8px;
    border-radius: 50%;
    background: var(--fear);
  }

  .timeline button {
    display: flex;
    align-items: baseline;
    gap: 0.75rem;
    padding: 0.5rem 0.7rem;
  }

  .when {
    flex: 0 0 9rem;
    font-size: 0.78rem;
    color: var(--gold);
  }

  .what {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;
  }

  @media (prefers-reduced-motion: reduce) {
    .caret { transition: none; }
  }
</style>
