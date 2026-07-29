<script>
  import EmptyState from '../EmptyState.svelte'
  import Gold from './Gold.svelte'
  import { active } from './active.svelte.js'
  import { goldLabel } from './api.js'
  import {
    AddClassItems,
    AdjustItemQty,
    DeleteItem,
    GetCharacter,
    ITEM_KINDS,
    ListInventory,
    SaveItem,
    SetItemEquipped,
    errorMessage,
    itemKindLabel
  } from './api.js'

  let character = $state(null)
  let items = $state([])
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let editingId = $state(null)
  let form = $state(blank())

  const id = $derived(active.id)

  function blank() {
    return { name: '', kind: 'item', qty: 1, equipped: false, detail: '' }
  }

  $effect(() => {
    const target = id
    if (!target) {
      items = []
      character = null
      loading = false
      return
    }
    loading = true
    Promise.all([GetCharacter(target), ListInventory(target)])
      .then(([c, list]) => {
        character = c
        items = list ?? []
        error = ''
      })
      .catch((e) => (error = errorMessage(e)))
      .finally(() => (loading = false))
  })

  const grouped = $derived(
    ITEM_KINDS.map((kind) => ({ ...kind, items: items.filter((i) => i.kind === kind.value) })).filter(
      (g) => g.items.length
    )
  )

  async function run(fn) {
    if (busy) return
    busy = true
    try {
      items = (await fn()) ?? []
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }

  function reset() {
    editingId = null
    form = blank()
  }

  function edit(item) {
    editingId = item.id
    form = { name: item.name, kind: item.kind, qty: item.qty, equipped: item.equipped, detail: item.detail }
  }

  async function submit(event) {
    event.preventDefault()
    await run(() =>
      SaveItem({
        id: editingId,
        characterId: id,
        name: form.name,
        kind: form.kind,
        qty: Number(form.qty),
        equipped: form.equipped,
        detail: form.detail
      })
    )
    if (!error) reset()
  }

  function remove(item) {
    if (!confirm(`Delete “${item.name}”?`)) return
    if (editingId === item.id) reset()
    return run(() => DeleteItem(item.id))
  }
</script>

{#if !id}
  <div class="pane">
    <EmptyState title="No character open">Pick one in Characters first.</EmptyState>
  </div>
{:else}
  <div class="pane">
    <header>
      <div>
        <h2>Inventory</h2>
        <p class="blurb">
          Weapons and armor are one-at-a-time — equipping a second unequips the first.
        </p>
      </div>
      <button class="btn ghost" disabled={busy || !character?.classSlug} onclick={() => run(() => AddClassItems(id))}>
        Add {character?.className || 'class'} items
      </button>
    </header>

    {#if error}<p class="error">{error}</p>{/if}

    {#if character}
      <section class="purse-card">
        <div class="purse-head">
          <h3>Gold</h3>
          <span class="carrying">{goldLabel(character.gold)}</span>
        </div>
        <Gold {character} {busy} onupdate={(c) => (character = c)} />
        <p class="hint">Ten handfuls make a bag; ten bags make a chest.</p>
      </section>
    {/if}

    <form onsubmit={submit}>
      <label class="grow">
        <span>Name</span>
        <input bind:value={form.name} placeholder="Shortsword" required />
      </label>
      <label>
        <span>Kind</span>
        <select bind:value={form.kind}>
          {#each ITEM_KINDS as kind (kind.value)}
            <option value={kind.value}>{kind.label}</option>
          {/each}
        </select>
      </label>
      <label class="narrow">
        <span>Qty</span>
        <input type="number" min="0" max="999" bind:value={form.qty} />
      </label>
      <label class="grow">
        <span>Detail</span>
        <input bind:value={form.detail} placeholder="Melee · d8+3 phy" />
      </label>
      <label class="check">
        <input type="checkbox" bind:checked={form.equipped} />
        <span>Equipped</span>
      </label>
      <button class="btn primary" type="submit" disabled={busy || !form.name.trim()}>
        {editingId === null ? 'Add' : 'Save'}
      </button>
      {#if editingId !== null}
        <button class="btn ghost" type="button" onclick={reset}>Cancel</button>
      {/if}
    </form>

    {#if loading}
      <p class="empty">Loading…</p>
    {:else if !items.length}
      <EmptyState title="Nothing carried yet">
        Add gear above, or pull in your class items with the button in the header.
      </EmptyState>
    {:else}
      {#each grouped as group (group.value)}
        <section>
          <h3>{group.label}</h3>
          <ul>
            {#each group.items as item (item.id)}
              <li class:equipped={item.equipped} class:editing={editingId === item.id}>
                <div class="what">
                  <span class="name">{item.name}</span>
                  {#if item.detail}<span class="detail">{item.detail}</span>{/if}
                </div>
                <div class="qty">
                  <button class="btn ghost" disabled={busy} onclick={() => run(() => AdjustItemQty(item.id, -1))}>−</button>
                  <span class="n">{item.qty}</span>
                  <button class="btn ghost" disabled={busy} onclick={() => run(() => AdjustItemQty(item.id, 1))}>+</button>
                </div>
                <button
                  class="btn ghost"
                  disabled={busy}
                  onclick={() => run(() => SetItemEquipped(item.id, !item.equipped))}
                >
                  {item.equipped ? 'Unequip' : 'Equip'}
                </button>
                <button class="btn ghost" onclick={() => edit(item)}>Edit</button>
                <button class="btn danger" onclick={() => remove(item)}>✕</button>
              </li>
            {/each}
          </ul>
        </section>
      {/each}
    {/if}
  </div>
{/if}

<style>
  .pane {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
    overflow-y: auto;
  }

  header {
    display: flex;
    align-items: flex-start;
    gap: 1rem;
    margin-bottom: 1rem;
  }

  header div { flex: 1; }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  .purse-card {
    max-width: 26rem;
    margin-bottom: 0.85rem;
    padding: 0.75rem 0.9rem;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: var(--panel);
  }

  .purse-head {
    display: flex;
    align-items: baseline;
    justify-content: space-between;
    gap: 0.75rem;
  }

  .purse-card h3 { margin: 0 0 0.6rem; }

  .carrying {
    font-size: 0.75rem;
    color: var(--gold);
  }

  .hint {
    margin: 0.6rem 0 0;
    font-size: 0.72rem;
    color: var(--muted);
  }

  h3 {
    margin: 1rem 0 0.4rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .blurb {
    margin: 0.25rem 0 0;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  form {
    display: flex;
    flex-wrap: wrap;
    align-items: flex-end;
    gap: 0.6rem;
    padding: 0.85rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.grow { flex: 1 1 12rem; }
  label.narrow { flex: 0 0 5rem; }

  label.check {
    flex-direction: row;
    align-items: center;
    gap: 0.35rem;
  }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  ul {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  li {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding: 0.45rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  li.equipped { border-color: var(--hope); }
  li.editing { border-color: var(--gold); }

  .what {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .name { font-size: 0.88rem; }

  .detail {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .qty {
    display: flex;
    align-items: center;
    gap: 0.3rem;
  }

  .qty .n {
    min-width: 1.5rem;
    font-size: 0.85rem;
    text-align: center;
  }
</style>
