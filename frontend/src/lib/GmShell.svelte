<script>
  import Shell from './Shell.svelte'
  import Encounters from './gm/Encounters.svelte'
  import Parties from './gm/Parties.svelte'
  import Adversaries from './gm/Adversaries.svelte'
  import Environments from './gm/Environments.svelte'
  import Dice from './gm/Dice.svelte'
  import CombatRunner from './gm/CombatRunner.svelte'
  import Campaigns from './gm/Campaigns.svelte'
  import Search from './gm/Search.svelte'

  const sections = [
    { id: 'campaigns', label: 'Campaigns', component: Campaigns },
    { id: 'encounters', label: 'Encounters', component: Encounters },
    { id: 'parties', label: 'Parties', component: Parties },
    { id: 'adversaries', label: 'Adversaries', component: Adversaries },
    { id: 'environments', label: 'Environments', component: Environments },
    { id: 'dice', label: 'Dice', component: Dice },
    { id: 'combat', label: 'Combat Runner', component: CombatRunner },
    { id: 'search', label: 'Search', component: Search }
  ]
</script>

<Shell {sections} />
