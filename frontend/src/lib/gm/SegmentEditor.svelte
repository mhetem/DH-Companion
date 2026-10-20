<script>
  import FeatureEditor from './FeatureEditor.svelte'
  import { MAX_SEGMENT_COUNT } from './api.js'

  // segments is the parent's $state array, edited in place the same way
  // FeatureEditor edits features. The attack-ranges and damage-types datalists
  // come from AdversaryForm, which is the only place this is mounted.
  let { segments } = $props()

  function add() {
    segments.push({
      name: '',
      count: 1,
      adjacent: '',
      difficulty: '',
      hp: '',
      standardAttack: { modifier: '', name: '', range: 'Melee', damage: '', damageType: 'phy' },
      features: []
    })
  }

  function remove(index) {
    segments.splice(index, 1)
  }

  function move(index, delta) {
    const next = index + delta
    if (next < 0 || next >= segments.length) return
    const [item] = segments.splice(index, 1)
    segments.splice(next, 0, item)
  }
</script>

<div class="editor">
  {#if !segments.length}
    <p class="hint">A colossus is built from segments — a head, a torso, two arms. Add at least one.</p>
  {/if}

  {#each segments as segment, i (i)}
    <fieldset>
      <div class="bar">
        <span class="index">{segment.name.trim() || `Segment ${i + 1}`}</span>
        <button class="btn ghost" type="button" onclick={() => move(i, -1)} disabled={i === 0} aria-label="Move up">↑</button>
        <button
          class="btn ghost"
          type="button"
          onclick={() => move(i, 1)}
          disabled={i === segments.length - 1}
          aria-label="Move down">↓</button
        >
        <button class="btn danger" type="button" onclick={() => remove(i)}>Remove</button>
      </div>

      <div class="row">
        <label class="grow">
          <span>Name</span>
          <input bind:value={segment.name} placeholder="Arm" />
        </label>
        <label class="narrow">
          <span>How many</span>
          <input type="number" min="1" max={MAX_SEGMENT_COUNT} bind:value={segment.count} />
        </label>
      </div>

      <div class="row">
        <label><span>Difficulty</span><input bind:value={segment.difficulty} placeholder="13" /></label>
        <label>
          <span>HP {segment.count > 1 ? '(each)' : ''}</span>
          <input bind:value={segment.hp} placeholder="3" />
        </label>
        <label class="grow">
          <span>Adjacent segments</span>
          <input bind:value={segment.adjacent} placeholder="Torso, the other leg" />
        </label>
      </div>
      <small>Leave HP blank for a segment that can't be damaged at all — it spawns with no track.</small>

      <div class="row">
        <label class="grow"><span>Attack</span><input bind:value={segment.standardAttack.name} placeholder="Punch" /></label>
        <label class="narrow"><span>Modifier</span><input bind:value={segment.standardAttack.modifier} placeholder="+2" /></label>
      </div>
      <div class="row">
        <label>
          <span>Range</span>
          <input bind:value={segment.standardAttack.range} list="attack-ranges" placeholder="Very Close" />
        </label>
        <label><span>Damage</span><input bind:value={segment.standardAttack.damage} placeholder="1d8+6" /></label>
        <label>
          <span>Damage type</span>
          <input bind:value={segment.standardAttack.damageType} list="damage-types" placeholder="phy" />
        </label>
      </div>
      <small>Leave the attack name and damage blank for a segment with no standard attack.</small>

      <div class="features">
        <span class="label">Segment features</span>
        <FeatureEditor features={segment.features} />
      </div>
    </fieldset>
  {/each}

  <button class="btn" type="button" onclick={add}>+ Add segment</button>
</div>

<style>
  .editor {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    align-items: flex-start;
  }

  .hint {
    margin: 0;
    font-size: 0.8rem;
    color: var(--muted);
  }

  fieldset {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
    margin: 0;
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-left: 0.2rem solid var(--fear);
    border-radius: 8px;
    background: var(--panel);
    box-sizing: border-box;
  }

  .bar {
    display: flex;
    align-items: center;
    gap: 0.3rem;
  }

  .index {
    flex: 1;
    font-size: 0.8rem;
    font-weight: 600;
  }

  .row {
    display: flex;
    gap: 0.5rem;
  }

  label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 0;
  }

  label.grow { flex: 2; }
  label.narrow { flex: 0 0 7rem; }

  label span,
  .label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  small {
    margin-top: -0.3rem;
    font-size: 0.7rem;
    color: var(--muted);
    opacity: 0.75;
  }

  .features {
    display: flex;
    flex-direction: column;
    gap: 0.4rem;
  }
</style>
