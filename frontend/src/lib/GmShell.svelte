<script>
  import Shell from './Shell.svelte'
  import Encounters from './gm/Encounters.svelte'
  import Parties from './gm/Parties.svelte'
  import Adversaries from './gm/Adversaries.svelte'
  import Environments from './gm/Environments.svelte'
  import Dice from './gm/Dice.svelte'

  const sections = [
    { id: 'encounters', label: 'Encounters', component: Encounters },
    { id: 'parties', label: 'Parties', component: Parties },
    { id: 'adversaries', label: 'Adversaries', component: Adversaries },
    { id: 'environments', label: 'Environments', component: Environments },
    { id: 'dice', label: 'Dice', component: Dice },
    { id: 'combat', label: 'Combat Runner', phase: 3, blurb: 'Run a saved encounter live — HP, stress, spotlight, Fear, and countdowns.' },
    { id: 'campaigns', label: 'Campaigns', phase: 4, blurb: 'Sessions, recaps, and typed notes for NPCs, locations, and plot threads.' }
  ]
</script>

<Shell {sections} />
