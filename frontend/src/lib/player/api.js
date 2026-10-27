// Single import surface for the Player module. Everything here is a method on
// player.Service, bound as a Wails service — the generated wrappers just forward to it.
export {
  AddClassItems,
  AddDomainCard,
  AdjustHope,
  AdjustItemQty,
  ApplyLevelUp,
  AvailableDomainCards,
  Beastforms,
  BrowseDomainCards,
  CreationBonuses,
  DeleteCharacter,
  DeleteCompanion,
  DropBeastform,
  DropStance,
  DeleteExperience,
  DeleteItem,
  DomainList,
  FavorInsteadOfHope,
  GetAncestry,
  GetCharacter,
  GetClass,
  GetCommunity,
  GetCompanion,
  GetDomainCard,
  GetLoadout,
  GetSheet,
  ItemKinds,
  LearnStance,
  Limits,
  ListAncestries,
  ListBeastforms,
  ListCharacters,
  ListClasses,
  ListCommunities,
  ListCompanions,
  ListExperiences,
  ListGear,
  ListInventory,
  ListLevelUps,
  ListTransformations,
  MarkArmor,
  MarkCompanionStress,
  MarkHP,
  MarkStress,
  MoveDomainCard,
  PlanLevelUp,
  Refocus,
  RemoveDomainCard,
  Rest,
  RestAllowance,
  RestMoves,
  RollBrawlersStrike,
  RollCombo,
  RollCompanionDamage,
  RollDamage,
  RollDuality,
  RollWeaponDamage,
  SaveCharacter,
  SaveCompanion,
  SaveExperience,
  SaveItem,
  SetClassTracks,
  SetCompanionStress,
  SetEffect,
  SetGold,
  SetItemEquipped,
  SetTransformation,
  SetVitals,
  ShiftStance,
  SwapDomainCard,
  TakeDamage,
  ToggleCompanionUpgrade,
  Traits,
  Transform
} from '../../../bindings/github.com/mhetem/DH-Companion/internal/player/service.js'

// The die sizes come from internal/dice, same as the GM side — the damage picker
// can't offer a die the backend would reject because the set is defined once there.
export { Sizes } from '../../../bindings/github.com/mhetem/DH-Companion/internal/dice/roller.js'

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
  'Dread',
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

// The SRD gear a kind of inventory item can be picked from. Weapons are split by
// category the same way the backend checks them, so a picked slug never fails
// validation on save.
export function gearFor(gear, kind) {
  if (!gear) return []
  if (kind === 'primary-weapon') return gear.weapons.filter((w) => w.category === 'Primary')
  if (kind === 'secondary-weapon') return gear.weapons.filter((w) => w.category === 'Secondary')
  if (kind === 'armor') return gear.armor
  if (kind === 'consumable') return gear.consumables
  return gear.items
}

// One line of the card's stats, as the SRD tables print them.
export function gearLine(card) {
  if (!card) return ''
  if (card.damage) {
    return [`Tier ${card.tier}`, card.trait, card.range, `${card.damage} ${card.damageType}`, card.burden].join(' · ')
  }
  if (card.baseScore !== undefined) {
    return `Tier ${card.tier} · Armor ${card.baseScore} · Thresholds ${card.thresholdMajor}/${card.thresholdSevere} + level`
  }
  return card.rarity ? `${card.rarity} · ${card.table}` : card.table ?? ''
}

// Mirrors usesFavor, hasFlowState and the brawler checks in internal/player — the
// dice panel runs standalone too, where there's no sheet to carry the flags.
export const usesFavor = (c) => c?.classSlug === 'warlock' || c?.multiclassSlug === 'warlock'
export const hasFlowState = (c) => c?.subclassSlug === 'martialArtist' && c?.subclassMastery >= 3
export const isBrawler = (c) => c?.classSlug === 'brawler'

// A weapon names the trait its attack rolls with. "Spellcast" isn't a trait of its
// own — it means whichever one the subclass casts with, and a subclass that
// doesn't cast leaves nothing to roll.
export function weaponTrait(weapon, character) {
  const name = weapon?.trait === 'Spellcast' ? character?.spellcastTrait : weapon?.trait
  return (name ?? '').toLowerCase()
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

// Recall Cost is printed on the card as text, and a card that doesn't charge one
// carries "0" rather than an empty string.
export function recallCost(card) {
  return Number(card?.recallCost) || 0
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
