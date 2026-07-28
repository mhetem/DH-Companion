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

