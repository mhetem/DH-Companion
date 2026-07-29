<script>
  import Modal from '../Modal.svelte'
  import { Rest, RestAllowance, RestMoves, errorMessage } from './api.js'

  let { open = $bindable(false), long = false, characterId, onrested } = $props()

  let moves = $state([])
  let allowed = $state(2)
  let loading = $state(false)
  let saving = $state(false)
  let error = $state('')

  // A move can be taken more than once — two goes at Tend to Wounds is a legal short
  // rest — so this counts picks per move rather than being a set of checkboxes.
  let counts = $state({})

  const taken = $derived(Object.values(counts).reduce((total, n) => total + n, 0))
  const room = $derived(allowed - taken)

  $effect(() => {
    if (!open) return
    const isLong = long
    loading = true
    error = ''
    counts = {}
    Promise.all([RestMoves(isLong), RestAllowance(isLong)])
      .then(([list, n]) => {
        moves = list ?? []
        allowed = n
      })
      .catch((e) => (error = errorMessage(e)))
      .finally(() => (loading = false))
  })

  function add(id) {
    if (room <= 0) return
    counts = { ...counts, [id]: (counts[id] ?? 0) + 1 }
  }

  function drop(id) {
    const next = (counts[id] ?? 0) - 1
    counts = { ...counts, [id]: Math.max(0, next) }
  }

  async function submit() {
    saving = true
    error = ''
    try {
      const picked = []
      for (const move of moves) {
        for (let i = 0; i < (counts[move.id] ?? 0); i += 1) picked.push(move.id)
      }
      const result = await Rest({ characterId, long, moves: picked })
      open = false
      onrested?.(result)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }
</script>

<Modal bind:open label={long ? 'Long rest' : 'Short rest'} width="34rem">
  <header>
    <h2>{long ? 'Long rest' : 'Short rest'}</h2>
    <p class="sub">
      Choose your downtime moves — <span class:ok={taken === allowed}>{taken}/{allowed}</span> picked.
      You can take the same move twice.
    </p>
  </header>

  {#if error}<p class="error">{error}</p>{/if}

  {#if loading}
    <p class="empty">Loading downtime moves…</p>
  {:else}
    <ul class="moves">
      {#each moves as move (move.id)}
        {@const n = counts[move.id] ?? 0}
        <li class:picked={n > 0}>
          <div class="what">
            <span class="label">{move.label}</span>
            <span class="desc">{move.description}</span>
          </div>
          <div class="count">
            <button class="btn ghost" disabled={n === 0} onclick={() => drop(move.id)}>−</button>
            <span class="n">{n}</span>
            <button class="btn ghost" disabled={room <= 0} onclick={() => add(move.id)}>+</button>
          </div>
        </li>
      {/each}
    </ul>
  {/if}

  <div class="foot">
    <button class="btn ghost" onclick={() => (open = false)}>Cancel</button>
    <span class="spacer"></span>
    <button class="btn primary" disabled={saving || taken === 0} onclick={submit}>
      Take {taken === 1 ? 'the move' : `${taken} moves`}
    </button>
  </div>
</Modal>

<style>
  header { margin-bottom: -0.4rem; }

  h2 {
    margin: 0;
    font-size: 1.2rem;
  }

  .sub {
    margin: 0.2rem 0 0;
    font-size: 0.8rem;
    color: var(--muted);
  }

  .sub span { color: var(--gold); }
  .sub span.ok { color: var(--hope); }

  .moves {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .moves li {
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.5rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
  }

  .moves li.picked { border-color: var(--hope); }

  .what {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-width: 0;
  }

  .label { font-size: 0.88rem; }

  .desc {
    font-size: 0.74rem;
    color: var(--muted);
  }

  .count {
    display: flex;
    align-items: center;
    gap: 0.35rem;
  }

  .count .n {
    min-width: 1.2rem;
    font-size: 0.9rem;
    text-align: center;
  }

  .error {
    margin: 0;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .foot {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .spacer { flex: 1; }
</style>
