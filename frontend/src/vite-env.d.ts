/// <reference types="svelte" />
/// <reference types="vite/client" />

// Wails injects its runtime onto `window` at load time, and the generated
// `wailsjs/runtime/runtime.js` reaches for `window.runtime` directly — so without
// this declaration every call in it is a type error, in a file we don't own and
// that `wails build` overwrites anyway. The shape is the generated runtime module's
// own, so the two can't drift.
//
// The bindings under `wailsjs/go/` need no equivalent: they use `window['go'][...]`,
// and element access on `Window` is already untyped.
interface Window {
  runtime: typeof import('../wailsjs/runtime/runtime')
}
