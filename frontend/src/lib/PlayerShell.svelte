<script>
  import Shell from './Shell.svelte'

  const sections = [
    { id: 'characters', label: 'Characters', phase: 5, blurb: 'Create and manage characters — class, subclass, ancestry, community, traits.' },
    { id: 'sheet', label: 'Sheet', phase: 5, blurb: 'Live-tracked HP, stress, Hope, evasion, armor, thresholds, and gold.' },
    { id: 'cards', label: 'Domain Cards', phase: 5, blurb: 'Move cards between your loadout (max 5) and your vault.' },
    { id: 'inventory', label: 'Inventory', phase: 5, blurb: 'Gear, equipment, and consumables.' },
    { id: 'dice', label: 'Dice', phase: 5, blurb: 'Duality dice with your trait and Experience modifiers, wired to Hope.' }
  ]
</script>

<Shell {sections} />
