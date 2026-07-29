export namespace cards {
	
	export class Feature {
	    common: string;
	    title: string;
	    type: string;
	    description: string;
	    questions?: string[];
	
	    static createFrom(source: any = {}) {
	        return new Feature(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.common = source["common"];
	        this.title = source["title"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.questions = source["questions"];
	    }
	}
	export class Attack {
	    modifier: string;
	    name: string;
	    range: string;
	    damage: string;
	    damageType: string;
	
	    static createFrom(source: any = {}) {
	        return new Attack(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.modifier = source["modifier"];
	        this.name = source["name"];
	        this.range = source["range"];
	        this.damage = source["damage"];
	        this.damageType = source["damageType"];
	    }
	}
	export class Adversary {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    hordeNumber: string;
	    motives: string;
	    experiences: string;
	    difficulty: string;
	    thresholdMinor: string;
	    thresholdMajor: string;
	    hp: string;
	    stress: string;
	    standardAttack: Attack;
	    features: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Adversary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.hordeNumber = source["hordeNumber"];
	        this.motives = source["motives"];
	        this.experiences = source["experiences"];
	        this.difficulty = source["difficulty"];
	        this.thresholdMinor = source["thresholdMinor"];
	        this.thresholdMajor = source["thresholdMajor"];
	        this.hp = source["hp"];
	        this.stress = source["stress"];
	        this.standardAttack = this.convertValues(source["standardAttack"], Attack);
	        this.features = this.convertValues(source["features"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Ancestry {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    features: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Ancestry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.features = this.convertValues(source["features"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class BeastformAttack {
	    range: string;
	    trait: string;
	    damage: string;
	    damageType: string;
	
	    static createFrom(source: any = {}) {
	        return new BeastformAttack(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.range = source["range"];
	        this.trait = source["trait"];
	        this.damage = source["damage"];
	        this.damageType = source["damageType"];
	    }
	}
	export class Beastform {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    examples: string[];
	    trait: string;
	    traitBonus: string;
	    evasionBonus: string;
	    attack: BeastformAttack;
	    advantages: string[];
	    features: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Beastform(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.examples = source["examples"];
	        this.trait = source["trait"];
	        this.traitBonus = source["traitBonus"];
	        this.evasionBonus = source["evasionBonus"];
	        this.attack = this.convertValues(source["attack"], BeastformAttack);
	        this.advantages = source["advantages"];
	        this.features = this.convertValues(source["features"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class Subclass {
	    slug: string;
	    name: string;
	    tagline: string;
	    spellcastTrait: string;
	    features: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Subclass(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tagline = source["tagline"];
	        this.spellcastTrait = source["spellcastTrait"];
	        this.features = this.convertValues(source["features"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CharacterClass {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    domains: string[];
	    startingEvasion: string;
	    startingHitPoints: string;
	    classItems: string[];
	    hopeFeature: Feature;
	    features: Feature[];
	    subclasses: Subclass[];
	    backgroundQuestions: string[];
	    connections: string[];
	
	    static createFrom(source: any = {}) {
	        return new CharacterClass(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.domains = source["domains"];
	        this.startingEvasion = source["startingEvasion"];
	        this.startingHitPoints = source["startingHitPoints"];
	        this.classItems = source["classItems"];
	        this.hopeFeature = this.convertValues(source["hopeFeature"], Feature);
	        this.features = this.convertValues(source["features"], Feature);
	        this.subclasses = this.convertValues(source["subclasses"], Subclass);
	        this.backgroundQuestions = source["backgroundQuestions"];
	        this.connections = source["connections"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Community {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    adjectives: string[];
	    features: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Community(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.adjectives = source["adjectives"];
	        this.features = this.convertValues(source["features"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Companion {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    startingEvasion: string;
	    startingDamageDie: string;
	    startingRange: string;
	    startingExperienceModifier: string;
	    setup: Feature[];
	    exampleExperiences: string[];
	    rules: Feature[];
	    levelUpOptions: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Companion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.startingEvasion = source["startingEvasion"];
	        this.startingDamageDie = source["startingDamageDie"];
	        this.startingRange = source["startingRange"];
	        this.startingExperienceModifier = source["startingExperienceModifier"];
	        this.setup = this.convertValues(source["setup"], Feature);
	        this.exampleExperiences = source["exampleExperiences"];
	        this.rules = this.convertValues(source["rules"], Feature);
	        this.levelUpOptions = this.convertValues(source["levelUpOptions"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DomainCard {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    domain: string;
	    level: string;
	    recallCost: string;
	
	    static createFrom(source: any = {}) {
	        return new DomainCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.domain = source["domain"];
	        this.level = source["level"];
	        this.recallCost = source["recallCost"];
	    }
	}
	export class Environment {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    difficulty: string;
	    impulses: string;
	    potentialAdversaries: string[];
	    features: Feature[];
	
	    static createFrom(source: any = {}) {
	        return new Environment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.difficulty = source["difficulty"];
	        this.impulses = source["impulses"];
	        this.potentialAdversaries = source["potentialAdversaries"];
	        this.features = this.convertValues(source["features"], Feature);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	

}

export namespace dice {
	
	export class DamageRoll {
	    count: number;
	    sides: number;
	    modifier: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new DamageRoll(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.count = source["count"];
	        this.sides = source["sides"];
	        this.modifier = source["modifier"];
	        this.total = source["total"];
	    }
	}
	export class DualityDice {
	    Hope: number;
	    Fear: number;
	    Result: number;
	    Msg: string;
	
	    static createFrom(source: any = {}) {
	        return new DualityDice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Hope = source["Hope"];
	        this.Fear = source["Fear"];
	        this.Result = source["Result"];
	        this.Msg = source["Msg"];
	    }
	}
	export class GMDice {
	    Result: number;
	    Msg: string;
	
	    static createFrom(source: any = {}) {
	        return new GMDice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Result = source["Result"];
	        this.Msg = source["Msg"];
	    }
	}

}

export namespace gm {
	
	export class BrowseAdversary {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    hordeNumber: string;
	    motives: string;
	    experiences: string;
	    difficulty: string;
	    thresholdMinor: string;
	    thresholdMajor: string;
	    hp: string;
	    stress: string;
	    standardAttack: cards.Attack;
	    features: cards.Feature[];
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowseAdversary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.hordeNumber = source["hordeNumber"];
	        this.motives = source["motives"];
	        this.experiences = source["experiences"];
	        this.difficulty = source["difficulty"];
	        this.thresholdMinor = source["thresholdMinor"];
	        this.thresholdMajor = source["thresholdMajor"];
	        this.hp = source["hp"];
	        this.stress = source["stress"];
	        this.standardAttack = this.convertValues(source["standardAttack"], cards.Attack);
	        this.features = this.convertValues(source["features"], cards.Feature);
	        this.source = source["source"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BrowseEnvironment {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    difficulty: string;
	    impulses: string;
	    potentialAdversaries: string[];
	    features: cards.Feature[];
	    source: string;
	
	    static createFrom(source: any = {}) {
	        return new BrowseEnvironment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.difficulty = source["difficulty"];
	        this.impulses = source["impulses"];
	        this.potentialAdversaries = source["potentialAdversaries"];
	        this.features = this.convertValues(source["features"], cards.Feature);
	        this.source = source["source"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Campaign {
	    id: number;
	    name: string;
	    description: string;
	    currentFear: number;
	    fearMax: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Campaign(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.currentFear = source["currentFear"];
	        this.fearMax = source["fearMax"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class CampaignInput {
	    id?: number;
	    name: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new CampaignInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	    }
	}
	export class CombatSummary {
	    id: number;
	    encounterId?: number;
	    encounterName: string;
	    campaignId?: number;
	    campaignName: string;
	    sessionId?: number;
	    sessionLabel: string;
	    fear: number;
	    active: boolean;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CombatSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.encounterId = source["encounterId"];
	        this.encounterName = source["encounterName"];
	        this.campaignId = source["campaignId"];
	        this.campaignName = source["campaignName"];
	        this.sessionId = source["sessionId"];
	        this.sessionLabel = source["sessionLabel"];
	        this.fear = source["fear"];
	        this.active = source["active"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class CombatantView {
	    id: number;
	    combatId: number;
	    displayName: string;
	    adversarySlug?: string;
	    hpMax: number;
	    hpMarked: number;
	    stressMax: number;
	    stressMarked: number;
	    spotlight: boolean;
	    adversary?: cards.Adversary;
	    source: string;
	    unresolved?: boolean;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CombatantView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.combatId = source["combatId"];
	        this.displayName = source["displayName"];
	        this.adversarySlug = source["adversarySlug"];
	        this.hpMax = source["hpMax"];
	        this.hpMarked = source["hpMarked"];
	        this.stressMax = source["stressMax"];
	        this.stressMarked = source["stressMarked"];
	        this.spotlight = source["spotlight"];
	        this.adversary = this.convertValues(source["adversary"], cards.Adversary);
	        this.source = source["source"];
	        this.unresolved = source["unresolved"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CombatView {
	    id: number;
	    encounterId?: number;
	    encounterName: string;
	    campaignId?: number;
	    campaignName: string;
	    sessionId?: number;
	    sessionLabel: string;
	    fear: number;
	    fearMax: number;
	    active: boolean;
	    combatants: CombatantView[];
	    environment?: cards.Environment;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new CombatView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.encounterId = source["encounterId"];
	        this.encounterName = source["encounterName"];
	        this.campaignId = source["campaignId"];
	        this.campaignName = source["campaignName"];
	        this.sessionId = source["sessionId"];
	        this.sessionLabel = source["sessionLabel"];
	        this.fear = source["fear"];
	        this.fearMax = source["fearMax"];
	        this.active = source["active"];
	        this.combatants = this.convertValues(source["combatants"], CombatantView);
	        this.environment = this.convertValues(source["environment"], cards.Environment);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CombatantInput {
	    id?: number;
	    combatId: number;
	    displayName: string;
	    adversarySlug?: string;
	    hpMax: number;
	    stressMax: number;
	
	    static createFrom(source: any = {}) {
	        return new CombatantInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.combatId = source["combatId"];
	        this.displayName = source["displayName"];
	        this.adversarySlug = source["adversarySlug"];
	        this.hpMax = source["hpMax"];
	        this.stressMax = source["stressMax"];
	    }
	}
	
	export class Countdown {
	    id: number;
	    campaignId?: number;
	    name: string;
	    value: number;
	    max: number;
	    kind: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Countdown(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.name = source["name"];
	        this.value = source["value"];
	        this.max = source["max"];
	        this.kind = source["kind"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class CountdownInput {
	    id?: number;
	    campaignId?: number;
	    name: string;
	    value: number;
	    max: number;
	    kind: string;
	
	    static createFrom(source: any = {}) {
	        return new CountdownInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.name = source["name"];
	        this.value = source["value"];
	        this.max = source["max"];
	        this.kind = source["kind"];
	    }
	}
	export class Pick {
	    slug: string;
	    count: number;
	
	    static createFrom(source: any = {}) {
	        return new Pick(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.slug = source["slug"];
	        this.count = source["count"];
	    }
	}
	export class EncounterInput {
	    id?: number;
	    name: string;
	    partyId?: number;
	    environmentSlug?: string;
	    adversaries: Pick[];
	    customAdversaries: Pick[];
	
	    static createFrom(source: any = {}) {
	        return new EncounterInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.partyId = source["partyId"];
	        this.environmentSlug = source["environmentSlug"];
	        this.adversaries = this.convertValues(source["adversaries"], Pick);
	        this.customAdversaries = this.convertValues(source["customAdversaries"], Pick);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EncounterSummary {
	    id: number;
	    name: string;
	    partyId?: number;
	    environmentSlug?: string;
	    totalCount: number;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new EncounterSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.partyId = source["partyId"];
	        this.environmentSlug = source["environmentSlug"];
	        this.totalCount = source["totalCount"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class Filter {
	    tier: string;
	    type: string;
	
	    static createFrom(source: any = {}) {
	        return new Filter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.tier = source["tier"];
	        this.type = source["type"];
	    }
	}
	export class ImportReport {
	    parties: number;
	    customAdversaries: number;
	    customEnvironments: number;
	    encounters: number;
	    campaigns: number;
	    sessions: number;
	    notes: number;
	    countdowns: number;
	    renamed: string[];
	    skipped: string[];
	
	    static createFrom(source: any = {}) {
	        return new ImportReport(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parties = source["parties"];
	        this.customAdversaries = source["customAdversaries"];
	        this.customEnvironments = source["customEnvironments"];
	        this.encounters = source["encounters"];
	        this.campaigns = source["campaigns"];
	        this.sessions = source["sessions"];
	        this.notes = source["notes"];
	        this.countdowns = source["countdowns"];
	        this.renamed = source["renamed"];
	        this.skipped = source["skipped"];
	    }
	}
	export class Note {
	    id: number;
	    campaignId: number;
	    kind: string;
	    title: string;
	    body: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Note(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.body = source["body"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class NoteInput {
	    id?: number;
	    campaignId: number;
	    kind: string;
	    title: string;
	    body: string;
	
	    static createFrom(source: any = {}) {
	        return new NoteInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.kind = source["kind"];
	        this.title = source["title"];
	        this.body = source["body"];
	    }
	}
	export class Party {
	    id: number;
	    name: string;
	    size: number;
	    tier: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Party(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.size = source["size"];
	        this.tier = source["tier"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class PartyInput {
	    name: string;
	    size: number;
	    tier: string;
	
	    static createFrom(source: any = {}) {
	        return new PartyInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.size = source["size"];
	        this.tier = source["tier"];
	    }
	}
	
	export class SearchHit {
	    entity: string;
	    entityId: number;
	    campaignId: number;
	    slug: string;
	    title: string;
	    excerpt: string;
	    score: number;
	
	    static createFrom(source: any = {}) {
	        return new SearchHit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.entity = source["entity"];
	        this.entityId = source["entityId"];
	        this.campaignId = source["campaignId"];
	        this.slug = source["slug"];
	        this.title = source["title"];
	        this.excerpt = source["excerpt"];
	        this.score = source["score"];
	    }
	}
	export class SessionInput {
	    id?: number;
	    campaignId: number;
	    number: number;
	    title: string;
	    date: string;
	    recap: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.number = source["number"];
	        this.title = source["title"];
	        this.date = source["date"];
	        this.recap = source["recap"];
	    }
	}
	export class SessionSummary {
	    id: number;
	    campaignId: number;
	    number: number;
	    title: string;
	    date: string;
	    recap: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.number = source["number"];
	        this.title = source["title"];
	        this.date = source["date"];
	        this.recap = source["recap"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class SessionView {
	    id: number;
	    campaignId: number;
	    number: number;
	    title: string;
	    date: string;
	    recap: string;
	    createdAt: string;
	    updatedAt: string;
	    encounters: EncounterSummary[];
	    combats: CombatSummary[];
	
	    static createFrom(source: any = {}) {
	        return new SessionView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.campaignId = source["campaignId"];
	        this.number = source["number"];
	        this.title = source["title"];
	        this.date = source["date"];
	        this.recap = source["recap"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.encounters = this.convertValues(source["encounters"], EncounterSummary);
	        this.combats = this.convertValues(source["combats"], CombatSummary);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SharePreview {
	    kind: string;
	    name: string;
	    slug: string;
	    tier: string;
	    type: string;
	    description: string;
	    renamed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SharePreview(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.name = source["name"];
	        this.slug = source["slug"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.renamed = source["renamed"];
	    }
	}

}

export namespace main {
	
	export class WindowSize {
	    label: string;
	    width: number;
	    height: number;
	    fits: boolean;
	
	    static createFrom(source: any = {}) {
	        return new WindowSize(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.width = source["width"];
	        this.height = source["height"];
	        this.fits = source["fits"];
	    }
	}

}

export namespace player {
	
	export class AdvancementChoice {
	    id: string;
	    traits: string[];
	    experienceIds: number[];
	    domainCardSlug: string;
	    multiclassSlug: string;
	    multiclassSubclassSlug: string;
	
	    static createFrom(source: any = {}) {
	        return new AdvancementChoice(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.traits = source["traits"];
	        this.experienceIds = source["experienceIds"];
	        this.domainCardSlug = source["domainCardSlug"];
	        this.multiclassSlug = source["multiclassSlug"];
	        this.multiclassSubclassSlug = source["multiclassSubclassSlug"];
	    }
	}
	export class Gold {
	    handfuls: number;
	    bags: number;
	    chests: number;
	
	    static createFrom(source: any = {}) {
	        return new Gold(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.handfuls = source["handfuls"];
	        this.bags = source["bags"];
	        this.chests = source["chests"];
	    }
	}
	export class Character {
	    id: number;
	    name: string;
	    pronouns: string;
	    classSlug: string;
	    className: string;
	    subclassSlug: string;
	    subclassName: string;
	    subclassMastery: number;
	    multiclassSlug: string;
	    multiclassName: string;
	    multiclassSubclassSlug: string;
	    multiclassSubclassName: string;
	    ancestrySlug: string;
	    ancestryName: string;
	    communitySlug: string;
	    communityName: string;
	    domains: string[];
	    level: number;
	    tier: number;
	    proficiency: number;
	    traits: Record<string, number>;
	    markedTraits: string[];
	    spellcastTrait: string;
	    beastformSlug: string;
	    beastformName: string;
	    hpMax: number;
	    hpMarked: number;
	    stressMax: number;
	    stressMarked: number;
	    hope: number;
	    hopeMax: number;
	    evasion: number;
	    armorScore: number;
	    armorMarked: number;
	    thresholdMajor: number;
	    thresholdSevere: number;
	    gold: Gold;
	    background: string;
	    connections: string;
	    notes: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Character(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.pronouns = source["pronouns"];
	        this.classSlug = source["classSlug"];
	        this.className = source["className"];
	        this.subclassSlug = source["subclassSlug"];
	        this.subclassName = source["subclassName"];
	        this.subclassMastery = source["subclassMastery"];
	        this.multiclassSlug = source["multiclassSlug"];
	        this.multiclassName = source["multiclassName"];
	        this.multiclassSubclassSlug = source["multiclassSubclassSlug"];
	        this.multiclassSubclassName = source["multiclassSubclassName"];
	        this.ancestrySlug = source["ancestrySlug"];
	        this.ancestryName = source["ancestryName"];
	        this.communitySlug = source["communitySlug"];
	        this.communityName = source["communityName"];
	        this.domains = source["domains"];
	        this.level = source["level"];
	        this.tier = source["tier"];
	        this.proficiency = source["proficiency"];
	        this.traits = source["traits"];
	        this.markedTraits = source["markedTraits"];
	        this.spellcastTrait = source["spellcastTrait"];
	        this.beastformSlug = source["beastformSlug"];
	        this.beastformName = source["beastformName"];
	        this.hpMax = source["hpMax"];
	        this.hpMarked = source["hpMarked"];
	        this.stressMax = source["stressMax"];
	        this.stressMarked = source["stressMarked"];
	        this.hope = source["hope"];
	        this.hopeMax = source["hopeMax"];
	        this.evasion = source["evasion"];
	        this.armorScore = source["armorScore"];
	        this.armorMarked = source["armorMarked"];
	        this.thresholdMajor = source["thresholdMajor"];
	        this.thresholdSevere = source["thresholdSevere"];
	        this.gold = this.convertValues(source["gold"], Gold);
	        this.background = source["background"];
	        this.connections = source["connections"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class BeastformView {
	    eligible: boolean;
	    tier: number;
	    available: cards.Beastform[];
	    active?: cards.Beastform;
	    baseEvasion: number;
	    evasion: number;
	    traitBonus: string;
	    character: Character;
	
	    static createFrom(source: any = {}) {
	        return new BeastformView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eligible = source["eligible"];
	        this.tier = source["tier"];
	        this.available = this.convertValues(source["available"], cards.Beastform);
	        this.active = this.convertValues(source["active"], cards.Beastform);
	        this.baseEvasion = source["baseEvasion"];
	        this.evasion = source["evasion"];
	        this.traitBonus = source["traitBonus"];
	        this.character = this.convertValues(source["character"], Character);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CharacterInput {
	    id?: number;
	    name: string;
	    pronouns: string;
	    classSlug: string;
	    subclassSlug: string;
	    ancestrySlug: string;
	    communitySlug: string;
	    traits: Record<string, number>;
	    hpMax: number;
	    stressMax: number;
	    evasion: number;
	    armorScore: number;
	    thresholdMajor: number;
	    thresholdSevere: number;
	    background: string;
	    connections: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CharacterInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.pronouns = source["pronouns"];
	        this.classSlug = source["classSlug"];
	        this.subclassSlug = source["subclassSlug"];
	        this.ancestrySlug = source["ancestrySlug"];
	        this.communitySlug = source["communitySlug"];
	        this.traits = source["traits"];
	        this.hpMax = source["hpMax"];
	        this.stressMax = source["stressMax"];
	        this.evasion = source["evasion"];
	        this.armorScore = source["armorScore"];
	        this.thresholdMajor = source["thresholdMajor"];
	        this.thresholdSevere = source["thresholdSevere"];
	        this.background = source["background"];
	        this.connections = source["connections"];
	        this.notes = source["notes"];
	    }
	}
	export class CompanionExperience {
	    name: string;
	    modifier: number;
	
	    static createFrom(source: any = {}) {
	        return new CompanionExperience(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.modifier = source["modifier"];
	    }
	}
	export class Companion {
	    id: number;
	    characterId: number;
	    name: string;
	    evasion: number;
	    damageDie: string;
	    attackRange: string;
	    attack: string;
	    stressMax: number;
	    stressMarked: number;
	    experiences: CompanionExperience[];
	    upgrades: string[];
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new Companion(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.characterId = source["characterId"];
	        this.name = source["name"];
	        this.evasion = source["evasion"];
	        this.damageDie = source["damageDie"];
	        this.attackRange = source["attackRange"];
	        this.attack = source["attack"];
	        this.stressMax = source["stressMax"];
	        this.stressMarked = source["stressMarked"];
	        this.experiences = this.convertValues(source["experiences"], CompanionExperience);
	        this.upgrades = source["upgrades"];
	        this.notes = source["notes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	export class CompanionInput {
	    characterId: number;
	    name: string;
	    evasion: number;
	    damageDie: string;
	    attackRange: string;
	    attack: string;
	    stressMax: number;
	    experiences: CompanionExperience[];
	    upgrades: string[];
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new CompanionInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.name = source["name"];
	        this.evasion = source["evasion"];
	        this.damageDie = source["damageDie"];
	        this.attackRange = source["attackRange"];
	        this.attack = source["attack"];
	        this.stressMax = source["stressMax"];
	        this.experiences = this.convertValues(source["experiences"], CompanionExperience);
	        this.upgrades = source["upgrades"];
	        this.notes = source["notes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class CompanionView {
	    eligible: boolean;
	    companion?: Companion;
	    reference?: cards.Companion;
	    proficiency: number;
	    damageDice: string[];
	    ranges: string[];
	
	    static createFrom(source: any = {}) {
	        return new CompanionView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.eligible = source["eligible"];
	        this.companion = this.convertValues(source["companion"], Companion);
	        this.reference = this.convertValues(source["reference"], cards.Companion);
	        this.proficiency = source["proficiency"];
	        this.damageDice = source["damageDice"];
	        this.ranges = source["ranges"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DamageResult {
	    label: string;
	    count: number;
	    sides: number;
	    modifier: number;
	    total: number;
	    critical: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DamageResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.count = source["count"];
	        this.sides = source["sides"];
	        this.modifier = source["modifier"];
	        this.total = source["total"];
	        this.critical = source["critical"];
	    }
	}
	export class DomainCard {
	    id: number;
	    cardSlug: string;
	    location: string;
	    unresolved: boolean;
	    card: cards.DomainCard;
	
	    static createFrom(source: any = {}) {
	        return new DomainCard(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.cardSlug = source["cardSlug"];
	        this.location = source["location"];
	        this.unresolved = source["unresolved"];
	        this.card = this.convertValues(source["card"], cards.DomainCard);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class DomainFilter {
	    domain: string;
	    maxLevel: number;
	    level: string;
	
	    static createFrom(source: any = {}) {
	        return new DomainFilter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.domain = source["domain"];
	        this.maxLevel = source["maxLevel"];
	        this.level = source["level"];
	    }
	}
	export class Effect {
	    traits: number;
	    evasion: number;
	    hpSlots: number;
	    stressSlots: number;
	    experiences: number;
	    experienceBonus: number;
	    domainCards: number;
	    proficiency: number;
	    subclassUpgrade: number;
	    multiclass: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Effect(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.traits = source["traits"];
	        this.evasion = source["evasion"];
	        this.hpSlots = source["hpSlots"];
	        this.stressSlots = source["stressSlots"];
	        this.experiences = source["experiences"];
	        this.experienceBonus = source["experienceBonus"];
	        this.domainCards = source["domainCards"];
	        this.proficiency = source["proficiency"];
	        this.subclassUpgrade = source["subclassUpgrade"];
	        this.multiclass = source["multiclass"];
	    }
	}
	export class Experience {
	    id: number;
	    name: string;
	    modifier: number;
	
	    static createFrom(source: any = {}) {
	        return new Experience(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.modifier = source["modifier"];
	    }
	}
	export class ExperienceInput {
	    id?: number;
	    characterId: number;
	    name: string;
	    modifier: number;
	
	    static createFrom(source: any = {}) {
	        return new ExperienceInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.characterId = source["characterId"];
	        this.name = source["name"];
	        this.modifier = source["modifier"];
	    }
	}
	
	export class Item {
	    id: number;
	    name: string;
	    kind: string;
	    qty: number;
	    equipped: boolean;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new Item(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.qty = source["qty"];
	        this.equipped = source["equipped"];
	        this.detail = source["detail"];
	    }
	}
	export class ItemInput {
	    id?: number;
	    characterId: number;
	    name: string;
	    kind: string;
	    qty: number;
	    equipped: boolean;
	    detail: string;
	
	    static createFrom(source: any = {}) {
	        return new ItemInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.characterId = source["characterId"];
	        this.name = source["name"];
	        this.kind = source["kind"];
	        this.qty = source["qty"];
	        this.equipped = source["equipped"];
	        this.detail = source["detail"];
	    }
	}
	export class LevelUpInput {
	    characterId: number;
	    choices: AdvancementChoice[];
	    newExperienceName: string;
	    domainCardSlug: string;
	
	    static createFrom(source: any = {}) {
	        return new LevelUpInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.choices = this.convertValues(source["choices"], AdvancementChoice);
	        this.newExperienceName = source["newExperienceName"];
	        this.domainCardSlug = source["domainCardSlug"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class PlanOption {
	    id: string;
	    label: string;
	    description: string;
	    slots: number;
	    cost: number;
	    effect: Effect;
	    used: number;
	    remaining: number;
	    available: boolean;
	    reason: string;
	
	    static createFrom(source: any = {}) {
	        return new PlanOption(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.description = source["description"];
	        this.slots = source["slots"];
	        this.cost = source["cost"];
	        this.effect = this.convertValues(source["effect"], Effect);
	        this.used = source["used"];
	        this.remaining = source["remaining"];
	        this.available = source["available"];
	        this.reason = source["reason"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LevelUpPlan {
	    characterId: number;
	    fromLevel: number;
	    toLevel: number;
	    fromTier: number;
	    tier: number;
	    tierName: string;
	    newTier: boolean;
	    achievements: string[];
	    needsNewExperience: boolean;
	    proficiency: number;
	    proficiencyBonus: number;
	    maxProficiency: number;
	    clearsMarkedTraits: boolean;
	    advancementsRequired: number;
	    thresholdIncrease: number;
	    domainCardsPerLevel: number;
	    options: PlanOption[];
	    availableCards: cards.DomainCard[];
	    experiences: Experience[];
	    traits: Record<string, number>;
	    markedTraits: string[];
	    classes: cards.CharacterClass[];
	
	    static createFrom(source: any = {}) {
	        return new LevelUpPlan(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.fromLevel = source["fromLevel"];
	        this.toLevel = source["toLevel"];
	        this.fromTier = source["fromTier"];
	        this.tier = source["tier"];
	        this.tierName = source["tierName"];
	        this.newTier = source["newTier"];
	        this.achievements = source["achievements"];
	        this.needsNewExperience = source["needsNewExperience"];
	        this.proficiency = source["proficiency"];
	        this.proficiencyBonus = source["proficiencyBonus"];
	        this.maxProficiency = source["maxProficiency"];
	        this.clearsMarkedTraits = source["clearsMarkedTraits"];
	        this.advancementsRequired = source["advancementsRequired"];
	        this.thresholdIncrease = source["thresholdIncrease"];
	        this.domainCardsPerLevel = source["domainCardsPerLevel"];
	        this.options = this.convertValues(source["options"], PlanOption);
	        this.availableCards = this.convertValues(source["availableCards"], cards.DomainCard);
	        this.experiences = this.convertValues(source["experiences"], Experience);
	        this.traits = source["traits"];
	        this.markedTraits = source["markedTraits"];
	        this.classes = this.convertValues(source["classes"], cards.CharacterClass);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class LevelUpRecord {
	    level: number;
	    tier: number;
	    choices: string[];
	    summary: string;
	    createdAt: string;
	
	    static createFrom(source: any = {}) {
	        return new LevelUpRecord(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.tier = source["tier"];
	        this.choices = source["choices"];
	        this.summary = source["summary"];
	        this.createdAt = source["createdAt"];
	    }
	}
	export class Loadout {
	    loadout: DomainCard[];
	    vault: DomainCard[];
	    loadoutMax: number;
	    held: number;
	    allowance: number;
	
	    static createFrom(source: any = {}) {
	        return new Loadout(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loadout = this.convertValues(source["loadout"], DomainCard);
	        this.vault = this.convertValues(source["vault"], DomainCard);
	        this.loadoutMax = source["loadoutMax"];
	        this.held = source["held"];
	        this.allowance = source["allowance"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class ModifierPart {
	    label: string;
	    value: number;
	
	    static createFrom(source: any = {}) {
	        return new ModifierPart(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.value = source["value"];
	    }
	}
	
	export class RestInput {
	    characterId: number;
	    long: boolean;
	    moves: string[];
	
	    static createFrom(source: any = {}) {
	        return new RestInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.long = source["long"];
	        this.moves = source["moves"];
	    }
	}
	export class RestMove {
	    id: string;
	    label: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new RestMove(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.label = source["label"];
	        this.description = source["description"];
	    }
	}
	export class RestResult {
	    character: Character;
	    long: boolean;
	    allowed: number;
	    outcomes: string[];
	
	    static createFrom(source: any = {}) {
	        return new RestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.character = this.convertValues(source["character"], Character);
	        this.long = source["long"];
	        this.allowed = source["allowed"];
	        this.outcomes = source["outcomes"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class Roll {
	    label: string;
	    hope: number;
	    fear: number;
	    result: number;
	    msg: string;
	    critical: boolean;
	    withHope: boolean;
	    withFear: boolean;
	    modifier: number;
	    modifierParts: ModifierPart[];
	    hopeSpent: number;
	    hopeGained: number;
	    stressCleared: number;
	    advantage: boolean;
	    disadvantage: boolean;
	    character: Character;
	
	    static createFrom(source: any = {}) {
	        return new Roll(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.label = source["label"];
	        this.hope = source["hope"];
	        this.fear = source["fear"];
	        this.result = source["result"];
	        this.msg = source["msg"];
	        this.critical = source["critical"];
	        this.withHope = source["withHope"];
	        this.withFear = source["withFear"];
	        this.modifier = source["modifier"];
	        this.modifierParts = this.convertValues(source["modifierParts"], ModifierPart);
	        this.hopeSpent = source["hopeSpent"];
	        this.hopeGained = source["hopeGained"];
	        this.stressCleared = source["stressCleared"];
	        this.advantage = source["advantage"];
	        this.disadvantage = source["disadvantage"];
	        this.character = this.convertValues(source["character"], Character);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class RollRequest {
	    characterId: number;
	    trait: string;
	    experienceIds: number[];
	    bonus: number;
	    advantage: boolean;
	    disadvantage: boolean;
	    spendHope: number;
	    label: string;
	
	    static createFrom(source: any = {}) {
	        return new RollRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.trait = source["trait"];
	        this.experienceIds = source["experienceIds"];
	        this.bonus = source["bonus"];
	        this.advantage = source["advantage"];
	        this.disadvantage = source["disadvantage"];
	        this.spendHope = source["spendHope"];
	        this.label = source["label"];
	    }
	}
	export class Sheet {
	    id: number;
	    name: string;
	    pronouns: string;
	    classSlug: string;
	    className: string;
	    subclassSlug: string;
	    subclassName: string;
	    subclassMastery: number;
	    multiclassSlug: string;
	    multiclassName: string;
	    multiclassSubclassSlug: string;
	    multiclassSubclassName: string;
	    ancestrySlug: string;
	    ancestryName: string;
	    communitySlug: string;
	    communityName: string;
	    domains: string[];
	    level: number;
	    tier: number;
	    proficiency: number;
	    traits: Record<string, number>;
	    markedTraits: string[];
	    spellcastTrait: string;
	    beastformSlug: string;
	    beastformName: string;
	    hpMax: number;
	    hpMarked: number;
	    stressMax: number;
	    stressMarked: number;
	    hope: number;
	    hopeMax: number;
	    evasion: number;
	    armorScore: number;
	    armorMarked: number;
	    thresholdMajor: number;
	    thresholdSevere: number;
	    gold: Gold;
	    background: string;
	    connections: string;
	    notes: string;
	    createdAt: string;
	    updatedAt: string;
	    class?: cards.CharacterClass;
	    subclass?: cards.Subclass;
	    multiclass?: cards.CharacterClass;
	    multiclassSubclass?: cards.Subclass;
	    ancestry?: cards.Ancestry;
	    community?: cards.Community;
	    loadout: DomainCard[];
	    vault: DomainCard[];
	    inventory: Item[];
	    experiences: Experience[];
	    loadoutMax: number;
	    levelUps: LevelUpRecord[];
	    canLevelUp: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Sheet(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.pronouns = source["pronouns"];
	        this.classSlug = source["classSlug"];
	        this.className = source["className"];
	        this.subclassSlug = source["subclassSlug"];
	        this.subclassName = source["subclassName"];
	        this.subclassMastery = source["subclassMastery"];
	        this.multiclassSlug = source["multiclassSlug"];
	        this.multiclassName = source["multiclassName"];
	        this.multiclassSubclassSlug = source["multiclassSubclassSlug"];
	        this.multiclassSubclassName = source["multiclassSubclassName"];
	        this.ancestrySlug = source["ancestrySlug"];
	        this.ancestryName = source["ancestryName"];
	        this.communitySlug = source["communitySlug"];
	        this.communityName = source["communityName"];
	        this.domains = source["domains"];
	        this.level = source["level"];
	        this.tier = source["tier"];
	        this.proficiency = source["proficiency"];
	        this.traits = source["traits"];
	        this.markedTraits = source["markedTraits"];
	        this.spellcastTrait = source["spellcastTrait"];
	        this.beastformSlug = source["beastformSlug"];
	        this.beastformName = source["beastformName"];
	        this.hpMax = source["hpMax"];
	        this.hpMarked = source["hpMarked"];
	        this.stressMax = source["stressMax"];
	        this.stressMarked = source["stressMarked"];
	        this.hope = source["hope"];
	        this.hopeMax = source["hopeMax"];
	        this.evasion = source["evasion"];
	        this.armorScore = source["armorScore"];
	        this.armorMarked = source["armorMarked"];
	        this.thresholdMajor = source["thresholdMajor"];
	        this.thresholdSevere = source["thresholdSevere"];
	        this.gold = this.convertValues(source["gold"], Gold);
	        this.background = source["background"];
	        this.connections = source["connections"];
	        this.notes = source["notes"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	        this.class = this.convertValues(source["class"], cards.CharacterClass);
	        this.subclass = this.convertValues(source["subclass"], cards.Subclass);
	        this.multiclass = this.convertValues(source["multiclass"], cards.CharacterClass);
	        this.multiclassSubclass = this.convertValues(source["multiclassSubclass"], cards.Subclass);
	        this.ancestry = this.convertValues(source["ancestry"], cards.Ancestry);
	        this.community = this.convertValues(source["community"], cards.Community);
	        this.loadout = this.convertValues(source["loadout"], DomainCard);
	        this.vault = this.convertValues(source["vault"], DomainCard);
	        this.inventory = this.convertValues(source["inventory"], Item);
	        this.experiences = this.convertValues(source["experiences"], Experience);
	        this.loadoutMax = source["loadoutMax"];
	        this.levelUps = this.convertValues(source["levelUps"], LevelUpRecord);
	        this.canLevelUp = source["canLevelUp"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class SwapInput {
	    characterId: number;
	    recall: string;
	    vault: string;
	    resting: boolean;
	
	    static createFrom(source: any = {}) {
	        return new SwapInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.characterId = source["characterId"];
	        this.recall = source["recall"];
	        this.vault = source["vault"];
	        this.resting = source["resting"];
	    }
	}
	export class SwapResult {
	    character: Character;
	    loadout: Loadout;
	    recallCost: number;
	    stressMarked: number;
	    outcome: string;
	
	    static createFrom(source: any = {}) {
	        return new SwapResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.character = this.convertValues(source["character"], Character);
	        this.loadout = this.convertValues(source["loadout"], Loadout);
	        this.recallCost = source["recallCost"];
	        this.stressMarked = source["stressMarked"];
	        this.outcome = source["outcome"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace rules {
	
	export class BudgetSummary {
	    partySize: number;
	    budget: number;
	    spent: number;
	    remaining: number;
	    over: boolean;
	    adjustments: string[];
	
	    static createFrom(source: any = {}) {
	        return new BudgetSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.partySize = source["partySize"];
	        this.budget = source["budget"];
	        this.spent = source["spent"];
	        this.remaining = source["remaining"];
	        this.over = source["over"];
	        this.adjustments = source["adjustments"];
	    }
	}
	export class EncounterAdversary {
	    kind: string;
	    slug: string;
	    name: string;
	    tier: string;
	    type: string;
	    description: string;
	    hordeNumber: string;
	    motives: string;
	    experiences: string;
	    difficulty: string;
	    thresholdMinor: string;
	    thresholdMajor: string;
	    hp: string;
	    stress: string;
	    standardAttack: cards.Attack;
	    features: cards.Feature[];
	    count: number;
	    source: string;
	    id: number;
	    unresolved?: boolean;
	
	    static createFrom(source: any = {}) {
	        return new EncounterAdversary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.slug = source["slug"];
	        this.name = source["name"];
	        this.tier = source["tier"];
	        this.type = source["type"];
	        this.description = source["description"];
	        this.hordeNumber = source["hordeNumber"];
	        this.motives = source["motives"];
	        this.experiences = source["experiences"];
	        this.difficulty = source["difficulty"];
	        this.thresholdMinor = source["thresholdMinor"];
	        this.thresholdMajor = source["thresholdMajor"];
	        this.hp = source["hp"];
	        this.stress = source["stress"];
	        this.standardAttack = this.convertValues(source["standardAttack"], cards.Attack);
	        this.features = this.convertValues(source["features"], cards.Feature);
	        this.count = source["count"];
	        this.source = source["source"];
	        this.id = source["id"];
	        this.unresolved = source["unresolved"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class EncounterSettings {
	    partySize: number;
	    partyTier: string;
	    difficulty: string;
	
	    static createFrom(source: any = {}) {
	        return new EncounterSettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.partySize = source["partySize"];
	        this.partyTier = source["partyTier"];
	        this.difficulty = source["difficulty"];
	    }
	}
	export class EncounterView {
	    id: number;
	    name: string;
	    partyId?: number;
	    adversaries: EncounterAdversary[];
	    totalCount: number;
	    environmentSlug?: string;
	    environment?: cards.Environment;
	    budget?: BudgetSummary;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new EncounterView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.partyId = source["partyId"];
	        this.adversaries = this.convertValues(source["adversaries"], EncounterAdversary);
	        this.totalCount = source["totalCount"];
	        this.environmentSlug = source["environmentSlug"];
	        this.environment = this.convertValues(source["environment"], cards.Environment);
	        this.budget = this.convertValues(source["budget"], BudgetSummary);
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}

}

export namespace update {
	
	export class Release {
	    current: string;
	    latest: string;
	    url: string;
	    notes: string;
	    newer: boolean;
	    known: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Release(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.url = source["url"];
	        this.notes = source["notes"];
	        this.newer = source["newer"];
	        this.known = source["known"];
	    }
	}

}

