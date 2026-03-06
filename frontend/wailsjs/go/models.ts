export namespace types {
	
	export class Domain {
	    id: number;
	    domain: string;
	    provider: string;
	    token: string;
	    accessKeyID: string;
	    accessKeySecret: string;
	    interval: number;
	    enabled: boolean;
	    currentIP: string;
	    lastUpdate: string;
	    createdAt: string;
	    updatedAt: string;
	
	    static createFrom(source: any = {}) {
	        return new Domain(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.domain = source["domain"];
	        this.provider = source["provider"];
	        this.token = source["token"];
	        this.accessKeyID = source["accessKeyID"];
	        this.accessKeySecret = source["accessKeySecret"];
	        this.interval = source["interval"];
	        this.enabled = source["enabled"];
	        this.currentIP = source["currentIP"];
	        this.lastUpdate = source["lastUpdate"];
	        this.createdAt = source["createdAt"];
	        this.updatedAt = source["updatedAt"];
	    }
	}
	export class DomainStatus {
	    id: number;
	    domain: string;
	    currentIP: string;
	    lastUpdate: string;
	    status: string;
	    message: string;
	    provider: string;
	    lastApiCall: string;
	    apiStatus: string;
	    apiMessage: string;
	
	    static createFrom(source: any = {}) {
	        return new DomainStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.domain = source["domain"];
	        this.currentIP = source["currentIP"];
	        this.lastUpdate = source["lastUpdate"];
	        this.status = source["status"];
	        this.message = source["message"];
	        this.provider = source["provider"];
	        this.lastApiCall = source["lastApiCall"];
	        this.apiStatus = source["apiStatus"];
	        this.apiMessage = source["apiMessage"];
	    }
	}
	export class IPInfo {
	    publicIP: string;
	    ipv4: string;
	    ipv6: string;
	    location: string;
	    isp: string;
	
	    static createFrom(source: any = {}) {
	        return new IPInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.publicIP = source["publicIP"];
	        this.ipv4 = source["ipv4"];
	        this.ipv6 = source["ipv6"];
	        this.location = source["location"];
	        this.isp = source["isp"];
	    }
	}
	export class LogEntry {
	    id: number;
	    timestamp: string;
	    level: string;
	    domain: string;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new LogEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.timestamp = source["timestamp"];
	        this.level = source["level"];
	        this.domain = source["domain"];
	        this.message = source["message"];
	    }
	}
	export class Settings {
	    autoStart: boolean;
	    checkInterval: number;
	    retryCount: number;
	    retryDelay: number;
	    logLevel: string;
	    timezone: string;
	    notifySuccess: boolean;
	    notifyError: boolean;
	    proxy: string;
	    apiEndpoint: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.autoStart = source["autoStart"];
	        this.checkInterval = source["checkInterval"];
	        this.retryCount = source["retryCount"];
	        this.retryDelay = source["retryDelay"];
	        this.logLevel = source["logLevel"];
	        this.timezone = source["timezone"];
	        this.notifySuccess = source["notifySuccess"];
	        this.notifyError = source["notifyError"];
	        this.proxy = source["proxy"];
	        this.apiEndpoint = source["apiEndpoint"];
	    }
	}

}

