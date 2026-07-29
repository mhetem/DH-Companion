<script>
  import Modal from '../Modal.svelte'
  import { player } from '../../../wailsjs/go/models'
  import { ApplyLevelUp, ListClasses, PlanLevelUp, errorMessage, signed, traitLabel } from './api.js'

  let { open = $bindable(false), characterId, onapplied, oncancel } = $props()

  let plan = $state(null)
  let loading = $state(false)
  let saving = $state(false)
  let error = $state('')

  // One entry per advancement selection; an option costing 2 still produces one
  // entry, and `spent` is what the footer checks against.
  let picks = $state([])
  let newExperience = $state('')
  let levelCard = $state('')

  const spent = $derived(
    picks.reduce((total, pick) => total + (plan?.options.find((o) => o.id === pick.id)?.cost || 1), 0)
  )
  const required = $derived(plan?.advancementsRequired ?? 2)

  // Traits marked earlier in this same level-up can't be picked again by a second
  // trait advancement, so the pool shrinks as choices are made.
  const markedNow = $derived([...(plan?.markedTraits ?? []), ...picks.flatMap((p) => p.traits ?? [])])

  const cardsChosen = $derived([levelCard, ...picks.map((p) => p.domainCardSlug)].filter(Boolean))

  // What's still missing, in words. A disabled button with no reason attached is the
  // thing that made this screen feel broken — every gate below now says so out loud,
  // and `complete` is just "nothing left to say".
  function missing(pick) {
    const option = plan?.options.find((o) => o.id === pick.id)
    if (!option) return 'That advancement is no longer available'
    const e = option.effect ?? {}
    if (e.traits > 0 && pick.traits.length !== e.traits) {
      const short = e.traits - pick.traits.length
      return short > 0
        ? `Choose ${short} more ${short === 1 ? 'trait' : 'traits'}`
        : `Choose only ${e.traits} traits`
    }
    if (e.experiences > 0 && pick.experienceIds.length !== e.experiences) {
      const short = e.experiences - pick.experienceIds.length
      return short > 0
        ? `Choose ${short} more ${short === 1 ? 'Experience' : 'Experiences'}`
        : `Choose only ${e.experiences} Experiences`
    }
    if (e.domainCards > 0 && !pick.domainCardSlug) return 'Choose the extra domain card'
    if (e.multiclass && !pick.multiclassSlug) return 'Choose a class to multiclass into'
    if (e.multiclass && !pick.multiclassSubclassSlug) return 'Choose a foundation subclass'
    return ''
  }

  const blockers = $derived.by(() => {
    if (!plan) return ['Loading…']
    const out = []
    if (spent < required) out.push(`Pick ${required - spent} more advancement${required - spent === 1 ? '' : 's'}`)
    if (spent > required) out.push(`Pick ${required} advancements — you have ${spent}`)
    if (plan.needsNewExperience && !newExperience.trim()) out.push('Name your new Experience')
    if (plan.domainCardsPerLevel > 0 && !levelCard) out.push(`Choose your level ${plan.toLevel} domain card`)
    for (const pick of picks) {
      const why = missing(pick)
      if (why) out.push(why)
    }
    return out
  })

  const complete = $derived(blockers.length === 0)

  let classes = $state([])

  $effect(() => {
    if (!open || !characterId) return
    loading = true
    error = ''
    picks = []
    newExperience = ''
    levelCard = ''
    Promise.all([PlanLevelUp(characterId), ListClasses()])
      .then(([p, k]) => {
        plan = p
        classes = k ?? []
      })
      .catch((e) => {
        error = errorMessage(e)
        plan = null
      })
      .finally(() => (loading = false))
  })

  function taken(id) {
    return picks.filter((p) => p.id === id).length
  }

  function toggle(option) {
    const index = picks.findIndex((p) => p.id === option.id)
    if (index !== -1 && option.effect.traits === 0 && option.effect.experiences === 0) {
      picks = picks.filter((_, i) => i !== index)
      return
    }
    if (taken(option.id) >= option.remaining) return
    if (spent + (option.cost || 1) > required) return
    picks = [
      ...picks,
      {
        id: option.id,
        traits: [],
        experienceIds: [],
        domainCardSlug: '',
        multiclassSlug: '',
        multiclassSubclassSlug: ''
      }
    ]
  }

  function drop(index) {
    picks = picks.filter((_, i) => i !== index)
  }

  function toggleTrait(index, key, limit) {
    const pick = picks[index]
    if (pick.traits.includes(key)) {
      pick.traits = pick.traits.filter((t) => t !== key)
      return
    }
    if (pick.traits.length >= limit) return
    pick.traits = [...pick.traits, key]
  }

  function toggleExperience(index, id, limit) {
    const pick = picks[index]
    if (pick.experienceIds.includes(id)) {
      pick.experienceIds = pick.experienceIds.filter((e) => e !== id)
      return
    }
    if (pick.experienceIds.length >= limit) return
    pick.experienceIds = [...pick.experienceIds, id]
  }

  function optionFor(id) {
    return plan?.options.find((o) => o.id === id)
  }

  function subclassesFor(slug) {
    return classes.find((c) => c.slug === slug)?.subclasses ?? []
  }

  async function submit() {
    saving = true
    error = ''
    try {
      // Nested inputs go through the generated class, the same as the GM builder's
      // EncounterInput — a bare object literal isn't assignable to it.
      const sheet = await ApplyLevelUp(
        player.LevelUpInput.createFrom({
          characterId,
          choices: picks,
          newExperienceName: newExperience,
          domainCardSlug: levelCard
        })
      )
      open = false
      onapplied?.(sheet)
    } catch (e) {
      error = errorMessage(e)
    } finally {
      saving = false
    }
  }
