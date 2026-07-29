<script>
  import EmptyState from '../EmptyState.svelte'
  import CharacterWizard from './CharacterWizard.svelte'
  import { active, setActive } from './active.svelte.js'
  import { DeleteCharacter, ListCharacters, errorMessage, goldLabel, signed } from './api.js'

  let characters = $state([])
  let loading = $state(true)
  let error = $state('')

  // null = the list; an object = the wizard, editing that character; 'new' = creating.
  let editing = $state(null)

  async function refresh() {
    try {
      characters = (await ListCharacters()) ?? []
      error = ''
      // A remembered character that has since been deleted must not stay selected.
      if (active.id && !characters.some((c) => c.id === active.id)) {
        setActive(characters[0]?.id ?? null)
      }
      if (!active.id && characters.length) setActive(characters[0].id)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  refresh()

  function saved(character) {
    editing = null
    setActive(character.id)
    refresh()
  }

  async function remove(character) {
    if (!confirm(`Delete ${character.name}? Their cards, gear and Experiences go with them.`)) return
    try {
      await DeleteCharacter(character.id)
      if (active.id === character.id) setActive(null)
      await refresh()
    } catch (e) {
      error = errorMessage(e)
    }
  }
</script>

{#if editing !== null}
  <CharacterWizard
    character={editing === 'new' ? null : editing}
    onsaved={saved}
    oncancel={() => (editing = null)}
  />
{:else}
  <div class="characters">
    <header>
      <div>
        <h2>Characters</h2>
        <p class="blurb">The character you open here is the one every other pane works on.</p>
      </div>
      <button class="btn primary" onclick={() => (editing = 'new')}>New character</button>
    </header>

    {#if error}
      <p class="error">{error}</p>
    {/if}

    {#if loading}
      <p class="empty">Loading…</p>
    {:else if !characters.length}
      <EmptyState title="No characters yet">
        Build one and the sheet, loadout, inventory and dice all follow.
      </EmptyState>
    {:else}
      <ul class="list">
        {#each characters as character (character.id)}
          <li class:active={active.id === character.id}>
            <button class="open" onclick={() => setActive(character.id)}>
              <span class="name">{character.name}</span>
              <span class="meta">
                Level {character.level} · {character.className || '—'}
                {#if character.subclassName}({character.subclassName}){/if}
                · {character.ancestryName || '—'} · {character.communityName || '—'}
              </span>
              <span class="line">
                {#each Object.entries(character.traits) as [key, value] (key)}
                  <span class="trait">{key.slice(0, 3)} {signed(value)}</span>
                {/each}
              </span>
              <span class="line muted">
                HP {character.hpMarked}/{character.hpMax} · Stress {character.stressMarked}/{character.stressMax}
                · Hope {character.hope}/{character.hopeMax} · {goldLabel(character.gold)}
              </span>
            </button>
            <div class="actions">
              <button class="btn ghost" onclick={() => (editing = character)}>Edit</button>
              <button class="btn danger" onclick={() => remove(character)}>Delete</button>
            </div>
          </li>
        {/each}
      </ul>
    {/if}
  </div>
{/if}

<style>
  .characters {
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

  .list {
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .list li {
    display: flex;
    align-items: center;
    gap: 0.6rem;
    margin-bottom: 0.45rem;
    padding: 0.15rem 0.6rem 0.15rem 0.15rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
  }

  .list li.active { border-color: var(--hope); }

  .open {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.15rem;
    min-width: 0;
    padding: 0.5rem 0.6rem;
    border: none;
    background: none;
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .name { font-size: 0.95rem; }

  .meta {
    font-size: 0.75rem;
    color: var(--muted);
  }

  .line {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
    font-size: 0.72rem;
    color: var(--gold);
  }

  .line.muted { color: var(--muted); }

  .trait { text-transform: capitalize; }

  .actions {
    display: flex;
    gap: 0.4rem;
  }
</style>
