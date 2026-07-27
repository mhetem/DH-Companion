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

