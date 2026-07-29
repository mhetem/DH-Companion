<script>
  import FeatureList from '../gm/FeatureList.svelte'
  import SrdText from '../SrdText.svelte'
  import {
    ListAncestries,
    ListClasses,
    ListCommunities,
    SaveCharacter,
    TRAITS,
    TRAIT_ARRAY,
    errorMessage,
    signed
  } from './api.js'

  let { character = null, onsaved, oncancel } = $props()

  const editing = $derived(character !== null)

  const STEPS = ['Identity', 'Class', 'Subclass', 'Ancestry', 'Community', 'Traits', 'Defenses']

  let step = $state(0)
  let saving = $state(false)
  let error = $state('')

  let classes = $state([])
  let ancestries = $state([])
  let communities = $state([])
  let loading = $state(true)

  let form = $state(seed())

  function seed() {
    if (!character) {
      return {
        name: '',
        pronouns: '',
        classSlug: '',
        subclassSlug: '',
        ancestrySlug: '',
        communitySlug: '',
        traits: { agility: 0, strength: 0, finesse: 0, instinct: 0, presence: 0, knowledge: 0 },
        hpMax: 0,
        stressMax: 6,
        evasion: 0,
        armorScore: 0,
        thresholdMajor: 0,
        thresholdSevere: 0,
        background: '',
        connections: ''
      }
    }
    return {
      name: character.name,
      pronouns: character.pronouns,
      classSlug: character.classSlug,
      subclassSlug: character.subclassSlug,
      ancestrySlug: character.ancestrySlug,
      communitySlug: character.communitySlug,
      traits: { ...character.traits },
      hpMax: character.hpMax,
      stressMax: character.stressMax,
      evasion: character.evasion,
      armorScore: character.armorScore,
      thresholdMajor: character.thresholdMajor,
      thresholdSevere: character.thresholdSevere,
      background: character.background,
      connections: character.connections
    }
  }

  const pickedClass = $derived(classes.find((c) => c.slug === form.classSlug) ?? null)
  const pickedSubclass = $derived(pickedClass?.subclasses?.find((s) => s.slug === form.subclassSlug) ?? null)
  const pickedAncestry = $derived(ancestries.find((a) => a.slug === form.ancestrySlug) ?? null)
  const pickedCommunity = $derived(communities.find((c) => c.slug === form.communitySlug) ?? null)

  // The array is spent, not assigned twice — this is what the Traits step checks.
  const spread = $derived(TRAITS.map((t) => form.traits[t.key]).sort((a, b) => b - a))
  const target = $derived([...TRAIT_ARRAY].sort((a, b) => b - a))
  const spreadOk = $derived(spread.join(',') === target.join(','))

  const canAdvance = $derived.by(() => {
    switch (step) {
      case 0:
        return form.name.trim().length > 0
      case 1:
        return form.classSlug !== ''
      case 2:
        return form.subclassSlug !== ''
      case 3:
        return form.ancestrySlug !== ''
      case 4:
        return form.communitySlug !== ''
      default:
        return true
    }
  })

  async function load() {
    try {
      const [k, a, m] = await Promise.all([ListClasses(), ListAncestries(), ListCommunities()])
      classes = k ?? []
      ancestries = a ?? []
      communities = m ?? []
    } catch (e) {
      error = errorMessage(e)
    } finally {
      loading = false
    }
  }

  load()

  function chooseClass(slug) {
    if (form.classSlug === slug) return
    form.classSlug = slug
    // A subclass belongs to exactly one class, so changing class invalidates it.
    form.subclassSlug = ''
    const k = classes.find((c) => c.slug === slug)
    if (k && !editing) {
      form.hpMax = Number(k.startingHitPoints) || 6
      form.evasion = Number(k.startingEvasion) || 10
    }
  }

  function assignTrait(key, value) {
    form.traits[key] = Number(value)
  }

  function autoSpread() {
    TRAITS.forEach((t, i) => {
      form.traits[t.key] = TRAIT_ARRAY[i]
    })
  }

  async function submit() {
    saving = true
    error = ''
    try {
      const saved = await SaveCharacter({
        id: character?.id ?? null,
        name: form.name,
        pronouns: form.pronouns,
        classSlug: form.classSlug,
        subclassSlug: form.subclassSlug,
        ancestrySlug: form.ancestrySlug,
        communitySlug: form.communitySlug,
        traits: form.traits,
        hpMax: Number(form.hpMax),
        stressMax: Number(form.stressMax),
        evasion: Number(form.evasion),
        armorScore: Number(form.armorScore),
        thresholdMajor: Number(form.thresholdMajor),
        thresholdSevere: Number(form.thresholdSevere),
        background: form.background,
        connections: form.connections,
        notes: character?.notes ?? ''
      })
      onsaved?.(saved)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }
</script>

<div class="wizard">
  <header>
    <h2>{editing ? `Edit ${character.name}` : 'New character'}</h2>
    <ol class="steps">
      {#each STEPS as label, i (label)}
        <li class:done={i < step} class:current={i === step}>
          <button type="button" onclick={() => (i <= step ? (step = i) : null)} disabled={i > step}>
            {label}
          </button>
        </li>
      {/each}
    </ol>
  </header>

  {#if error}
    <p class="error">{error}</p>
  {/if}

  {#if loading}
    <p class="empty">Loading reference data…</p>
  {:else}
    <div class="body">
      {#if step === 0}
        <div class="stack">
          <label>
            <span>Name</span>
            <input bind:value={form.name} placeholder="Sabine Vale" required />
          </label>
          <label>
            <span>Pronouns</span>
            <input bind:value={form.pronouns} placeholder="they/them" />
          </label>
          <label>
            <span>Background</span>
            <textarea bind:value={form.background} rows="4" placeholder="Where you come from, and what you left behind."></textarea>
          </label>
          <label>
            <span>Connections</span>
            <textarea bind:value={form.connections} rows="3" placeholder="What ties you to the rest of the party."></textarea>
          </label>
          {#if pickedClass?.backgroundQuestions?.length}
            <div class="prompts">
              <h4>{pickedClass.name} background questions</h4>
              <ul>
                {#each pickedClass.backgroundQuestions as q (q)}<li>{q}</li>{/each}
              </ul>
            </div>
          {/if}
        </div>
      {:else if step === 1}
        <div class="split">
          <ul class="picker">
            {#each classes as k (k.slug)}
              <li>
                <button type="button" class:picked={form.classSlug === k.slug} onclick={() => chooseClass(k.slug)}>
                  <span class="name">{k.name}</span>
                  <span class="meta">{k.domains.join(' · ')}</span>
                </button>
              </li>
            {/each}
          </ul>
          <div class="detail">
            {#if pickedClass}
              <h3>{pickedClass.name}</h3>
              <p class="stats">
                Evasion {pickedClass.startingEvasion} · {pickedClass.startingHitPoints} HP · {pickedClass.domains.join(' & ')}
              </p>
              <SrdText text={pickedClass.description} />
              <h4>Hope feature</h4>
              <FeatureList features={[pickedClass.hopeFeature]} />
              <h4>Class features</h4>
              <FeatureList features={pickedClass.features} />
              {#if pickedClass.classItems?.length}
                <h4>Class items</h4>
                <ul class="bullets">
                  {#each pickedClass.classItems as item (item)}<li>{item}</li>{/each}
                </ul>
              {/if}
            {:else}
              <p class="empty">Pick a class to read what it does.</p>
            {/if}
          </div>
        </div>
      {:else if step === 2}
        <div class="split">
          <ul class="picker">
            {#each pickedClass?.subclasses ?? [] as sub (sub.slug)}
              <li>
                <button type="button" class:picked={form.subclassSlug === sub.slug} onclick={() => (form.subclassSlug = sub.slug)}>
                  <span class="name">{sub.name}</span>
                  <span class="meta">Spellcast: {sub.spellcastTrait || '—'}</span>
                </button>
              </li>
            {/each}
          </ul>
          <div class="detail">
            {#if pickedSubclass}
              <h3>{pickedSubclass.name}</h3>
              <p class="prose">{pickedSubclass.tagline}</p>
              <p class="stats">Spellcast trait: {pickedSubclass.spellcastTrait || '—'}</p>
              <FeatureList features={pickedSubclass.features} />
              <p class="note">You start with the Foundation card. Specialization and Mastery unlock through level-up advancements.</p>
            {:else}
              <p class="empty">Pick a subclass.</p>
            {/if}
          </div>
        </div>
      {:else if step === 3}
        <div class="split">
          <ul class="picker">
            {#each ancestries as a (a.slug)}
              <li>
                <button type="button" class:picked={form.ancestrySlug === a.slug} onclick={() => (form.ancestrySlug = a.slug)}>
                  <span class="name">{a.name}</span>
                </button>
              </li>
            {/each}
          </ul>
          <div class="detail">
            {#if pickedAncestry}
              <h3>{pickedAncestry.name}</h3>
              <SrdText text={pickedAncestry.description} />
              <FeatureList features={pickedAncestry.features} />
            {:else}
              <p class="empty">Pick an ancestry.</p>
            {/if}
          </div>
        </div>
      {:else if step === 4}
        <div class="split">
          <ul class="picker">
            {#each communities as m (m.slug)}
              <li>
                <button type="button" class:picked={form.communitySlug === m.slug} onclick={() => (form.communitySlug = m.slug)}>
                  <span class="name">{m.name}</span>
                </button>
              </li>
            {/each}
          </ul>
          <div class="detail">
            {#if pickedCommunity}
              <h3>{pickedCommunity.name}</h3>
              <SrdText text={pickedCommunity.description} />
              {#if pickedCommunity.adjectives?.length}
                <p class="stats">{pickedCommunity.adjectives.join(', ')}</p>
              {/if}
              <FeatureList features={pickedCommunity.features} />
            {:else}
              <p class="empty">Pick a community.</p>
            {/if}
          </div>
        </div>
      {:else if step === 5}
        <div class="stack">
          <p class="blurb">
            Spend {TRAIT_ARRAY.map(signed).join(', ')} across the six traits — one value each.
          </p>
          <div class="traits">
            {#each TRAITS as trait (trait.key)}
              <label class="trait">
                <span>{trait.label}</span>
                <select value={form.traits[trait.key]} onchange={(e) => assignTrait(trait.key, e.currentTarget.value)}>
                  {#each [-1, 0, 1, 2] as v (v)}
                    <option value={v}>{signed(v)}</option>
                  {/each}
                </select>
                <span class="hint">{trait.blurb}</span>
              </label>
            {/each}
          </div>
          <div class="row">
            <button class="btn ghost" type="button" onclick={autoSpread}>Use the standard spread</button>
            <span class="status" class:ok={spreadOk}>
              {spreadOk ? 'Spread matches the array.' : 'Each of +2, +1, +1, 0, 0, −1 exactly once.'}
            </span>
          </div>
        </div>
      {:else}
        <div class="stack">
          <p class="blurb">
            Hit Points and Evasion come from your class. Armor Score and both thresholds come from the
            armor you're wearing — fill them in from your armor card.
          </p>
          <div class="grid">
            <label><span>Hit Points</span><input type="number" min="1" max="20" bind:value={form.hpMax} /></label>
            <label><span>Stress</span><input type="number" min="1" max="20" bind:value={form.stressMax} /></label>
            <label><span>Evasion</span><input type="number" min="0" max="30" bind:value={form.evasion} /></label>
            <label><span>Armor Score</span><input type="number" min="0" max="20" bind:value={form.armorScore} /></label>
            <label><span>Major threshold</span><input type="number" min="0" max="99" bind:value={form.thresholdMajor} /></label>
            <label><span>Severe threshold</span><input type="number" min="0" max="99" bind:value={form.thresholdSevere} /></label>
          </div>
        </div>
      {/if}
    </div>
  {/if}

  <footer>
    <button class="btn ghost" type="button" onclick={() => oncancel?.()}>Cancel</button>
    <span class="spacer"></span>
    {#if step > 0}
      <button class="btn ghost" type="button" onclick={() => (step -= 1)}>Back</button>
    {/if}
    {#if step < STEPS.length - 1}
      <button class="btn primary" type="button" disabled={!canAdvance} onclick={() => (step += 1)}>Next</button>
    {:else}
      <button class="btn primary" type="button" disabled={saving} onclick={submit}>
        {editing ? 'Save changes' : 'Create character'}
      </button>
    {/if}
  </footer>
</div>

<style>
  .wizard {
    display: flex;
    flex: 1;
    flex-direction: column;
    min-height: 0;
    padding: 1rem 1.25rem;
  }

  header { margin-bottom: 0.85rem; }

  h2 {
    margin: 0 0 0.6rem;
    font-size: 1.25rem;
  }

  h3 {
    margin: 0 0 0.2rem;
    font-size: 1.05rem;
  }

  h4 {
    margin: 1rem 0 0.4rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .steps {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .steps button {
    padding: 0.25rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.72rem;
    cursor: pointer;
  }

  .steps button:disabled { cursor: default; opacity: 0.5; }
  .steps .done button { color: var(--text); border-color: var(--gold-deep); }
  .steps .current button { color: var(--bg); background: var(--gold); border-color: var(--gold); }

  .error {
    margin: 0 0 0.75rem;
    padding: 0.5rem 0.75rem;
    border: 1px solid var(--danger);
    border-radius: 6px;
    font-size: 0.8rem;
    color: var(--danger);
  }

  .body {
    display: flex;
    flex: 1;
    min-height: 0;
  }

  .split {
    display: grid;
    flex: 1;
    grid-template-columns: minmax(11rem, 15rem) 1fr;
    gap: 1rem;
    min-height: 0;
  }

  .picker {
    margin: 0;
    padding: 0;
    list-style: none;
    overflow-y: auto;
  }

  .picker button {
    display: flex;
    flex-direction: column;
    gap: 0.1rem;
    width: 100%;
    margin-bottom: 0.3rem;
    padding: 0.5rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .picker button:hover { border-color: var(--gold-deep); }
  .picker button.picked { border-color: var(--gold); background: var(--panel-2); }

  .name { font-size: 0.9rem; }

  .meta {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .detail {
    max-width: 44rem;
    padding-right: 0.5rem;
    overflow-y: auto;
  }

  .stack {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.75rem;
    max-width: 44rem;
    overflow-y: auto;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(9rem, 1fr));
    gap: 0.6rem;
  }

  .traits {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(10rem, 1fr));
    gap: 0.6rem;
  }

  .trait .hint {
    font-size: 0.68rem;
    text-transform: none;
    letter-spacing: 0;
    opacity: 0.8;
  }

  .row {
    display: flex;
    align-items: center;
    gap: 0.6rem;
  }

  .status {
    font-size: 0.78rem;
    color: var(--danger);
  }

  .status.ok { color: var(--hope); }

  .prose {
    margin: 0.4rem 0;
    font-size: 0.85rem;
    line-height: 1.55;
    color: var(--muted);
    white-space: pre-wrap;
  }

  .stats {
    margin: 0;
    font-size: 0.78rem;
    color: var(--gold);
  }

  .blurb {
    margin: 0;
    font-size: 0.82rem;
    color: var(--muted);
  }

  .note {
    margin: 0.8rem 0 0;
    font-size: 0.75rem;
    font-style: italic;
    color: var(--muted);
  }

  .bullets,
  .prompts ul {
    margin: 0.3rem 0 0;
    padding-left: 1.1rem;
    font-size: 0.82rem;
    line-height: 1.6;
    color: var(--muted);
  }

  .prompts h4 { margin-top: 0.6rem; }

  footer {
    display: flex;
    align-items: center;
    gap: 0.5rem;
    padding-top: 0.85rem;
    margin-top: 0.85rem;
    border-top: 1px solid var(--line);
  }

  .spacer { flex: 1; }
</style>
