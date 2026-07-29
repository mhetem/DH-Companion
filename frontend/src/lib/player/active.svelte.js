// The Player shell splits one character across five nav sections, so which
// character is open has to outlive any single pane. Same trick the GM side uses
// to remember the last campaign, but shared through a rune so switching sections
// doesn't lose the selection.
const KEY = 'hilt.player.activeCharacter'

function stored() {
  const raw = localStorage.getItem(KEY)
  const id = Number(raw)
  return raw && Number.isFinite(id) && id > 0 ? id : null
}

export const active = $state({ id: stored() })

export function setActive(id) {
  active.id = id ?? null
  if (id) {
    localStorage.setItem(KEY, String(id))
  } else {
    localStorage.removeItem(KEY)
  }
}
