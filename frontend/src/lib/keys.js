const TYPING_TAGS = new Set(['INPUT', 'TEXTAREA', 'SELECT'])

// A stepper key must never fire while the GM is filling in a modifier field, so the
// event target decides before anything else does.
export function isTyping(target) {
  if (!target) return false
  return TYPING_TAGS.has(target.tagName) || target.isContentEditable === true
}

// Normalises a keydown into the string the maps below are keyed by: 'h' for the bare
// key, 'H' when shift is held, and the DOM name for everything else ('Escape',
// 'ArrowDown'). Any ctrl/meta/alt chord is left alone so the platform keeps its own.
export function combo(event) {
  if (event.ctrlKey || event.metaKey || event.altKey) return null
  const { key } = event
  if (key.length !== 1) return key
  return event.shiftKey ? key.toUpperCase() : key.toLowerCase()
}

// Every listener is registered on window, so an overlay can't swallow keys by sitting
// on top of the page — nothing is nested. This counter is what makes a modal exclusive:
// while one is up, ordinary handlers stand down, or `h` would quietly mark HP on the
// roster behind the shortcut sheet.
let modalDepth = 0

export const modal = {
  open() {
    modalDepth += 1
  },
  close() {
    modalDepth = Math.max(0, modalDepth - 1)
  },
  get active() {
    return modalDepth > 0
  }
}

// `read` is called per keypress rather than captured once, so a map built out of
// runes always sees current state instead of the values it closed over on mount.
//
// `whenModal` marks the handlers that own an overlay. Those listen in the capture
// phase and stop the event there, which is what actually makes the modal exclusive:
// the counter alone cannot, because a microtask checkpoint runs between two window
// listeners, so Svelte flushes the state change that closes the overlay — and with it
// modal.close() — before the next listener is called. Escape would then both dismiss
// the sheet and drop the roster selection behind it.
export function onKeys(read, { whenModal = false } = {}) {
  function handle(event) {
    if (isTyping(event.target)) return
    if (whenModal && modal.active) event.stopPropagation()
    if (modal.active && !whenModal) return
    const name = combo(event)
    if (!name) return
    const fn = read()[name]
    if (!fn) return
    event.preventDefault()
    fn(event)
  }

  window.addEventListener('keydown', handle, whenModal)
  return () => window.removeEventListener('keydown', handle, whenModal)
}

// The cheat sheet reads this; the components bind the same key strings by hand. Keep
// the two in step — there is no check that they agree.
export const SHORTCUTS = [
  {
    scope: 'Combat runner',
    note: 'While a fight is running.',
    keys: [
      { key: '↑ / ↓', label: 'Select the previous / next combatant' },
      { key: 'k / j', label: 'The same, without leaving the home row' },
      { key: 'h', label: 'Mark 1 HP on the selected combatant' },
      { key: 'H', label: 'Clear 1 HP' },
      { key: 's', label: 'Mark 1 Stress' },
      { key: 'S', label: 'Clear 1 Stress' },
      { key: 'x', label: 'Spotlight the selected combatant' },
      { key: 'c', label: 'Clear every spotlight' },
      { key: 'f', label: 'Gain a Fear' },
      { key: 'F', label: 'Spend a Fear' },
      { key: 'Esc', label: 'Drop the selection' }
    ]
  },
  {
    scope: 'Dice',
    note: 'In the Dice section, and in the runner’s rail when the rollers are shown.',
    keys: [
      { key: 'r', label: 'Roll the d20' },
      { key: 'd', label: 'Roll damage' },
      { key: 'a', label: 'Toggle advantage' },
      { key: 'z', label: 'Toggle disadvantage' },
      { key: '1 … 7', label: 'Pick the damage die, smallest to largest' },
      { key: 'Esc', label: 'Dismiss the result flash' }
    ]
  },
  {
    scope: 'Anywhere',
    keys: [{ key: '?', label: 'Open and close this list' }]
  }
]
