// Single import surface for the GM module. Everything here is bound on the Go
// side as window.go.gm.Service.* — the generated wrappers just forward to it.
export {
  AdjustCampaignFear,
  AdjustFear,
  AdvanceCountdown,
  BrowseAdversaries,
  BrowseArmor,
  BrowseConsumables,
  BrowseItems,
  BrowseEnvironments,
  BrowseWeapons,
  ClearSpotlights,
  ComputeBudget,
  CreateCustomAdversary,
  CreateCustomArmor,
  CreateCustomConsumable,
  CreateCustomItem,
  CreateCustomWeapon,
  CreateCustomEnvironment,
  CreateParty,
  DeleteCampaign,
  DeleteCombat,
  DeleteCountdown,
  DeleteCustomAdversary,
  DeleteCustomArmor,
  DeleteCustomConsumable,
  DeleteCustomItem,
  DeleteCustomWeapon,
  DeleteCustomEnvironment,
  DeleteEncounter,
  DeleteNote,
  DeleteParty,
  DeleteSession,
  EndCombat,
  GetActiveCombat,
  GetArmor,
  GetAdversary,
  GetCampaign,
  GetCombat,
  GetConsumable,
  GetCountdown,
  GetCustomAdversary,
  GetCustomArmor,
  GetCustomConsumable,
  GetCustomItem,
  GetCustomWeapon,
  GetCustomEnvironment,
  GetEncounter,
  GetItem,
  GetEnvironment,
  GetMasterNote,
  GetNote,
  GetParty,
  GetWeapon,
  GetSession,
  ImportShareCode,
  LinkCombat,
  LinkEncounter,
  ListCampaigns,
  ListCombats,
  ListCountdowns,
  ListCountdownsForCampaign,
  ListCustomAdversaries,
  ListCustomEnvironments,
  ListEncounters,
  ListNotes,
  ListNotesByKind,
  ListParties,
  ListSessions,
  ListUnassignedCountdowns,
  LookupLoot,
  MarkHP,
  PreviewShareCode,
  MarkStress,
  RemoveCombatant,
  ResumeCombat,
  RollLoot,
  SaveCampaign,
  SaveCombatant,
  SaveCountdown,
  SaveEncounter,
  SaveMasterNote,
  SaveNote,
  SaveSession,
  Search,
  SessionsForEncounter,
  SetCampaignFear,
  SetFear,
  SetSpotlight,
  SetVitals,
  ShareAdversary,
  ShareArmor,
  ShareConsumable,
  ShareItem,
  ShareWeapon,
  ShareEnvironment,
  StartCombat,
  UnlinkEncounter,
  UpdateCustomAdversary,
  UpdateCustomArmor,
  UpdateCustomConsumable,
  UpdateCustomItem,
  UpdateCustomWeapon,
  UpdateCustomEnvironment,
  UpdateParty
} from '../../../wailsjs/go/gm/Service.js'

// The roller is bound as its own struct too — window.go.dice.Roller.*. Sizes()
// serves the rollable die list, so the allowed set lives only in internal/dice.
export { Damage, GM, Sizes } from '../../../wailsjs/go/dice/Roller.js'

// Mirrors internal/gm/validate.go — keep the two in step.
export const TIERS = ['1', '2', '3', '4']

export const ADVERSARY_TYPES = [
  'Bruiser',
  'Horde',
  'Leader',
  'Minion',
  'Ranged',
  'Skulk',
  'Social',
  'Solo',
  'Standard',
  'Support'
]

export const ENVIRONMENT_TYPES = ['Event', 'Exploration', 'Social', 'Traversal']

// Mirrors internal/cards/cards.go — a weapon is Primary or Secondary, and deals
// physical or magic damage. Combat wheelchairs are equipped as primary weapons,
// so they need no category of their own.
export const WEAPON_CATEGORIES = ['Primary', 'Secondary']

export const WEAPON_TYPES = ['Physical', 'Magic']

