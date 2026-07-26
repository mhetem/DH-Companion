<script>
  import Shell from './Shell.svelte'

  const sections = [
    { id: 'encounters', label: 'Encounters', phase: 2, blurb: 'Build encounters against a party and watch the budget update live.' },
    { id: 'adversaries', label: 'Adversaries', phase: 2, blurb: 'Browse the SRD roster and your own homebrew adversaries.' },
    { id: 'environments', label: 'Environments', phase: 2, blurb: 'Browse environments and attach one to an encounter.' },
    { id: 'combat', label: 'Combat Runner', phase: 3, blurb: 'Run a saved encounter live — HP, stress, spotlight, Fear, and countdowns.' },
    { id: 'campaigns', label: 'Campaigns', phase: 4, blurb: 'Sessions, recaps, and typed notes for NPCs, locations, and plot threads.' }
  ]
</script>

<Shell {sections} />