</script>

<Modal bind:open label="Level up" width="46rem">
  {#if loading}
    <p class="empty">Working out what you can take…</p>
  {:else if !plan}
    <p class="error">{error || 'No advancement plan available.'}</p>
    <div class="foot"><button class="btn ghost" onclick={() => (open = false)}>Close</button></div>
  {:else}
    <header>
      <h2>Level {plan.fromLevel} → {plan.toLevel}</h2>
      <p class="sub">
        {plan.tierName}{#if plan.newTier} · new tier{/if} · pick {required} advancements
      </p>
    </header>

    {#if error}<p class="error">{error}</p>{/if}

    {#if plan.newTier && plan.achievements.length}
      <section>
        <h3>Tier achievements</h3>
        <ul class="bullets">
          {#each plan.achievements as line (line)}<li>{line}</li>{/each}
        </ul>
        {#if plan.needsNewExperience}
          <label>
            <span>New Experience (+2)</span>
            <input bind:value={newExperience} placeholder="Survived the Tangle" />
          </label>
        {/if}
      </section>
    {/if}

    <section>
      <h3>Advancements <span class="count" class:ok={spent === required}>{spent}/{required}</span></h3>
      <ul class="options">
        {#each plan.options as option (option.id)}
          {@const used = taken(option.id)}
          <li class:picked={used > 0} class:off={!option.available}>
            <button
              type="button"
              disabled={!option.available || used >= option.remaining || spent + (option.cost || 1) > required}
              onclick={() => toggle(option)}
            >
              <span class="label">{option.label}</span>
              <span class="slots">
                {#each Array(option.slots) as _, i (i)}
                  <span class="slot" class:filled={i < option.used + used}></span>
                {/each}
                {#if option.cost > 1}<span class="cost">costs {option.cost}</span>{/if}
              </span>
              {#if option.description}<span class="desc">{option.description}</span>{/if}
              {#if !option.available}<span class="why">{option.reason}</span>{/if}
            </button>
          </li>
        {/each}
      </ul>
    </section>

    {#if picks.length}
      <section>
        <h3>Your picks</h3>
        <ul class="picks">
          {#each picks as pick, index (index)}
            {@const option = optionFor(pick.id)}
            <li>
              <div class="pick-head">
                <span class="label">{option?.label}</span>
                <button class="btn ghost" onclick={() => drop(index)}>Remove</button>
              </div>
              {#if missing(pick)}
                <p class="todo">{missing(pick)}</p>
              {/if}

              {#if option?.effect.traits > 0}
                <p class="need">
                  Choose {option.effect.traits} unmarked traits —
                  <strong>{pick.traits.length}/{option.effect.traits}</strong> chosen.
                </p>
                <div class="chips">
                  {#each plan.traits ? Object.keys(plan.traits) : [] as key (key)}
                    {@const blocked = markedNow.includes(key) && !pick.traits.includes(key)}
                    <button
                      class="chip"
                      class:on={pick.traits.includes(key)}
                      disabled={blocked}
                      onclick={() => toggleTrait(index, key, option.effect.traits)}
                    >
                      {traitLabel(key)} {signed(plan.traits[key])}
                    </button>
                  {/each}
                </div>
              {/if}

              {#if option?.effect.experiences > 0}
                <p class="need">
                  Choose {option.effect.experiences} Experiences to raise by
                  +{option.effect.experienceBonus || 1} —
                  <strong>{pick.experienceIds.length}/{option.effect.experiences}</strong> chosen.
                </p>
                <div class="chips">
                  {#each plan.experiences as experience (experience.id)}
                    <button
                      class="chip"
                      class:on={pick.experienceIds.includes(experience.id)}
                      onclick={() => toggleExperience(index, experience.id, option.effect.experiences)}
                    >
                      {experience.name} {signed(experience.modifier)}
                    </button>
                  {/each}
                  {#if !plan.experiences.length}
                    <span class="empty">No Experiences to raise yet.</span>
                  {/if}
                </div>
              {/if}

              {#if option?.effect.domainCards > 0}
                <label>
                  <span>Extra domain card</span>
                  <select bind:value={pick.domainCardSlug}>
                    <option value="">Choose a card…</option>
                    {#each plan.availableCards as card (card.slug)}
                      <option value={card.slug} disabled={card.slug !== pick.domainCardSlug && cardsChosen.includes(card.slug)}>
                        {card.name} — {card.domain} {card.level}
                      </option>
                    {/each}
                  </select>
                </label>
              {/if}

              {#if option?.effect.multiclass}
                <div class="pair">
                  <label>
                    <span>Class</span>
                    <select bind:value={pick.multiclassSlug} onchange={() => (pick.multiclassSubclassSlug = '')}>
                      <option value="">Choose a class…</option>
                      {#each classes as k (k.slug)}
                        <option value={k.slug}>{k.name}</option>
                      {/each}
                    </select>
                  </label>
                  <label>
                    <span>Foundation subclass</span>
                    <select bind:value={pick.multiclassSubclassSlug} disabled={!pick.multiclassSlug}>
                      <option value="">Choose a subclass…</option>
                      {#each subclassesFor(pick.multiclassSlug) as sub (sub.slug)}
                        <option value={sub.slug}>{sub.name}</option>
                      {/each}
                    </select>
                  </label>
                </div>
              {/if}
            </li>
          {/each}
        </ul>
      </section>
    {/if}

    <section>
      <h3>Level {plan.toLevel} domain card</h3>
      <select bind:value={levelCard}>
        <option value="">Choose a card…</option>
        {#each plan.availableCards as card (card.slug)}
          <option value={card.slug} disabled={card.slug !== levelCard && picks.some((p) => p.domainCardSlug === card.slug)}>
            {card.name} — {card.domain} {card.level} ({card.type})
          </option>
        {/each}
      </select>
      <p class="hint">
        New cards go straight into your loadout while there's room, and to the vault after that.
        Both damage thresholds go up by {plan.thresholdIncrease}.
      </p>
    </section>

    {#if blockers.length}
      <ul class="blockers">
        {#each blockers as blocker, i (i)}<li>{blocker}</li>{/each}
      </ul>
    {/if}

    <div class="foot">
      <button class="btn ghost" onclick={() => { open = false; oncancel?.() }}>Cancel</button>
      <span class="spacer"></span>
      <button class="btn primary" disabled={!complete || saving} onclick={submit}>
        Level up to {plan.toLevel}
      </button>
    </div>
  {/if}
</Modal>

<style>
  header { margin-bottom: -0.4rem; }

  h2 {
    margin: 0;
    font-size: 1.2rem;
  }

  h3 {
    display: flex;
    align-items: baseline;
    gap: 0.5rem;
    margin: 0 0 0.5rem;
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .sub {
    margin: 0.2rem 0 0;
    font-size: 0.8rem;
    color: var(--muted);
  }

  .count { color: var(--danger); }
  .count.ok { color: var(--hope); }

  section {
    display: flex;
    flex-direction: column;
    gap: 0.5rem;
  }

  .options,
  .picks {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .options button {
    display: flex;
    flex-direction: column;
    gap: 0.2rem;
    width: 100%;
    padding: 0.5rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--panel-2);
    color: var(--text);
    font: inherit;
    text-align: left;
    cursor: pointer;
  }

  .options button:hover:not(:disabled) { border-color: var(--gold-deep); }
  .options button:disabled { cursor: default; opacity: 0.55; }
  .options li.picked button { border-color: var(--hope); }

  .label { font-size: 0.85rem; }

  .desc,
  .why {
    font-size: 0.72rem;
    color: var(--muted);
  }

  .why { color: var(--danger); }

  .slots {
    display: flex;
    align-items: center;
    gap: 0.25rem;
  }

  .slot {
    width: 0.6rem;
    height: 0.6rem;
    border: 1px solid var(--line);
    border-radius: 2px;
  }

  .slot.filled {
    border-color: var(--gold);
    background: var(--gold);
  }

  .cost {
    margin-left: 0.3rem;
    font-size: 0.65rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--gold);
  }

  .picks li {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
    padding: 0.6rem;
    border: 1px solid var(--hope);
    border-radius: 8px;
  }

  .pick-head {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.5rem;
  }

  .need {
    margin: 0;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .need strong {
    font-weight: 600;
    color: var(--hope);
  }

  .todo {
    margin: 0;
    font-size: 0.75rem;
    color: var(--gold);
  }

  .blockers {
    display: flex;
    flex-direction: column;
    gap: 0.15rem;
    margin: 0;
    padding: 0.5rem 0.75rem 0.5rem 1.6rem;
    border: 1px solid var(--gold-deep);
    border-radius: 8px;
    font-size: 0.78rem;
    color: var(--muted);
  }

  .chips {
    display: flex;
    flex-wrap: wrap;
    gap: 0.3rem;
  }

  .chip {
    padding: 0.25rem 0.5rem;
    border: 1px solid var(--line);
    border-radius: 999px;
    background: var(--panel-2);
    color: var(--muted);
    font: inherit;
    font-size: 0.75rem;
    cursor: pointer;
  }

  .chip.on {
    border-color: var(--hope);
    color: var(--text);
  }

  .chip:disabled {
    cursor: default;
    opacity: 0.4;
  }

  .pair {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 0.5rem;
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

  .bullets {
    margin: 0;
    padding-left: 1.1rem;
    font-size: 0.82rem;
    line-height: 1.6;
    color: var(--muted);
  }

  .hint {
    margin: 0;
    font-size: 0.72rem;
    line-height: 1.5;
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

  .foot {
    display: flex;
    align-items: center;
    gap: 0.5rem;
  }

  .spacer { flex: 1; }
</style>
