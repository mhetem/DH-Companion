// Every length in the app is in rem, so the root font size is the one lever that
// scales the layout along with the text — controls, gutters and the prose max-widths
// all grow together instead of text swelling inside fixed-size chrome.
export function applyScale(percent) {
  document.documentElement.style.fontSize = `${percent}%`
}
