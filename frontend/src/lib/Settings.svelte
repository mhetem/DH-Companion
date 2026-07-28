<script>
  import Modal from './Modal.svelte'
  import {
    CheckForUpdate,
    DataDirectory,
    ExportDatabase,
    ExportLibrary,
    ImportDatabase,
    ImportLibrary,
    OpenReleasesPage,
    Version
  } from '../../wailsjs/go/main/App.js'
  import { errorMessage } from './gm/api.js'

  let { onreload } = $props()

  let open = $state(false)
  let busy = $state('')
  let error = $state('')
  let done = $state('')

  let version = $state('')
  let dataDir = $state('')
  let release = $state(null)

  function show() {
    open = true
    error = ''
    done = ''
    Version().then((v) => (version = v)).catch(() => {})
    DataDirectory().then((d) => (dataDir = d)).catch(() => {})
  }

  // Every action reports into the same two lines, so the panel never grows a
  // per-button status of its own.
  async function run(label, fn) {
    busy = label
    error = ''
    done = ''
    try {
      done = (await fn()) ?? ''
    } catch (e) {
      error = errorMessage(e)
    } finally {
      busy = ''
    }
  }

  const backup = () =>
    run('backup', async () => {
      const path = await ExportDatabase()
      return path ? `Backup written to ${path}` : ''
    })

  const restore = () =>
    run('restore', async () => {
      if (!confirm('Replace your current database with the one you pick?\n\nThe current database is backed up first, and the app reloads from the new one.')) {
        return ''
      }
      const saved = await ImportDatabase()
      if (!saved) return ''
      onreload?.()
      return `Database replaced. Your previous data was saved to ${saved}`
    })

  const exportLibrary = () =>
    run('export', async () => {
      const path = await ExportLibrary()
      return path ? `Library exported to ${path}` : ''
    })

  const importLibrary = () =>
    run('import', async () => {
      const report = await ImportLibrary()
      if (!report) return ''
      onreload?.()
      const counts = [
        [report.parties, 'party', 'parties'],
        [report.customAdversaries, 'adversary', 'adversaries'],
        [report.customEnvironments, 'environment', 'environments'],
        [report.encounters, 'encounter', 'encounters'],
        [report.campaigns, 'campaign', 'campaigns'],
        [report.sessions, 'session', 'sessions'],
        [report.notes, 'note', 'notes'],
        [report.countdowns, 'countdown', 'countdowns']
      ]
        .filter(([n]) => n > 0)
        .map(([n, one, many]) => `${n} ${n === 1 ? one : many}`)
      let msg = counts.length ? `Imported ${counts.join(', ')}.` : 'Nothing new to import.'
      if (report.renamed?.length) msg += ` Renamed to avoid collisions: ${report.renamed.join(', ')}.`
      if (report.skipped?.length) msg += ` Skipped ${report.skipped.length}: ${report.skipped.join('; ')}`
      return msg
    })

  const checkUpdate = () =>
    run('update', async () => {
      release = await CheckForUpdate()
      return ''
    })
</script>

<button class="iconbtn" onclick={show} title="Settings and data" aria-label="Settings">⚙</button>

<Modal bind:open label="Settings" width="40rem">
  <header>
    <h2>Settings &amp; data</h2>
    <button class="btn ghost" onclick={() => (open = false)}>Close</button>
  </header>

  {#if error}<p class="error">{error}</p>{/if}
  {#if done}<p class="done">{done}</p>{/if}

  <section>
    <h3>Backup</h3>
    <p class="note">
      A backup is a consistent snapshot of the whole database — everything, exactly as it
      stands. Restoring replaces what you have now, so the current database is saved
      alongside it first.
    </p>
    <div class="row">
      <button class="btn" onclick={backup} disabled={!!busy}>
        {busy === 'backup' ? 'Working…' : 'Back up database'}
      </button>
      <button class="btn" onclick={restore} disabled={!!busy}>
        {busy === 'restore' ? 'Working…' : 'Restore from backup'}
      </button>
    </div>
  </section>

  <section>
    <h3>Library</h3>
    <p class="note">
      A readable JSON export of your parties, homebrew cards, encounters and campaigns.
      Importing <strong>adds</strong> to what you have rather than replacing it, and
      renames anything whose name is already taken.
    </p>
    <div class="row">
      <button class="btn" onclick={exportLibrary} disabled={!!busy}>
        {busy === 'export' ? 'Working…' : 'Export library'}
      </button>
      <button class="btn" onclick={importLibrary} disabled={!!busy}>
        {busy === 'import' ? 'Working…' : 'Import library'}
      </button>
    </div>
  </section>

  <section>
    <h3>Version</h3>
    <p class="note">
      Hilt {version || '…'}{#if version === 'dev'} — an unreleased build, so update checks
        have nothing to compare against{/if}.
      Checking asks GitHub for the latest release; nothing is sent, and nothing is
      checked unless you press the button.
    </p>
    <div class="row">
      <button class="btn" onclick={checkUpdate} disabled={!!busy}>
        {busy === 'update' ? 'Checking…' : 'Check for updates'}
      </button>
      {#if release?.newer}
        <button class="btn primary" onclick={() => OpenReleasesPage(release.url)}>
          Get {release.latest}
        </button>
      {/if}
    </div>
    {#if release}
      <p class="note result">
        {#if release.newer}
          <strong>{release.latest}</strong> is out — you have {release.current}.
        {:else if release.known}
          You are on the latest release ({release.current}).
        {:else}
          Latest release is {release.latest || 'unknown'}; this build reports
          “{release.current}”, which can't be compared.
        {/if}
      </p>
    {/if}
  </section>

  <section>
    <h3>Where your data lives</h3>
    <p class="path">{dataDir || '…'}</p>
    <p class="note">Set <code>DH_DATA_DIR</code> before launching to keep it somewhere else.</p>
  </section>
</Modal>

<style>
  header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
  }

  h2 {
    margin: 0;
    font-size: 1.05rem;
  }

  h3 {
    margin: 0 0 0.3rem;
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    color: var(--gold);
  }

  .note {
    margin: 0 0 0.6rem;
    font-size: 0.78rem;
    line-height: 1.5;
    color: var(--muted);
  }

  .note.result {
    margin: 0.5rem 0 0;
    color: var(--text);
  }

  .row {
    display: flex;
    flex-wrap: wrap;
    gap: 0.5rem;
  }

  .path {
    margin: 0 0 0.3rem;
    padding: 0.4rem 0.6rem;
    border: 1px solid var(--line);
    border-radius: 6px;
    background: var(--bg);
    font-size: 0.78rem;
    word-break: break-all;
  }

  code {
    padding: 0.05rem 0.3rem;
    border-radius: 4px;
    background: var(--bg);
    font-size: 0.72rem;
  }

  .error,
  .done {
    margin: 0;
    padding: 0.5rem 0.75rem;
    border-radius: 6px;
    font-size: 0.8rem;
    line-height: 1.5;
  }

  .error {
    border: 1px solid var(--danger);
    color: var(--danger);
  }

  .done {
    border: 1px solid var(--ok);
    color: var(--ok);
  }
</style>
