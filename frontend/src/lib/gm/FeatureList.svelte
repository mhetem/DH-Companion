<script>
  // SRD feature descriptions carry inline markup (<strong>, <em>) and hard
  // newlines for their bullet lists, so they render as HTML inside a pre-wrap
  // block rather than as plain text.
  let { features = [] } = $props()
</script>

{#if features.length}
  <ul class="features">
    {#each features as feature, i (feature.title + i)}
      <li>
        <div class="head">
          <span class="title">{feature.title}</span>
          {#if feature.type}<span class="chip">{feature.type}</span>{/if}
          {#if feature.common}<span class="chip">{feature.common}</span>{/if}
        </div>
        <div class="body">{@html feature.description}</div>
        {#if feature.questions?.length}
          <ul class="questions">
            {#each feature.questions as question (question)}
              <li>{question}</li>
            {/each}
          </ul>
        {/if}
      </li>
    {/each}
  </ul>
{/if}

<style>
  .features {
    display: flex;
    flex-direction: column;
    gap: 0.85rem;
    margin: 0;
    padding: 0;
    list-style: none;
  }

  .head {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-bottom: 0.2rem;
  }

  .title {
    font-weight: 600;
    font-size: 0.9rem;
  }

  .body {
    font-size: 0.85rem;
    line-height: 1.5;
    color: var(--muted);
    white-space: pre-wrap;
  }

  .questions {
    margin: 0.4rem 0 0;
    padding-left: 1rem;
    list-style: none;
  }

  .questions li {
    font-size: 0.8rem;
    font-style: italic;
    line-height: 1.5;
    color: var(--muted);
    opacity: 0.85;
  }

  .questions li::before {
    content: '↳ ';
    opacity: 0.6;
  }
</style>
