<script>
  import EmptyState from '../EmptyState.svelte'
  import FeatureList from '../gm/FeatureList.svelte'
  import SrdText from '../SrdText.svelte'
  import { active } from './active.svelte.js'
  import { Beastforms, DropBeastform, Transform, errorMessage } from './api.js'

  let view = $state(null)
  let loading = $state(true)
  let error = $state('')
  let busy = $state(false)
  let selectedSlug = $state('')
  let tierFilter = $state('All')

  const id = $derived(active.id)

  $effect(() => {
    const target = id
    if (!target) {
      view = null
      loading = false
      return
    }
    loading = true
    Beastforms(target)
      .then((v) => {
        view = v
        selectedSlug = v.active?.slug ?? v.available[0]?.slug ?? ''
        error = ''
      })
      .catch((e) => {
        error = errorMessage(e)
        view = null
      })
      .finally(() => (loading = false))
  })

  const tiers = $derived(['All', ...new Set((view?.available ?? []).map((f) => f.tier))])

  const listed = $derived(
    (view?.available ?? []).filter((f) => tierFilter === 'All' || f.tier === tierFilter)
  )

  const selected = $derived((view?.available ?? []).find((f) => f.slug === selectedSlug) ?? null)

  const isActive = $derived(view?.active?.slug === selectedSlug)

  async function run(fn) {
    if (busy) return
    busy = true
    try {
      view = await fn()
      error = ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = false
    }
  }
</script>

