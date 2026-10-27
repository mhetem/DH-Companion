// How a map, scene or text is framed on the player window: a quarter-turn
// rotation, a zoom, and a pan measured in fractions of the screen, so the GM's
// small preview and the full-size window show the same framing. Mirrors
// NormalizeView in internal/gm/handouts.go — keep the bounds in step.

export const DEFAULT_VIEW = { rotate: 0, zoom: 1, x: 0, y: 0 }

const MIN_ZOOM = 0.25
const MAX_ZOOM = 8
const MAX_PAN = 3
const ZOOM_STEP = 1.25

const clamp = (v, lo, hi) => Math.min(Math.max(v, lo), hi)

export function clampView(v) {
  return {
    rotate: (((Math.round((v?.rotate ?? 0) / 90) * 90) % 360) + 360) % 360,
    zoom: clamp(v?.zoom || 1, MIN_ZOOM, MAX_ZOOM),
    x: clamp(v?.x ?? 0, -MAX_PAN, MAX_PAN),
    y: clamp(v?.y ?? 0, -MAX_PAN, MAX_PAN)
  }
}

export function rotateBy(v, degrees) {
  return clampView({ ...v, rotate: v.rotate + degrees })
}

// Zooming keeps the point under the cursor where it is: px and py are that
// point's offset from the centre, as fractions of the box. Without a point it
// zooms about the centre.
export function zoomBy(v, factor, px = 0, py = 0) {
  const zoom = clamp(v.zoom * factor, MIN_ZOOM, MAX_ZOOM)
  const k = zoom / v.zoom
  return clampView({ ...v, zoom, x: px - k * (px - v.x), y: py - k * (py - v.y) })
}

export const zoomIn = (v) => zoomBy(v, ZOOM_STEP)
export const zoomOut = (v) => zoomBy(v, 1 / ZOOM_STEP)

// The content box is laid out at the screen's size — swapped on a quarter turn,
// so a rotated map is fitted to the screen it ends up on rather than to the one
// it started on — then rotated, scaled and panned about the screen's centre.
export function frameStyle(v, width, height) {
  const sideways = v.rotate % 180 !== 0
  const w = sideways ? height : width
  const h = sideways ? width : height
  return [
    `left: ${(width - w) / 2}px`,
    `top: ${(height - h) / 2}px`,
    `width: ${w}px`,
    `height: ${h}px`,
    `transform: translate(${v.x * width}px, ${v.y * height}px) scale(${v.zoom}) rotate(${v.rotate}deg)`
  ].join('; ')
}

// Wheel to zoom at the cursor, drag to pan. get() reads the view the element is
// showing now and set() is handed the next one; both windows use this, so a
// drag means the same thing on the preview as on the table.
export function gestures(get, set) {
  return (node) => {
    let last = null

    function wheel(event) {
      if (!get()) return
      event.preventDefault()
      const r = node.getBoundingClientRect()
      const px = (event.clientX - r.left) / r.width - 0.5
      const py = (event.clientY - r.top) / r.height - 0.5
      set(zoomBy(get(), Math.exp(-event.deltaY * 0.0015), px, py))
    }

    function down(event) {
      if (event.button !== 0 || !get()) return
      last = { x: event.clientX, y: event.clientY }
      node.setPointerCapture(event.pointerId)
    }

    function move(event) {
      const v = get()
      if (!last || !v) return
      const r = node.getBoundingClientRect()
      set(clampView({ ...v, x: v.x + (event.clientX - last.x) / r.width, y: v.y + (event.clientY - last.y) / r.height }))
      last = { x: event.clientX, y: event.clientY }
    }

    function up() {
      last = null
    }

    node.addEventListener('wheel', wheel, { passive: false })
    node.addEventListener('pointerdown', down)
    node.addEventListener('pointermove', move)
    node.addEventListener('pointerup', up)
    node.addEventListener('pointercancel', up)
    return () => {
      node.removeEventListener('wheel', wheel)
      node.removeEventListener('pointerdown', down)
      node.removeEventListener('pointermove', move)
      node.removeEventListener('pointerup', up)
      node.removeEventListener('pointercancel', up)
    }
  }
}

// At most one call per interval, always with the latest arguments — a drag
// sends its position to Go a few dozen times a second, not on every pixel.
export function throttle(fn, ms) {
  let timer = null
  let latest = []
  return (...args) => {
    latest = args
    if (timer) return
    timer = setTimeout(() => {
      timer = null
      fn(...latest)
    }, ms)
  }
}
