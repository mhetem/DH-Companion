<script>
  import {
    CurrentWindowSize,
    SetUIScale,
    SetWindowSize,
    UIScales,
    WindowPresets
  } from '../../wailsjs/go/main/App.js'
  import { applyScale } from './display.js'
  import ShortcutHelp from './ShortcutHelp.svelte'
  import lockup from '../assets/images/hilt-app-lockup.svg'

  let { role, scale, onswitch, onscale, onerror } = $props()

  let scales = $state([])

  UIScales()
    .then((list) => (scales = list ?? []))
    .catch((e) => onerror?.(String(e)))

  // Applied straight away so the change is instant, then persisted — a failed write
  // puts the old scale back rather than leaving the screen disagreeing with the DB.
  async function rescale(raw) {
    const next = Number(raw)
    const previous = scale
    applyScale(next)
    onscale?.(next)
    try {
      await SetUIScale(next)
    } catch (e) {
      applyScale(previous)
      onscale?.(previous)
      onerror?.(typeof e === 'string' ? e : String(e))
    }
  }

  const other = $derived(role === 'gm' ? 'player' : 'gm')
  const otherLabel = $derived(other === 'gm' ? 'Game Master' : 'Player')

  let presets = $state([])
  let current = $state(null)

  // A size the presets don't cover — dragged by hand, or clamped down to fit the
  // screen — still needs an entry, or the select would show the wrong one.
  const options = $derived.by(() => {
    if (!current) return presets
    if (presets.some((p) => p.width === current.width && p.height === current.height)) return presets
    return [...presets, { ...current, label: 'Custom', fits: true }]
  })

  const value = $derived(current ? `${current.width}x${current.height}` : '')

  Promise.all([WindowPresets(), CurrentWindowSize()])
    .then(([list, size]) => {
      presets = list ?? []
      current = size
    })
    .catch((e) => onerror?.(String(e)))

  async function resize(raw) {
    const [width, height] = raw.split('x').map(Number)
    try {
      current = await SetWindowSize(width, height)
    } catch (e) {
      onerror?.(typeof e === 'string' ? e : String(e))
      current = await CurrentWindowSize()
    }
  }
</script>

<header>
  <img class="brand" src={lockup} alt="Hilt" />
  <span class="badge" class:gm={role === 'gm'} class:player={role === 'player'}>
    {role === 'gm' ? 'Game Master' : 'Player'}
  </span>

  <div class="display">
    <ShortcutHelp />

    <label>
      <span class="sronly">Window size</span>
      <select {value} onchange={(e) => resize(e.currentTarget.value)} title="Window size">
        {#each options as preset (preset.label + preset.width)}
          <option value="{preset.width}x{preset.height}" disabled={!preset.fits}>
            {preset.label} — {preset.width}×{preset.height}{preset.fits ? '' : ' (too big for this screen)'}
          </option>
        {/each}
      </select>
    </label>

    <label>
      <span class="sronly">UI scale</span>
      <select value={String(scale)} onchange={(e) => rescale(e.currentTarget.value)} title="UI scale">
        {#each scales as option (option)}
          <option value={String(option)}>{option}%{option === 100 ? ' (default)' : ''}</option>
        {/each}
      </select>
    </label>
  </div>

  <button class="switch" onclick={() => onswitch(other)}>
    Switch to {otherLabel}
  </button>
</header>

<style>
  header {
    position: relative;
    display: flex;
    align-items: center;
    gap: 0.75rem;
    padding: 0.6rem 1rem;
    border-bottom: 1px solid var(--line);
    background: var(--panel);
  }

  /* A gold hairline over the border, fading out past the lockup — the brand frame
     without a second hard rule across the window. */
  header::after {
    content: '';
    position: absolute;
    inset: auto 0 -1px 0;
    height: 2px;
    background: linear-gradient(to right, var(--gold), var(--gold-deep) 20%, transparent 55%);
  }

  /* The lockup carries ~16% dead space above and below its artwork, so it is sized
     past the header's line box and pulled back in rather than padding it out. */
  .brand {
    display: block;
    height: 2.9rem;
    margin: -0.55rem 0;
  }

  .badge {
    padding: 0.15rem 0.6rem;
    border-radius: 999px;
    border: 1px solid var(--line);
    font-size: 0.75rem;
    text-transform: uppercase;
    letter-spacing: 0.06em;
    color: var(--muted);
  }

  .badge.gm { border-color: var(--fear); color: var(--fear); }
  .badge.player { border-color: var(--hope); color: var(--hope); }

  .display {
    display: flex;
    align-items: center;
    gap: 0.4rem;
    margin-left: auto;
  }

  .display select { font-size: 0.8rem; }

  .sronly {
    position: absolute;
    width: 1px;
    height: 1px;
    overflow: hidden;
    clip-path: inset(50%);
    white-space: nowrap;
  }

  .switch {
    padding: 0.35rem 0.8rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: transparent;
    color: var(--muted);
    font: inherit;
    font-size: 0.85rem;
    cursor: pointer;
  }

  .switch:hover {
    color: var(--text);
    border-color: var(--muted);
  }
</style>
