// Single import surface for the GM module. Everything here is bound on the Go
// side as window.go.gm.Service.* — the generated wrappers just forward to it.
export {
  AdjustFear,
  AdvanceCountdown,
  BrowseAdversaries,
  BrowseEnvironments,
  ClearSpotlights,
  ComputeBudget,
  CreateCustomAdversary,
  CreateCustomEnvironment,
  CreateParty,
  DeleteCombat,
  DeleteCountdown,
  DeleteCustomAdversary,
  DeleteCustomEnvironment,
  DeleteEncounter,
  DeleteParty,
  EndCombat,
  GetActiveCombat,
  GetAdversary,
  GetCombat,
  GetCountdown,
  GetCustomAdversary,
  GetCustomEnvironment,
  GetEncounter,
  GetEnvironment,
  GetParty,
  ListCombats,
  ListCountdowns,
  ListCustomAdversaries,
  ListCustomEnvironments,
  ListEncounters,
  ListParties,
  MarkHP,
  MarkStress,
  RemoveCombatant,
  ResumeCombat,
  SaveCombatant,
  SaveCountdown,
  SaveEncounter,
  SetFear,
  SetSpotlight,
  SetVitals,
  StartCombat,
  UpdateCustomAdversary,
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
