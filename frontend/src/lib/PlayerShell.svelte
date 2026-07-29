<script>
  import Shell from './Shell.svelte'
  import Characters from './player/Characters.svelte'
  import Sheet from './player/Sheet.svelte'
  import DomainCards from './player/DomainCards.svelte'
  import Inventory from './player/Inventory.svelte'
  import DualityDice from './player/DualityDice.svelte'
  import Beastform from './player/Beastform.svelte'
  import Companion from './player/Companion.svelte'
  import { active } from './player/active.svelte.js'
  import { BEASTFORM_CLASS, COMPANION_SUBCLASS, GetCharacter } from './player/api.js'

  // The two subclass-specific pages only exist for the characters that have them, so
  // the nav follows whoever is open rather than listing a Beastform page at every bard.
  let character = $state(null)

  $effect(() => {
    const id = active.id
    if (!id) {
      character = null
      return
    }
    GetCharacter(id)
      .then((c) => (character = c))
      .catch(() => (character = null))
  })

  const beastformer = $derived(
    character?.classSlug === BEASTFORM_CLASS || character?.multiclassSlug === BEASTFORM_CLASS
  )
  const beastbound = $derived(
    character?.subclassSlug === COMPANION_SUBCLASS ||
      character?.multiclassSubclassSlug === COMPANION_SUBCLASS
  )

  // Order of use: pick a character, read the sheet, then the things you touch during
  // play. Everything after Characters works on whatever is selected there.
  const sections = $derived([
    { id: 'characters', label: 'Characters', component: Characters },
    { id: 'sheet', label: 'Sheet', component: Sheet },
    { id: 'cards', label: 'Domain Cards', component: DomainCards },
    ...(beastformer ? [{ id: 'beastform', label: 'Beastform', component: Beastform }] : []),
    ...(beastbound ? [{ id: 'companion', label: 'Companion', component: Companion }] : []),
    { id: 'inventory', label: 'Inventory', component: Inventory },
    { id: 'dice', label: 'Dice', component: DualityDice }
  ])
</script>

<Shell {sections} />
