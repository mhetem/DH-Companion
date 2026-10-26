<script>
  import { Events, Window } from '@wailsio/runtime'
  import { TableNow } from '../../../bindings/github.com/mhetem/DH-Companion/app.js'
  import { renderMarkdown } from '../gm/markdown.js'

  // The players' screen. It never writes anything: it asks once for whatever the
  // GM is showing, then follows table:show.
  //
  // Each display is a layer that stays invisible until its image or first video
  // frame has loaded, then fades in over the one before. A switch never shows a
  // half-drawn map or the empty window in between, and a file that fails to load
  // fades to black rather than a broken-image icon.
  const FADE_MS = 600
  const IDLE_MS = 2500

  let layers = $state([])
  let seq = 0
  let idle = $state(false)
  let idleTimer = null

  // Subscribe before asking, so a switch made while the answer is on its way
  // can't be overwritten by the older state it carries.
  $effect(() => {
    let heard = false
    const off = Events.On('table:show', (e) => {
      heard = true
      present(e.data)
    })
    TableNow()
      .then((display) => {
        if (!heard) present(display)
      })
      .catch(() => {})
    return off
  })

  function present(display) {
    const layer = { key: ++seq, display, ready: false, failed: false }
    // Only one layer loads at a time: one still loading when the next arrives
    // is dropped, since the players were never going to see it.
    layers = [...layers.filter((l) => l.ready), layer]
    if (!display.mediaUrl) reveal(layer.key)
  }

  function reveal(key, failed = false) {
    const layer = layers.find((l) => l.key === key)
    if (!layer || layer.ready) return
    layer.failed = failed
    layer.ready = true
    setTimeout(() => (layers = layers.filter((l) => l.key >= key)), FADE_MS)
  }

  // A cached image can finish loading before the listener is attached, so the
  // element's own state is checked too.
  function loaded(key) {
    return (node) => {
      const video = node.tagName === 'VIDEO'
      const event = video ? 'loadeddata' : 'load'
      const done = () => reveal(key)
      const fail = () => reveal(key, true)
      node.addEventListener(event, done)
      node.addEventListener('error', fail)
      if (video ? node.readyState >= 2 : node.complete) {
        if (video || node.naturalWidth > 0) done()
        else fail()
      }
      return () => {
        node.removeEventListener(event, done)
        node.removeEventListener('error', fail)
      }
    }
  }

  function wake() {
    idle = false
    clearTimeout(idleTimer)
    idleTimer = setTimeout(() => (idle = true), IDLE_MS)
  }

  function keydown(event) {
    if (event.key === 'F11') {
      event.preventDefault()
      Window.ToggleFullscreen()
    } else if (event.key === 'Escape') {
      Window.UnFullscreen()
    }
  }
</script>

<svelte:head>
  <title>Hilt — Player window</title>
</svelte:head>

<svelte:window onkeydown={keydown} onmousemove={wake} ondblclick={() => Window.ToggleFullscreen()} />

<main class="table" class:idle style:--fade="{FADE_MS}ms">
  {#each layers as layer (layer.key)}
    {@const display = layer.display}
    <div class="layer {display.kind}" class:ready={layer.ready}>
      {#if display.mediaUrl && !layer.failed}
        {#if display.mediaType === 'video'}
          <video src={display.mediaUrl} autoplay muted loop playsinline {@attach loaded(layer.key)}></video>
        {:else}
          <img src={display.mediaUrl} alt="" {@attach loaded(layer.key)} />
        {/if}
      {/if}
      {#if display.kind === 'text' && display.body.trim()}
        <article class:over={display.mediaUrl && !layer.failed}>
          {@html renderMarkdown(display.body)}
        </article>
      {/if}
    </div>
  {/each}
</main>

<style>
  .table {
    position: fixed;
    inset: 0;
    overflow: hidden;
    background: #000;
    user-select: none;
  }

  .table.idle { cursor: none; }

  .layer {
    position: absolute;
    inset: 0;
    display: flex;
    align-items: center;
    justify-content: center;
    opacity: 0;
    transition: opacity var(--fade) ease;
  }

  .layer.ready { opacity: 1; }

  img,
  video {
    position: absolute;
    inset: 0;
    width: 100%;
    height: 100%;
    object-fit: cover;
  }

  /* A map is shown whole, letterboxed if it has to be — cropping one could cut
     off the very edge the party is about to walk off. A scene fills the screen. */
  .map img,
  .map video { object-fit: contain; }

  .text img,
  .text video { filter: brightness(0.3) blur(2px); }

  article {
    position: relative;
    max-width: 38em;
    max-height: 92vh;
    padding: 1.5em 2em;
    overflow: hidden;
    color: var(--text);
    font-size: clamp(1.25rem, 2.4vw, 3rem);
    line-height: 1.5;
  }

  article.over {
    border-radius: 0.4em;
    background: rgba(16, 13, 23, 0.55);
  }

  article :global(h3),
  article :global(h4),
  article :global(h5),
  article :global(h6) {
    margin: 0 0 0.5em;
    color: var(--gold);
    font-size: 1.6em;
    line-height: 1.2;
  }

  article :global(h4) { font-size: 1.3em; }

  article :global(h5),
  article :global(h6) { font-size: 1.1em; }

  article :global(p) { margin: 0 0 0.6em; }

  article :global(ul),
  article :global(ol) { margin: 0 0 0.6em; padding-left: 1.2em; }

  article :global(blockquote) {
    margin: 0 0 0.6em;
    padding-left: 0.7em;
    border-left: 3px solid var(--gold-deep);
    font-style: italic;
  }

  article :global(code) {
    padding: 0 0.25em;
    border-radius: 4px;
    background: rgba(255, 255, 255, 0.08);
  }

  @media (prefers-reduced-motion: reduce) {
    .layer { transition: none; }
  }
</style>
