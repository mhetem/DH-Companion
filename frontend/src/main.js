import './style.css'
import { mount } from 'svelte'
import App from './App.svelte'
import Table from './lib/table/Table.svelte'

// The player window loads the same bundle at /?table. Its page turns black before
// anything mounts, so the app's purple never flashes on the players' screen.
const table = new URLSearchParams(location.search).has('table')
if (table) document.documentElement.style.background = '#000'

const app = mount(table ? Table : App, {
  target: document.getElementById('app')
})

export default app
