export namespace main {
	
	export class ChannelMessage {
	    channel: string;
	    chat_id: string;
	    sender_id: string;
	    content: string;
	    role: string;
	    time: string;
	
	    static createFrom(source: any = {}) {
	        return new ChannelMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.channel = source["channel"];
	        this.chat_id = source["chat_id"];
	        this.sender_id = source["sender_id"];
	        this.content = source["content"];
	        this.role = source["role"];
	        this.time = source["time"];
	    }
	}
	export class ChatMessage {
	    id: string;
	    role: string;
	    content: string;
	    // Go type: time
	    timestamp: any;
	    session: string;
	    channel: string;
	    streaming: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ChatMessage(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.role = source["role"];
	        this.content = source["content"];
	        this.timestamp = this.convertValues(source["timestamp"], null);
	        this.session = source["session"];
	        this.channel = source["channel"];
	        this.streaming = source["streaming"];
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
	export class GatewayStatus {
	    running: boolean;
	    pid: number;
	    port: number;
	    // Go type: time
	    startedAt: any;
	    configPath: string;
	    botId: string;
	    uptime: string;
	
	    static createFrom(source: any = {}) {
	        return new GatewayStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.running = source["running"];
	        this.pid = source["pid"];
	        this.port = source["port"];
	        this.startedAt = this.convertValues(source["startedAt"], null);
	        this.configPath = source["configPath"];
	        this.botId = source["botId"];
	        this.uptime = source["uptime"];
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
	export class SessionInfo {
	    key: string;
	    created_at: string;
	    updated_at: string;
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new SessionInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.created_at = source["created_at"];
	        this.updated_at = source["updated_at"];
	        this.path = source["path"];
	    }
	}

}

