export namespace main {
	
	export class DoctorCheckView {
	    level: string;
	    message: string;
	    fix?: string;
	
	    static createFrom(source: any = {}) {
	        return new DoctorCheckView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.level = source["level"];
	        this.message = source["message"];
	        this.fix = source["fix"];
	    }
	}
	export class DoctorSection {
	    title: string;
	    checks: DoctorCheckView[];
	
	    static createFrom(source: any = {}) {
	        return new DoctorSection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.checks = this.convertValues(source["checks"], DoctorCheckView);
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
	export class DoctorStatus {
	    configured: boolean;
	    errors: number;
	    warnings: number;
	    sections: DoctorSection[];
	
	    static createFrom(source: any = {}) {
	        return new DoctorStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configured = source["configured"];
	        this.errors = source["errors"];
	        this.warnings = source["warnings"];
	        this.sections = this.convertValues(source["sections"], DoctorSection);
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
	export class EffectiveIdentityView {
	    alias: string;
	    source: string;
	    path?: string;
	
	    static createFrom(source: any = {}) {
	        return new EffectiveIdentityView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.source = source["source"];
	        this.path = source["path"];
	    }
	}
	export class IdentityView {
	    alias: string;
	    name: string;
	    email: string;
	    githubUsername: string;
	    githubAvatarUrl: string;
	    sshKeyPath: string;
	    sshPublicKey?: string;
	    sshPublicKeyStatus: string;
	    sshKeyStatus: string;
	    active: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IdentityView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.githubUsername = source["githubUsername"];
	        this.githubAvatarUrl = source["githubAvatarUrl"];
	        this.sshKeyPath = source["sshKeyPath"];
	        this.sshPublicKey = source["sshPublicKey"];
	        this.sshPublicKeyStatus = source["sshPublicKeyStatus"];
	        this.sshKeyStatus = source["sshKeyStatus"];
	        this.active = source["active"];
	    }
	}
	export class DesktopStatus {
	    configured: boolean;
	    setupCompleted: boolean;
	    activeAlias: string;
	    identityCount: number;
	    workspaceCount: number;
	    bindingCount: number;
	    identities: IdentityView[];
	    activeIdentity?: IdentityView;
	    effectiveIdentity?: EffectiveIdentityView;
	
	    static createFrom(source: any = {}) {
	        return new DesktopStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.configured = source["configured"];
	        this.setupCompleted = source["setupCompleted"];
	        this.activeAlias = source["activeAlias"];
	        this.identityCount = source["identityCount"];
	        this.workspaceCount = source["workspaceCount"];
	        this.bindingCount = source["bindingCount"];
	        this.identities = this.convertValues(source["identities"], IdentityView);
	        this.activeIdentity = this.convertValues(source["activeIdentity"], IdentityView);
	        this.effectiveIdentity = this.convertValues(source["effectiveIdentity"], EffectiveIdentityView);
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
	export class BackupActionResult {
	    message: string;
	    archivePath?: string;
	    usersCount?: number;
	    workspacesCount?: number;
	    bindingsCount?: number;
	    activeUser?: string;
	    status?: DesktopStatus;
	    doctor?: DoctorStatus;
	
	    static createFrom(source: any = {}) {
	        return new BackupActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.archivePath = source["archivePath"];
	        this.usersCount = source["usersCount"];
	        this.workspacesCount = source["workspacesCount"];
	        this.bindingsCount = source["bindingsCount"];
	        this.activeUser = source["activeUser"];
	        this.status = this.convertValues(source["status"], DesktopStatus);
	        this.doctor = this.convertValues(source["doctor"], DoctorStatus);
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
	export class BackupExportRequest {
	    password: string;
	    confirmPassword: string;
	    outputPath: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupExportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.password = source["password"];
	        this.confirmPassword = source["confirmPassword"];
	        this.outputPath = source["outputPath"];
	    }
	}
	export class BackupImportRequest {
	    password: string;
	    archivePath: string;
	
	    static createFrom(source: any = {}) {
	        return new BackupImportRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.password = source["password"];
	        this.archivePath = source["archivePath"];
	    }
	}
	export class DeleteIdentityRequest {
	    alias: string;
	    deleteKeys: boolean;
	
	    static createFrom(source: any = {}) {
	        return new DeleteIdentityRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.deleteKeys = source["deleteKeys"];
	    }
	}
	
	
	
	
	
	export class IdentityActionResult {
	    message: string;
	    status?: DesktopStatus;
	
	    static createFrom(source: any = {}) {
	        return new IdentityActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.status = this.convertValues(source["status"], DesktopStatus);
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
	export class IdentityRequest {
	    alias: string;
	    name: string;
	    email: string;
	    githubUsername: string;
	    sshKeyPath: string;
	    generateSSHKey: boolean;
	
	    static createFrom(source: any = {}) {
	        return new IdentityRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.githubUsername = source["githubUsername"];
	        this.sshKeyPath = source["sshKeyPath"];
	        this.generateSSHKey = source["generateSSHKey"];
	    }
	}
	
	export class UpdateIdentityRequest {
	    alias: string;
	    name: string;
	    email: string;
	    githubUsername: string;
	    sshKeyPath: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateIdentityRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.githubUsername = source["githubUsername"];
	        this.sshKeyPath = source["sshKeyPath"];
	    }
	}

}

