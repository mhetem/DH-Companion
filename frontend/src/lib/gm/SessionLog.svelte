<script>
  import {
    DeleteSession,
    GetSession,
    LinkEncounter,
    ListEncounters,
    ListSessions,
    SaveSession,
    UnlinkEncounter,
    errorMessage
  } from './api.js'

  let { campaignId } = $props()

  let sessions = $state([])
  let encounters = $state([])
  let loading = $state(true)
  let error = $state('')
  let saving = $state(false)
  let openId = $state(null)
  let detail = $state(null)
  let editingId = $state(null)
  let form = $state(blank())
  let linkPick = $state('')

  function blank() {
    return { number: '', title: '', date: today(), recap: '' }
  }

  function today() {
    return new Date().toISOString().slice(0, 10)
  }

  // The backend stores a full timestamp; the date input only speaks YYYY-MM-DD.
  const dateOnly = (value) => (value ?? '').slice(0, 10)

  async function refresh() {
    try {
      const [rows, enc] = await Promise.all([ListSessions(campaignId), ListEncounters()])
      sessions = rows ?? []
      encounters = enc ?? []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

  async function toggle(session) {
    if (openId === session.id) {
      openId = null
      detail = null
      return
    }
    openId = session.id
    detail = null
    await reopen(session.id)
  }

  function edit(session) {
    editingId = session.id
    form = {
      number: session.number,
      title: session.title,
      date: dateOnly(session.date),
      recap: session.recap
    }
  }

  function reset() {
    editingId = null
    form = blank()
  }

  async function submit(event) {
    event.preventDefault()
    saving = true
    try {
      await SaveSession({
        id: editingId,
        campaignId,
        number: Number(form.number) || 0,
        title: form.title,
        date: form.date,
        recap: form.recap
      })
      reset()
      await refresh()
      if (openId) await reopen(openId)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }

  async function reopen(id) {
    try {
      detail = await GetSession(id)
      error = ''
    } catch (e) {
      error = errorMessage(e)
    }
  }

  async function remove(session) {
    if (!confirm(`Delete session ${session.number} — “${session.title}”?`)) return
    try {
      await DeleteSession(session.id)
      if (editingId === session.id) reset()
      if (openId === session.id) {
        openId = null
        detail = null
      }
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    }
  }

  async function link(session) {
    if (!linkPick) return
    try {
      detail = await LinkEncounter(session.id, Number(linkPick))
      linkPick = ''
      error = ''
    } catch (e) {
      error = errorMessage(e)
    }
  }

  async function unlink(session, encounterId) {
    try {
      detail = await UnlinkEncounter(session.id, encounterId)
      error = ''
    } catch (e) {
      error = errorMessage(e)
    }
  }

  const unlinked = $derived(
    encounters.filter((e) => !(detail?.encounters ?? []).some((x) => x.id === e.id))
  )
</script>

<section class="sessions">
  <header>
    <h3>Session log</h3>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  <form onsubmit={submit}>
    <div class="row">
      <label class="tiny">
        <span>#</span>
        <input type="number" min="1" bind:value={form.number} placeholder="auto" title="Leave blank to take the next number" />
      </label>
      <label>
        <span>Title</span>
        <input bind:value={form.title} placeholder="Into the Mire" />
      </label>
      <label class="narrow">
        <span>Date</span>
        <input type="date" bind:value={form.date} />
      </label>
    </div>
    <label>
      <span>Recap</span>
      <textarea bind:value={form.recap} rows="3" placeholder="What happened, who they met, what they owe."></textarea>
    </label>
    <div class="actions">
      <button class="btn primary" type="submit" disabled={saving}>
        {editingId === null ? 'Log session' : 'Save changes'}
      </button>
      {#if editingId !== null}
        <button class="btn ghost" type="button" onclick={reset}>Cancel</button>
      {/if}
    </div>
  </form>

  {#if loading}
    <p class="empty">Loading…</p>
  {:else if !sessions.length}
    <p class="empty">No sessions logged yet. The number fills itself in if you leave it blank.</p>
  {:else}
    <ul class="list">
      {#each sessions as session (session.id)}
        <li class:editing={editingId === session.id}>
          <div class="top">
            <button class="disclose" onclick={() => toggle(session)}>
              <span class="caret" class:open={openId === session.id}>›</span>
              <span class="num">#{session.number}</span>
              <span class="name">{session.title}</span>
            </button>
            <span class="date">{dateOnly(session.date)}</span>
            <button class="btn ghost" onclick={() => edit(session)}>Edit</button>
            <button class="btn danger" onclick={() => remove(session)} title="Delete">✕</button>
          </div>

          {#if openId === session.id}
            <div class="detail">
              {#if session.recap.trim()}
                <p class="recap">{session.recap}</p>
              {:else}
                <p class="empty">No recap written.</p>
              {/if}

              <div class="linked">
                <span class="tag">Encounters run</span>
                {#if !detail}
                  <p class="empty">Loading…</p>
                {:else if !detail.encounters.length}
                  <p class="empty">None linked yet.</p>
                {:else}
                  <ul class="chips">
                    {#each detail.encounters as enc (enc.id)}
                      <li>
                        <span class="chip">{enc.name || 'Untitled encounter'} · {enc.totalCount}</span>
                        <button class="btn danger" onclick={() => unlink(session, enc.id)} title="Unlink">✕</button>
                      </li>
                    {/each}
                  </ul>
                {/if}

                {#if unlinked.length}
                  <div class="linkrow">
                    <select bind:value={linkPick}>
                      <option value="">Link an encounter…</option>
                      {#each unlinked as enc (enc.id)}
                        <option value={enc.id}>{enc.name || 'Untitled encounter'}</option>
                      {/each}
                    </select>
                    <button class="btn ghost" onclick={() => link(session)} disabled={!linkPick}>Link</button>
                  </div>
                {/if}
              </div>

              <div class="linked">
                <span class="tag">Fights run</span>
                {#if !detail}
                  <p class="empty">Loading…</p>
                {:else if !detail.combats.length}
                  <p class="empty">No combats logged to this session yet — pick it in the Combat Runner when you start one.</p>
                {:else}
                  <ul class="fights">
                    {#each detail.combats as fight (fight.id)}
                      <li>
                        <span class="fname">{fight.encounterName || 'Untitled encounter'}</span>
                        <span class="fmeta">{fight.createdAt.slice(0, 10)} · Fear {fight.fear}{fight.active ? ' · running' : ''}</span>
                      </li>
                    {/each}
                  </ul>
                {/if}
              </div>
            </div>
          {/if}
        </li>
      {/each}
    </ul>
  {/if}
</section>

<style>
  .sessions {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
  }

  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
  }

  h3 {
    margin: 0;
    font-size: 0.8rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .error {
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }

  form {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .row {
    display: flex;
    gap: 0.6rem;
  }

  label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.tiny { flex: 0 0 4.5rem; }
  label.narrow { flex: 0 0 9rem; }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .actions {
    display: flex;
    gap: 0.5rem;
  }

  .list {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .list li {
    padding: 0.5rem 0.7rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .list li.editing { border-color: var(--fear); }

  .top {
    display: flex;
    align-items: center;
    gap: 0.4rem;
  }

  .disclose {
    display: flex;
    flex: 1;
    align-items: center;
    gap: 0.35rem;
    min-width: 0;
    padding: 0;
    border: none;
    background: none;
    color: var(--text);
    font: inherit;
    font-size: 0.85rem;
    text-align: left;
    cursor: pointer;
  }

  .caret {
    display: inline-block;
    color: var(--muted);
    transition: transform 0.12s ease;
  }

  .caret.open { transform: rotate(90deg); }

  .num {
    color: var(--fear);
    font-variant-numeric: tabular-nums;
  }

  .name {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
  }

  .date {
    font-size: 0.75rem;
    color: var(--muted);
    font-variant-numeric: tabular-nums;
  }

  .detail {
    margin-top: 0.5rem;
    padding-top: 0.5rem;
    border-top: 1px solid var(--line);
  }

  .recap {
    margin: 0 0 0.6rem;
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
    white-space: pre-wrap;
  }

  .tag {
    display: block;
    margin-bottom: 0.3rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    margin: 0 0 0.5rem;
    padding: 0;
    list-style: none;
  }

  .chips li {
    display: flex;
    align-items: center;
    gap: 0.15rem;
    padding: 0;
    border: none;
    background: none;
  }

  .fights {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    margin: 0 0 0.5rem;
    padding: 0;
    list-style: none;
  }

  .fights li {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    padding: 0;
    border: none;
    background: none;
  }

  .fname { font-size: 0.85rem; }

  .fmeta {
    font-size: 0.75rem;
    color: var(--muted);
  }

  .linkrow {
    display: flex;
    gap: 0.35rem;
  }

  .linkrow select { flex: 1; min-width: 0; }

  .empty {
    margin: 0;
    font-size: 0.8rem;
    color: var(--muted);
  }

  @media (prefers-reduced-motion: reduce) {
    .caret { transition: none; }
  }
</style>
