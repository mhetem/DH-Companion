<script>
  import { ListSessions, errorMessage } from './api.js'

  // Picking a session implies its campaign — the Go side overrides campaign_id from
  // the session, so the campaign select goes read-only rather than lying about it.
  let { campaigns = [], campaignId = null, sessionId = null, busy = false, onchange } = $props()

  let sessions = $state([])
  let error = $state('')

  // Fully controlled: the parent owns the pair and hands it back down, so this can't
  // drift from the combat it is editing. Changing campaign clears the session, since
  // a session from the old campaign would be overridden server-side anyway.
  const campaignPick = $derived(campaignId === null ? '' : String(campaignId))
  const sessionPick = $derived(sessionId === null ? '' : String(sessionId))

  $effect(() => {
    const id = campaignId
    if (!id) {
      sessions = []
      return
    }
    let stale = false
    ListSessions(id)
      .then((rows) => {
        if (!stale) sessions = rows ?? []
      })
      .catch((e) => {
        if (!stale) error = errorMessage(e)
      })
    return () => (stale = true)
  })

  const pickCampaign = (value) => onchange?.(value ? Number(value) : null, null)
  const pickSession = (value) => onchange?.(campaignId, value ? Number(value) : null)
</script>

<div class="links">
  <label>
    <span>Campaign</span>
    <select value={campaignPick} onchange={(e) => pickCampaign(e.currentTarget.value)} disabled={busy}>
      <option value="">No campaign — this fight keeps its own Fear</option>
      {#each campaigns as campaign (campaign.id)}
        <option value={String(campaign.id)}>{campaign.name} · Fear {campaign.currentFear}/{campaign.fearMax}</option>
      {/each}
    </select>
  </label>

  <label>
    <span>Session</span>
    <select
      value={sessionPick}
      onchange={(e) => pickSession(e.currentTarget.value)}
      disabled={busy || !campaignPick}
    >
      <option value="">{campaignPick ? 'Not logged to a session' : 'Pick a campaign first'}</option>
      {#each sessions as session (session.id)}
        <option value={String(session.id)}>#{session.number} {session.title}</option>
      {/each}
    </select>
  </label>

  {#if error}
    <p class="error">{error}</p>
  {/if}
</div>

<style>
  .links {
    display: flex;
    gap: 0.6rem;
    flex-wrap: wrap;
  }

  label {
    display: flex;
    flex: 1;
    flex-direction: column;
    gap: 0.25rem;
    min-width: 12rem;
  }

  label span {
    font-size: 0.7rem;
    text-transform: uppercase;
    letter-spacing: 0.07em;
    color: var(--muted);
  }

  .error {
    flex-basis: 100%;
    margin: 0;
    font-size: 0.75rem;
    color: var(--danger);
  }
</style>
