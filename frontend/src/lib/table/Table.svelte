<script>
  import { Events, Window } from '@wailsio/runtime'
  import { SetTableView, TableNow } from '../../../bindings/github.com/mhetem/DH-Companion/app.js'
  import { renderMarkdown } from '../gm/markdown.js'
  import { DEFAULT_VIEW, clampView, frameStyle, gestures, rotateBy, throttle, zoomIn, zoomOut } from './view.js'

  // The players' screen. It asks once for whatever the GM is showing, then follows
  // table:show. The only thing it ever sends back is the framing — rotation, zoom
  // and pan — when the GM handles it here directly; Go relays that to the GM's
  // preview and keeps it with the handout.
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
  let width = $state(0)
  let height = $state(0)

  // The top layer is the one on show, and the one the wheel and drag move.
  const top = $derived(layers.at(-1) ?? null)

  // Subscribe before asking, so a switch made while the answer is on its way
  // can't be overwritten by the older state it carries.
  $effect(() => {
    let heard = false
    const offShow = Events.On('table:show', (e) => {
      heard = true
      present(e.data)
    })
    const offView = Events.On('table:view', (e) => {
      const { handoutId, view, origin } = e.data
      if (origin !== 'table' && top?.display.handoutId === handoutId) top.view = clampView(view)
    })
    TableNow()
      .then((display) => {
        if (!heard) present(display)
      })
      .catch(() => {})
    return () => {
      offShow()
      offView()
    }
  })

  function present(display) {
    const layer = { key: ++seq, display, view: clampView(display.view ?? DEFAULT_VIEW), ready: false, failed: false }
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

  const send = throttle((handoutId, view) => SetTableView(handoutId, view, 'table').catch(() => {}), 40)

  // The framing moves here at once and reaches Go a moment later. Only a display
  // with something on it can be framed; a black screen has nothing to move.
  function setView(next) {
    if (!top?.display.kind) return
    top.view = clampView(next)
    send(top.display.handoutId, $state.snapshot(top.view))
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
    } else if (top?.display.kind) {
      if (event.key === 'r') setView(rotateBy(top.view, 90))
      else if (event.key === 'R') setView(rotateBy(top.view, -90))
      else if (event.key === '+' || event.key === '=') setView(zoomIn(top.view))
      else if (event.key === '-') setView(zoomOut(top.view))
      else if (event.key === '0') setView(DEFAULT_VIEW)
    }
  }
</script>

<svelte:head>
  <title>Hilt — Player window</title>
</svelte:head>

<svelte:window onkeydown={keydown} onmousemove={wake} ondblclick={() => Window.ToggleFullscreen()} />

<main
  class="table"
  class:idle
  style:--fade="{FADE_MS}ms"
  bind:clientWidth={width}
  bind:clientHeight={height}
  {@attach gestures(() => (top?.display.kind ? top.view : null), setView)}
>
  {#each layers as layer (layer.key)}
    {@const display = layer.display}
    <div class="layer {display.kind}" class:ready={layer.ready}>
      <div class="frame" style={frameStyle(layer.view, width, height)}>
        {#if display.mediaUrl && !layer.failed}
          {#if display.mediaType === 'video'}
            <video src={display.mediaUrl} autoplay muted loop playsinline {@attach loaded(layer.key)}></video>
          {:else}
            <img src={display.mediaUrl} alt="" draggable="false" {@attach loaded(layer.key)} />
          {/if}
        {/if}
        {#if display.kind === 'text' && display.body.trim()}
          <article class:over={display.mediaUrl && !layer.failed}>
            {@html renderMarkdown(display.body)}
          </article>
        {/if}
      </div>
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
    touch-action: none;
  }

  .table.idle { cursor: none; }

  .layer {
    position: absolute;
    inset: 0;
    opacity: 0;
    transition: opacity var(--fade) ease;
  }

  .layer.ready { opacity: 1; }

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

  /* A map is shown whole, letterboxed if it has to be — cropping one could cut
     off the very edge the party is about to walk off. A scene fills the screen. */
  .map img,
  .map video { object-fit: contain; }

  .text img,
  .text video { filter: brightness(0.3) blur(2px); }

  article {
    position: relative;
    max-width: 38em;
    max-height: 92%;
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