{#if !id}
  <div class="pane">
    <EmptyState title="No character open">Pick one in Characters first.</EmptyState>
  </div>
{:else if loading}
  <div class="pane"><p class="empty">Loading Beastform options…</p></div>
{:else if !view}
  <div class="pane"><p class="error">{error}</p></div>
{:else}
  <div class="pane">
    <header>
      <div class="intro">
        <h2>Beastform</h2>
        <p class="blurb">
          {#if view.eligible}
            Mark a Stress to transform into a creature of your tier or lower. You stay in the form
            until you drop out of it — dropping out costs nothing.
          {:else}
            Beastform is the druid's class feature. {view.character.name} isn't a druid, so this is
            here as reference only.
          {/if}
        </p>
      </div>
      {#if view.active}
        <div class="active-badge">
          <span class="l">Transformed</span>
          <span class="n">{view.active.name}</span>
          <span class="stat">Evasion {view.evasion} ({view.baseEvasion} + {view.active.evasionBonus})</span>
          {#if view.traitBonus}<span class="stat">{view.traitBonus}</span>{/if}
          <button class="btn ghost" disabled={busy} onclick={() => run(() => DropBeastform(id))}>
            Drop out
          </button>
        </div>
      {/if}
    </header>

    {#if error}<p class="error">{error}</p>{/if}

    <div class="split">
      <div class="list-side">
        <div class="chips">
          {#each tiers as tier (tier)}
            <button class="chip" class:on={tierFilter === tier} onclick={() => (tierFilter = tier)}>
              {tier === 'All' ? 'All tiers' : `Tier ${tier}`}
            </button>
          {/each}
        </div>
        <ul class="forms">
          {#each listed as form (form.slug)}
            <li>
              <button
                class="face"
                class:picked={selectedSlug === form.slug}
                class:current={view.active?.slug === form.slug}
                onclick={() => (selectedSlug = form.slug)}
              >
                <span class="name">{form.name}</span>
                <span class="meta">Tier {form.tier} · {form.trait} {form.traitBonus} · Evasion {form.evasionBonus}</span>
                <span class="examples">{form.examples.join(', ')}</span>
              </button>
            </li>
          {/each}
          {#if !listed.length}
            <li class="empty">No forms at this tier.</li>
          {/if}
        </ul>
      </div>

      <div class="detail">
        {#if selected}
          <div class="detail-head">
            <div>
              <h3>{selected.name}</h3>
              <p class="meta">Tier {selected.tier} · {selected.examples.join(', ')}</p>
            </div>
            {#if view.eligible}
              {#if isActive}
                <button class="btn ghost" disabled={busy} onclick={() => run(() => DropBeastform(id))}>
                  Drop out
                </button>
              {:else}
                <div class="transform">
                  <button class="btn primary" disabled={busy} onclick={() => run(() => Transform(id, selected.slug, true))}>
                    Transform (mark a Stress)
                  </button>
                  <button class="btn ghost" disabled={busy} onclick={() => run(() => Transform(id, selected.slug, false))}>
                    Without marking
                  </button>
                </div>
              {/if}
            {/if}
          </div>

          <SrdText text={selected.description} />

          <div class="stats">
            <div class="stat"><span class="n">{selected.trait} {selected.traitBonus}</span><span class="l">Trait bonus</span></div>
            <div class="stat"><span class="n">{selected.evasionBonus}</span><span class="l">Evasion</span></div>
            <div class="stat"><span class="n">{selected.attack.damage} {selected.attack.damageType}</span><span class="l">{selected.attack.range} attack</span></div>
            <div class="stat"><span class="n">{selected.attack.trait}</span><span class="l">Attack trait</span></div>
          </div>

          {#if selected.advantages?.length}
            <h4>Advantage on</h4>
            <p class="advantages">{selected.advantages.join(', ')}</p>
          {/if}

          <h4>Features</h4>
          <FeatureList features={selected.features} />
        {:else}
          <p class="empty">Pick a form to read it.</p>
        {/if}
      </div>
    </div>
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

  header .intro { flex: 1; }

  h2 {
    margin: 0;
    font-size: 1.25rem;
  }

  h3 {
    margin: 0;
    font-size: 1.05rem;
  }

  h4 {
    margin: 1rem 0 0.4rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .blurb {
    margin: 0.25rem 0 0;
    max-width: 44rem;
    font-size: 0.85rem;
    color: var(--muted);
  }

  .active-badge {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
    gap: 0.2rem;
    padding: 0.55rem 0.75rem;
    border: 1px solid var(--hope);
    border-radius: 10px;
    background: var(--panel);
  }

  .active-badge .l {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--muted);
  }

  .active-badge .n {
    font-size: 1rem;
    color: var(--hope);
  }

  .active-badge .stat {
    font-size: 0.72rem;
    color: var(--gold);
  }

  .active-badge button { margin-top: 0.35rem; }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .split {
    display: grid;
    grid-template-columns: minmax(13rem, 18rem) 1fr;
    gap: 1rem;
    align-items: start;
  }

  .list-side {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
    min-width: 0;
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.25rem;
  }

  .chip {
    padding: 0.2rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel-2);
    color: var(--muted);
    font: inherit;
    font-size: 0.72rem;
    cursor: pointer;
  }

  .chip.on {
    border-color: var(--hope);
    color: var(--text);
  }

  .forms {
    display: flex;
    flex-direction: column;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
    max-height: 34rem;
    overflow-y: auto;
  }

  .face {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    width: 100%;
    padding: 0.5rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .face:hover { border-color: var(--gold-deep); }
  .face.picked { border-color: var(--gold); background: var(--panel-2); }
  .face.current { border-color: var(--hope); }

  .name { font-size: 0.88rem; }

  .meta {
    margin: 0;
    font-size: 0.72rem;
    color: var(--muted);
  }

  .examples {
    font-size: 0.68rem;
    font-style: italic;
    color: var(--muted);
    opacity: 0.85;
  }

  .detail {
    max-width: 44rem;
    min-width: 0;
    padding: 0.9rem 1rem;
    border: 1px solid var(--line);
    border-radius: 10px;
    background: var(--panel);
  }

  .detail-head {
    display: flex;
    align-items: flex-start;
    justify-content: space-between;
    gap: 1rem;
    margin-bottom: 0.5rem;
  }

  .transform {
    display: flex;
    flex-wrap: wrap;
    gap: 0.35rem;
    justify-content: flex-end;
  }

  .stats {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(7rem, 1fr));
    gap: 0.5rem;
    margin-top: 0.75rem;
  }

  .stat {
    display: flex;
    flex-direction: column;
    align-items: center;
    gap: 0.1rem;
    padding: 0.5rem 0.3rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
    text-align: center;
  }

  .stat .n {
    font-size: 0.95rem;
    color: var(--gold);
  }

  .stat .l {
    font-size: 0.62rem;
    text-transform: uppercase;
    letter-spacing: 0.05em;
    color: var(--muted);
  }

  .advantages {
    margin: 0;
    font-size: 0.85rem;
    color: var(--muted);
  }
</style>
