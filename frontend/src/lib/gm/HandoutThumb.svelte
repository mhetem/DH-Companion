<script>
  import { renderMarkdown } from './markdown.js'

  // A small 16:9 stand-in for the player window, laid out the way it will be
  // there: a map letterboxed, a scene cropped to fill, text over its backdrop. A
  // video sits on its first frame and plays while the pointer is over it.
  let { kind = '', mediaUrl = '', mediaType = '', body = '', missing = false } = $props()

  let video = $state(null)

  function play() {
    video?.play().catch(() => {})
  }

  function pause() {
    video?.pause()
  }
</script>

<div class="thumb {kind}" role="presentation" onmouseenter={play} onmouseleave={pause}>
  {#if missing}
    <span class="note">File missing</span>
  {:else if mediaUrl}
    {#if mediaType === 'video'}
      <video bind:this={video} src="{mediaUrl}#t=0.1" muted loop playsinline preload="metadata"></video>
      <span class="badge">Animated</span>
    {:else}
      <img src={mediaUrl} alt="" loading="lazy" />
    {/if}
  {/if}
  {#if kind === 'text' && body.trim()}
    <div class="text">{@html renderMarkdown(body)}</div>
  {:else if !kind}
    <span class="note">Black screen</span>
  {/if}
</div>

<style>
  .thumb {
    position: relative;
    display: flex;
    align-items: center;
    justify-content: center;
    aspect-ratio: 16 / 9;
    overflow: hidden;
    border-radius: 6px;
    background: #000;
  }

  img,
  video {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  .map img,
  .map video { object-fit: contain; }

  .text img,
  .text video { filter: brightness(0.3); }

  .text {
    position: relative;
    max-height: 100%;
    padding: 0.4rem 0.6rem;
    overflow: hidden;
    color: var(--text);
    font-size: 0.6rem;
    line-height: 1.35;
  }

  .text :global(h3),
  .text :global(h4),
  .text :global(h5),
  .text :global(h6) {
    margin: 0 0 0.2rem;
    color: var(--gold);
    font-size: 0.75rem;
  }

  .text :global(p),
  .text :global(ul),
  .text :global(ol),
  .text :global(blockquote) { margin: 0 0 0.2rem; }

  .note {
    position: relative;
    font-size: 0.75rem;
    color: var(--muted);
  }

  .badge {
    position: absolute;
    right: 0.35rem;
    bottom: 0.35rem;
    padding: 0.05rem 0.4rem;
    border-radius: 999px;
    background: rgba(16, 13, 23, 0.75);
    color: var(--text);
    font-size: 0.65rem;
  }
</style>
