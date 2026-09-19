<script>
  import Shell from './Shell.svelte'
  import Encounters from './gm/Encounters.svelte'
  import Parties from './gm/Parties.svelte'
  import Adversaries from './gm/Adversaries.svelte'
  import Environments from './gm/Environments.svelte'
  import Equipment from './gm/Equipment.svelte'
  import Loot from './gm/Loot.svelte'
  import Dice from './gm/Dice.svelte'
  import CombatRunner from './gm/CombatRunner.svelte'
  import Campaigns from './gm/Campaigns.svelte'
  import Search from './gm/Search.svelte'

  const sections = [
    { id: 'campaigns', label: 'Campaigns', component: Campaigns },
    { id: 'parties', label: 'Parties', component: Parties },
    { id: 'encounters', label: 'Encounters', component: Encounters },
    { id: 'combat', label: 'Combat Runner', component: CombatRunner },
    { id: 'adversaries', label: 'Adversaries', component: Adversaries },
    { id: 'environments', label: 'Environments', component: Environments },
    { id: 'equipment', label: 'Equipment', component: Equipment },
    { id: 'loot', label: 'Loot', component: Loot },
    { id: 'dice', label: 'Dice', component: Dice },
    { id: 'search', label: 'Search', component: Search }
  ]
</script>

<Shell {sections} />
