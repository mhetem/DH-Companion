// Single import surface for the Player module. Everything here is bound on the Go
// side as window.go.player.Service.* — the generated wrappers just forward to it.
export {
  AddClassItems,
  AddDomainCard,
  AdjustHope,
  AdjustItemQty,
  ApplyLevelUp,
  AvailableDomainCards,
  Beastforms,
  BrowseDomainCards,
  DeleteCharacter,
  DeleteCompanion,
  DropBeastform,
  DeleteExperience,
  DeleteItem,
  DomainList,
  GetAncestry,
  GetCharacter,
  GetClass,
  GetCommunity,
  GetCompanion,
  GetDomainCard,
  GetLoadout,
  GetSheet,
  ItemKinds,
  Limits,
  ListAncestries,
  ListBeastforms,
  ListCharacters,
  ListClasses,
  ListCommunities,
  ListCompanions,
  ListExperiences,
  ListInventory,
  ListLevelUps,
  MarkArmor,
  MarkCompanionStress,
  MarkHP,
  MarkStress,
  MoveDomainCard,
  PlanLevelUp,
  RemoveDomainCard,
  Rest,
  RestAllowance,
  RestMoves,
  RollCompanionDamage,
  RollDamage,
  RollDuality,
  SaveCharacter,
  SaveCompanion,
  SaveExperience,
  SaveItem,
  SetCompanionStress,
  SetGold,
  SetItemEquipped,
  SetVitals,
  ToggleCompanionUpgrade,
  Traits,
  Transform
} from '../../../wailsjs/go/player/Service.js'

// The die sizes come from internal/dice, same as the GM side — the damage picker
// can't offer a die the backend would reject because the set is defined once there.
export { Sizes } from '../../../wailsjs/go/dice/Roller.js'

// Mirrors internal/player/validate.go — keep the two in step.
export const TRAITS = [
  { key: 'agility', label: 'Agility', blurb: 'Sprint, Leap, Maneuver' },
  { key: 'strength', label: 'Strength', blurb: 'Lift, Smash, Grapple' },
  { key: 'finesse', label: 'Finesse', blurb: 'Control, Hide, Tinker' },
  { key: 'instinct', label: 'Instinct', blurb: 'Perceive, Sense, Navigate' },
  { key: 'presence', label: 'Presence', blurb: 'Charm, Perform, Deceive' },
  { key: 'knowledge', label: 'Knowledge', blurb: 'Recall, Analyze, Comprehend' }
]

// The level 1 spread every character distributes across the six traits.
export const TRAIT_ARRAY = [2, 1, 1, 0, 0, -1]

export const ITEM_KINDS = [
  { value: 'primary-weapon', label: 'Primary weapon' },
  { value: 'secondary-weapon', label: 'Secondary weapon' },
  { value: 'armor', label: 'Armor' },
  { value: 'consumable', label: 'Consumable' },
  { value: 'item', label: 'Item' }
]

// Only one of each of these can be equipped at a time; the backend unequips the rest.
export const EXCLUSIVE_KINDS = ['primary-weapon', 'secondary-weapon', 'armor']

export const DOMAINS = [
  'Arcana',
  'Blade',
  'Bone',
  'Codex',
  'Grace',
  'Midnight',
  'Sage',
  'Splendor',
  'Valor'
]

export const CARD_LEVELS = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10']

export const HOPE_MAX = 6
export const LOADOUT_MAX = 5
export const MAX_LEVEL = 10

export const MASTERY_LABELS = ['Foundation', 'Specialization', 'Mastery']

// Mirrors internal/player — the Beastform class feature is the druid's, and the
// companion sheet belongs to the Beastbound ranger.
export const BEASTFORM_CLASS = 'druid'
export const COMPANION_SUBCLASS = 'beastbound'

export function itemKindLabel(kind) {
  return ITEM_KINDS.find((k) => k.value === kind)?.label ?? kind
}

export function traitLabel(key) {
  return TRAITS.find((t) => t.key === key)?.label ?? key
}

export function signed(n) {
  return n >= 0 ? `+${n}` : `${n}`
}

// Subclass features are printed in three groups and unlock as the character
// advances, so a level 3 character shouldn't be reading their Mastery card yet.
export function featuresUpTo(features = [], mastery = 1) {
  const unlocked = MASTERY_LABELS.slice(0, Math.max(1, Math.min(3, mastery)))
  return features.filter((f) => !MASTERY_LABELS.includes(f.type) || unlocked.includes(f.type))
}

export function goldLabel(gold) {
  if (!gold) return '—'
  const parts = []
  if (gold.chests) parts.push(`${gold.chests} chest${gold.chests === 1 ? '' : 's'}`)
  if (gold.bags) parts.push(`${gold.bags} bag${gold.bags === 1 ? '' : 's'}`)
  if (gold.handfuls) parts.push(`${gold.handfuls} handful${gold.handfuls === 1 ? '' : 's'}`)
  return parts.length ? parts.join(' · ') : 'no gold'
}

// Wails rejects with a plain string; anything else is a real JS error.
export function errorMessage(e) {
  if (typeof e === 'string') return e
  return e?.message ?? String(e)
}
