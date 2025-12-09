export namespace logging {
	
	export class Event {
	    // Go type: time
	    Timestamp: any;
	    Level: string;
	    Category: string;
	    Message: string;
	    Details: Record<string, any>;
	
	    static createFrom(source: any = {}) {
	        return new Event(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Timestamp = this.convertValues(source["Timestamp"], null);
	        this.Level = source["Level"];
	        this.Category = source["Category"];
	        this.Message = source["Message"];
	        this.Details = source["Details"];
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

export namespace profiles {
	
	export class Profile {
	    Name: string;
	    Description: string;
	    Active: boolean;
	    Rules: string[];
	
	    static createFrom(source: any = {}) {
	        return new Profile(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Description = source["Description"];
	        this.Active = source["Active"];
	        this.Rules = source["Rules"];
	    }
	}

}

export namespace rules {
	
	export class Rule {
	    Name: string;
	    Application: string;
	    Action: string;
	    Protocol: string;
	    Ports: number[];
	    Direction: string;
	
	    static createFrom(source: any = {}) {
	        return new Rule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Name = source["Name"];
	        this.Application = source["Application"];
	        this.Action = source["Action"];
	        this.Protocol = source["Protocol"];
	        this.Ports = source["Ports"];
	        this.Direction = source["Direction"];
	    }
	}

}

export namespace stats {
	
	export class ConnectionStat {
	    // Go type: time
	    Timestamp: any;
	    Application: string;
	    Protocol: string;
	    Direction: string;
	    BytesSent: number;
	    BytesRecv: number;
	    Action: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionStat(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Timestamp = this.convertValues(source["Timestamp"], null);
	        this.Application = source["Application"];
	        this.Protocol = source["Protocol"];
	        this.Direction = source["Direction"];
	        this.BytesSent = source["BytesSent"];
	        this.BytesRecv = source["BytesRecv"];
	        this.Action = source["Action"];
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
	export class Filter {
	    Application: string;
	    Protocol: string;
	    Direction: string;
	    Action: string;
	    // Go type: time
	    Since: any;
	    // Go type: time
	    Until: any;
	
	    static createFrom(source: any = {}) {
	        return new Filter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Application = source["Application"];
	        this.Protocol = source["Protocol"];
	        this.Direction = source["Direction"];
	        this.Action = source["Action"];
	        this.Since = this.convertValues(source["Since"], null);
	        this.Until = this.convertValues(source["Until"], null);
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

