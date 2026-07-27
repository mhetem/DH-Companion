<script>
  import { COMMON_FEATURES, FEATURE_TYPES } from './api.js'

  // features is the parent's $state array — Svelte 5 state is deeply reactive,
  // so pushing and splicing here updates the form in place.
  let { features, withQuestions = false, withCommon = false } = $props()

  function add() {
    features.push({ common: '', title: '', type: 'Action', description: '', questions: [] })
  }

  function remove(index) {
    features.splice(index, 1)
  }

  function move(index, delta) {
    const next = index + delta
    if (next < 0 || next >= features.length) return
    const [item] = features.splice(index, 1)
    features.splice(next, 0, item)
  }

  function addQuestion(feature) {
    feature.questions = [...(feature.questions ?? []), '']
  }

  function removeQuestion(feature, index) {
    feature.questions.splice(index, 1)
  }
</script>

<div class="editor">
  {#each features as feature, i (i)}
    <fieldset>
      <div class="bar">
        <span class="index">Feature {i + 1}</span>
        <button class="btn ghost" type="button" onclick={() => move(i, -1)} disabled={i === 0} aria-label="Move up">↑</button>
        <button
          class="btn ghost"
          type="button"
          onclick={() => move(i, 1)}
          disabled={i === features.length - 1}
          aria-label="Move down">↓</button
        >
        <button class="btn danger" type="button" onclick={() => remove(i)}>Remove</button>
      </div>

      <div class="row">
        <label class="grow">
          <span>Title</span>
          <input bind:value={feature.title} placeholder="Earth Eruption" />
        </label>
        <label class="narrow">
          <span>Type</span>
          <select bind:value={feature.type}>
            {#each FEATURE_TYPES as t (t)}
              <option value={t}>{t}</option>
            {/each}
          </select>
        </label>
        {#if withCommon}
          <label class="narrow">
            <span>Keyword</span>
            <input bind:value={feature.common} list="common-features" placeholder="none" />
          </label>
        {/if}
      </div>

      <label>
        <span>Description</span>
        <textarea rows="4" bind:value={feature.description}></textarea>
        <small>Inline &lt;strong&gt; and &lt;em&gt; render as markup, matching the SRD cards.</small>
      </label>

      {#if withQuestions}
        <div class="questions">
          <span class="label">GM questions</span>
          {#each feature.questions ?? [] as _, q (q)}
            <div class="question">
              <input bind:value={feature.questions[q]} placeholder="Why is the grove unused now?" />
              <button class="btn danger" type="button" onclick={() => removeQuestion(feature, q)} aria-label="Remove question">×</button>
            </div>
          {/each}
          <button class="btn ghost" type="button" onclick={() => addQuestion(feature)}>+ Question</button>
        </div>
      {/if}
    </fieldset>
  {/each}

  <button class="btn" type="button" onclick={add}>+ Add feature</button>
</div>

{#if withCommon}
  <datalist id="common-features">
    {#each COMMON_FEATURES as name (name)}
      <option value={name}></option>
    {/each}
  </datalist>
{/if}

<style>
  .editor {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    align-items: flex-start;
  }

  fieldset {
    display: flex;
    flex-direction: column;
    gap: 0.6rem;
    width: 100%;
    margin: 0;
    padding: 0.75rem;
    border: 1px solid var(--line);
    border-radius: 8px;
    background: var(--bg);
  }

  .bar {
    display: flex;
    align-items: center;
    gap: 0.3rem;
  }

  .index {
    flex: 1;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .row {
    display: flex;
    gap: 0.5rem;
  }

  label {
    display: flex;
    flex-direction: column;
    gap: 0.25rem;
  }

  label.grow { flex: 1; }
  label.narrow { flex: 0 0 8rem; }

  label span,
  .label {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  small {
    font-size: 0.7rem;
    color: var(--muted);
    opacity: 0.75;
  }

  .questions {
    display: flex;
    flex-direction: column;
    gap: 0.35rem;
    align-items: flex-start;
  }

  .question {
    display: flex;
    gap: 0.3rem;
    width: 100%;
  }

  .question input { flex: 1; }
</style>
