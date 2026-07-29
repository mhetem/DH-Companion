<script>
  // SRD prose carries inline <strong>/<em> and hard newlines, exactly like the feature
  // descriptions FeatureList renders — so it needs the same pre-wrap HTML treatment or
  // the tags show up as literal text on the page.
  //
  // Point this only at catalog text. Anything the user typed (homebrew descriptions,
  // a character's background, note bodies) must keep going through normal
  // interpolation, which escapes it.
  let { text = '', tone = 'muted' } = $props()
</script>

{#if text}
  <div class="srd {tone}">{@html text}</div>
{/if}

<style>
  .srd {
    font-size: 0.85rem;
    line-height: 1.55;
    white-space: pre-wrap;
  }

  .muted { color: var(--muted); }
  .plain { color: var(--text); }
</style>
