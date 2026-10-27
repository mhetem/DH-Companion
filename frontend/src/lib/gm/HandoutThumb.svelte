<script>
  import { renderMarkdown } from './markdown.js'
  import { DEFAULT_VIEW, frameStyle } from '../table/view.js'

  // A small 16:9 stand-in for the player window, laid out the way it will be
  // there: a map letterboxed, a scene cropped to fill, text over its backdrop,
  // all framed by the handout's rotation, zoom and pan. A video sits on its first
  // frame and plays while the pointer is over it.
  //
  // attach is how the On the table preview hands its wheel and drag handling in;
  // the cards in the list leave it out.
  let { kind = '', mediaUrl = '', mediaType = '', body = '', missing = false, view = DEFAULT_VIEW, attach = null } = $props()

  let video = $state(null)
  let width = $state(0)
  let height = $state(0)

  function play() {
    video?.play().catch(() => {})
  }

  function pause() {
    video?.pause()
  }

  const noop = () => {}
</script>

<div
  class="thumb {kind}"
  class:handled={!!attach}
  role="presentation"
  onmouseenter={play}
  onmouseleave={pause}
  bind:clientWidth={width}
  bind:clientHeight={height}
  {@attach attach ?? noop}
>
  {#if missing}
    <span class="note">File missing</span>
  {:else if !kind}
    <span class="note">Black screen</span>
  {:else}
    <div class="frame" style={frameStyle(view ?? DEFAULT_VIEW, width, height)}>
      {#if mediaUrl}
        {#if mediaType === 'video'}
          <video bind:this={video} src="{mediaUrl}#t=0.1" muted loop playsinline preload="metadata"></video>
        {:else}
          <img src={mediaUrl} alt="" loading="lazy" draggable="false" />
        {/if}
      {/if}
      {#if kind === 'text' && body.trim()}
        <div class="words">{@html renderMarkdown(body)}</div>
      {/if}
    </div>
    {#if mediaType === 'video'}<span class="badge">Animated</span>{/if}
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

  .thumb.handled {
    cursor: grab;
    touch-action: none;
    user-select: none;
  }

  .thumb.handled:active { cursor: grabbing; }

  .frame {
    position: absolute;
    display: flex;
    align-items: center;
    justify-content: center;
  }

  img,
  video {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
    -webkit-user-drag: none;
  }

  .map img,
  .map video { object-fit: contain; }

  .text img,
  .text video { filter: brightness(0.3); }

  .words {
    position: relative;
    max-height: 100%;
    padding: 0.4rem 0.6rem;
    overflow: hidden;
    color: var(--text);
    font-size: 0.6rem;
    line-height: 1.35;
  }

  .words :global(h3),
  .words :global(h4),
  .words :global(h5),
  .words :global(h6) {
    margin: 0 0 0.2rem;
    color: var(--gold);
    font-size: 0.75rem;
  }

  .words :global(p),
  .words :global(ul),
  .words :global(ol),
  .words :global(blockquote) { margin: 0 0 0.2rem; }

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