// The rest of a weapon's closed sets. These aren't validated by the backend the
// way category and type are — the SRD's own tables are the only authority — so
// they're here purely to keep the homebrew form's pickers honest.
export const WEAPON_TRAITS = [
  'Agility',
  'Strength',
  'Finesse',
  'Instinct',
  'Presence',
  'Knowledge',
  'Spellcast'
]

export const WEAPON_RANGES = ['Melee', 'Very Close', 'Close', 'Far', 'Very Far']

export const WEAPON_BURDENS = ['One-Handed', 'Two-Handed']

// The SRD abbreviates these in its tables; the value is what crosses the bridge.
export const DAMAGE_TYPES = [
  { value: 'phy', label: 'Physical' },
  { value: 'mag', label: 'Magic' },
  { value: 'phy/mag', label: 'Either' }
]

// Mirrors maxArmorScore in internal/gm/validate.go.
export const MAX_ARMOR_SCORE = 12

// Mirrors internal/rules/loot.go — keep the two in step.
//
// Rarity in the SRD describes the roll, not the entry: it names how many d12s to
// throw, and the sum indexes a 60-row table. Both dice counts are offered, so the
// roller lets the GM pick — the smaller keeps results low in the table, the larger
// spreads them across the whole band.
export const LOOT_RARITIES = [
  { name: 'Common', minDice: 1, maxDice: 2 },
  { name: 'Uncommon', minDice: 2, maxDice: 3 },
  { name: 'Rare', minDice: 3, maxDice: 4 },
  { name: 'Legendary', minDice: 4, maxDice: 5 }
]

// The source books the item and consumable tables come from, as data/items.json
// and data/consumables.json label them.
export const LOOT_TABLES = ['Core Set', 'Hope & Fear']

// Homebrew loot has no roll number, so it isn't one of the books above — it's a
// separate bucket the browser can filter to. Mirrors homebrewTable in
// internal/gm/loot.go.
export const HOMEBREW_TABLE = 'Homebrew'

export function rarityDice(name) {
  const rarity = LOOT_RARITIES.find((r) => r.name === name)
  if (!rarity) return ''
  return `${rarity.minDice}d12 or ${rarity.maxDice}d12`
}

// Note kinds are a closed set in SQL (notes.kind CHECK) as well as in validate.go.
// Labels are display-only; the value is what crosses the bridge.
export const NOTE_KINDS = [
  { value: 'npc', label: 'NPC' },
  { value: 'location', label: 'Location' },
  { value: 'faction', label: 'Faction' },
  { value: 'lore', label: 'Lore' },
  { value: 'plot', label: 'Plot thread' }
]

export function noteKindLabel(kind) {
  return NOTE_KINDS.find((k) => k.value === kind)?.label ?? kind
}

// Not validated by the backend, but every SRD feature uses one of these three.
export const FEATURE_TYPES = ['Action', 'Passive', 'Reaction']

// Feature.common tags a card with a shared keyword the rules define elsewhere.
// These are the ones the SRD data actually uses; the field stays free text.
export const COMMON_FEATURES = [
  'groupAttack',
  'horde',
  'minion',
  'momentum',
  'relentless',
  'slow',
  'terrifying'
]

// Difficulty is a builder-side knob only: it shifts the budget but isn't part of
// the saved encounter, so the values match rules.EncounterSettings.Difficulty.
export const DIFFICULTIES = [
  { value: '', label: 'Standard' },
  { value: 'easy', label: 'Easy' },
  { value: 'hard', label: 'Hard' }
]

// Battle-point cost per adversary type, mirroring rules.adversaryCost. Minions
// are batched party-sized by the budget math, so they have no flat cost here.
const COSTS = {
  Social: 1,
  Support: 1,
  Standard: 2,
  Horde: 2,
  Skulk: 2,
  Ranged: 2,
  Leader: 3,
  Bruiser: 4,
  Solo: 5
}

export function costLabel(type) {
  if (type === 'Minion') return 'batched'
  return `${COSTS[type] ?? 2} pt${(COSTS[type] ?? 2) === 1 ? '' : 's'}`
}

// Wails rejects with a plain string; anything else is a real JS error.
export function errorMessage(e) {
  if (typeof e === 'string') return e
  return e?.message ?? String(e)
}
