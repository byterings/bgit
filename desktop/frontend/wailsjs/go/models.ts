export namespace main {
	
	export class AppInfo {
	    version: string;
	    hookHelperPath?: string;
	
	    static createFrom(source: any = {}) {
	        return new AppInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.version = source["version"];
	        this.hookHelperPath = source["hookHelperPath"];
	    }
	}
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
	export class WorkspaceView {
	    path: string;
	    alias: string;
	    identityName?: string;
	    identityEmail?: string;
	    missingIdentity: boolean;
	    exists: boolean;
	    repositoryCount: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.alias = source["alias"];
	        this.identityName = source["identityName"];
	        this.identityEmail = source["identityEmail"];
	        this.missingIdentity = source["missingIdentity"];
	        this.exists = source["exists"];
	        this.repositoryCount = source["repositoryCount"];
	    }
	}
	export class RepoBindingView {
	    path: string;
	    alias: string;
	    name?: string;
	    email?: string;
	    githubUsername?: string;
	    missingIdentity: boolean;
	
	    static createFrom(source: any = {}) {
	        return new RepoBindingView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.alias = source["alias"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.githubUsername = source["githubUsername"];
	        this.missingIdentity = source["missingIdentity"];
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
	    bindings: RepoBindingView[];
	    workspaces: WorkspaceView[];
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
	        this.bindings = this.convertValues(source["bindings"], RepoBindingView);
	        this.workspaces = this.convertValues(source["workspaces"], WorkspaceView);
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
	export class BranchRequest {
	    path: string;
	    name: string;
	
	    static createFrom(source: any = {}) {
	        return new BranchRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	    }
	}
	export class CloneDefaults {
	    parentDir: string;
	    fromWorkspace: boolean;
	
	    static createFrom(source: any = {}) {
	        return new CloneDefaults(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.parentDir = source["parentDir"];
	        this.fromWorkspace = source["fromWorkspace"];
	    }
	}
	export class CloneRequest {
	    url: string;
	    parentDir: string;
	    folderName: string;
	    alias: string;
	
	    static createFrom(source: any = {}) {
	        return new CloneRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.url = source["url"];
	        this.parentDir = source["parentDir"];
	        this.folderName = source["folderName"];
	        this.alias = source["alias"];
	    }
	}
	export class CommitChangesRequest {
	    path: string;
	    summary: string;
	    description: string;
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new CommitChangesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.summary = source["summary"];
	        this.description = source["description"];
	        this.files = source["files"];
	    }
	}
	export class CommitFilesRequest {
	    path: string;
	    commit: string;
	
	    static createFrom(source: any = {}) {
	        return new CommitFilesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.commit = source["commit"];
	    }
	}
	export class ConnectionTestResult {
	    alias: string;
	    githubUsername: string;
	    passed: boolean;
	    authenticated?: string;
	    message: string;
	    fix?: string;
	
	    static createFrom(source: any = {}) {
	        return new ConnectionTestResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.alias = source["alias"];
	        this.githubUsername = source["githubUsername"];
	        this.passed = source["passed"];
	        this.authenticated = source["authenticated"];
	        this.message = source["message"];
	        this.fix = source["fix"];
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
	
	export class DiscardChangesRequest {
	    path: string;
	    files: string[];
	
	    static createFrom(source: any = {}) {
	        return new DiscardChangesRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.files = source["files"];
	    }
	}
	
	
	
	
	export class FileDiffRequest {
	    path: string;
	    file: string;
	    commit: string;
	
	    static createFrom(source: any = {}) {
	        return new FileDiffRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.file = source["file"];
	        this.commit = source["commit"];
	    }
	}
	export class RepositoryView {
	    path: string;
	    name: string;
	    source: string;
	    sourcePath?: string;
	    alias: string;
	    identitySource?: string;
	    identityName?: string;
	    identityEmail?: string;
	    githubUsername?: string;
	    exists: boolean;
	    health: string;
	    problems: string[];
	    notes: string[];
	    canFix: boolean;
	    remoteUrl?: string;
	    branch?: string;
	    changes: number;
	    ahead: number;
	    behind: number;
	
	    static createFrom(source: any = {}) {
	        return new RepositoryView(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.name = source["name"];
	        this.source = source["source"];
	        this.sourcePath = source["sourcePath"];
	        this.alias = source["alias"];
	        this.identitySource = source["identitySource"];
	        this.identityName = source["identityName"];
	        this.identityEmail = source["identityEmail"];
	        this.githubUsername = source["githubUsername"];
	        this.exists = source["exists"];
	        this.health = source["health"];
	        this.problems = source["problems"];
	        this.notes = source["notes"];
	        this.canFix = source["canFix"];
	        this.remoteUrl = source["remoteUrl"];
	        this.branch = source["branch"];
	        this.changes = source["changes"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	    }
	}
	export class RepoState {
	    repository?: RepositoryView;
	    status?: models.GitStatus;
	
	    static createFrom(source: any = {}) {
	        return new RepoState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repository = this.convertValues(source["repository"], RepositoryView);
	        this.status = this.convertValues(source["status"], models.GitStatus);
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
	export class GitActionResult {
	    ok: boolean;
	    kind?: string;
	    message: string;
	    detail?: string;
	    state?: RepoState;
	
	    static createFrom(source: any = {}) {
	        return new GitActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ok = source["ok"];
	        this.kind = source["kind"];
	        this.message = source["message"];
	        this.detail = source["detail"];
	        this.state = this.convertValues(source["state"], RepoState);
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
	export class HistoryRequest {
	    path: string;
	    limit: number;
	    skip: number;
	
	    static createFrom(source: any = {}) {
	        return new HistoryRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.limit = source["limit"];
	        this.skip = source["skip"];
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
	export class IdentityDetectionResult {
	    path: string;
	    repoRoot?: string;
	    alias?: string;
	    source?: string;
	    sourcePath?: string;
	    name?: string;
	    email?: string;
	    githubUsername?: string;
	    gitConfigScope?: string;
	    gitName?: string;
	    gitEmail?: string;
	    gitMatches: boolean;
	    gitConfigChecked: boolean;
	    syncAvailable: boolean;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new IdentityDetectionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.repoRoot = source["repoRoot"];
	        this.alias = source["alias"];
	        this.source = source["source"];
	        this.sourcePath = source["sourcePath"];
	        this.name = source["name"];
	        this.email = source["email"];
	        this.githubUsername = source["githubUsername"];
	        this.gitConfigScope = source["gitConfigScope"];
	        this.gitName = source["gitName"];
	        this.gitEmail = source["gitEmail"];
	        this.gitMatches = source["gitMatches"];
	        this.gitConfigChecked = source["gitConfigChecked"];
	        this.syncAvailable = source["syncAvailable"];
	        this.message = source["message"];
	    }
	}
	export class IdentityDetectionActionResult {
	    message: string;
	    detected?: IdentityDetectionResult;
	    status?: DesktopStatus;
	
	    static createFrom(source: any = {}) {
	        return new IdentityDetectionActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.detected = this.convertValues(source["detected"], IdentityDetectionResult);
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
	export class IdentityDetectionRequest {
	    path: string;
	
	    static createFrom(source: any = {}) {
	        return new IdentityDetectionRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
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
	
	export class OnboardingState {
	    gitInstalled: boolean;
	    gitVersion?: string;
	    configured: boolean;
	    setupCompleted: boolean;
	    pushProtection: boolean;
	    hooksPath?: string;
	    hooksPathManaged: boolean;
	    identityCount: number;
	    activeAlias?: string;
	    needed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new OnboardingState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.gitInstalled = source["gitInstalled"];
	        this.gitVersion = source["gitVersion"];
	        this.configured = source["configured"];
	        this.setupCompleted = source["setupCompleted"];
	        this.pushProtection = source["pushProtection"];
	        this.hooksPath = source["hooksPath"];
	        this.hooksPathManaged = source["hooksPathManaged"];
	        this.identityCount = source["identityCount"];
	        this.activeAlias = source["activeAlias"];
	        this.needed = source["needed"];
	    }
	}
	export class RepoBindingActionResult {
	    message: string;
	    path?: string;
	    status?: DesktopStatus;
	
	    static createFrom(source: any = {}) {
	        return new RepoBindingActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.path = source["path"];
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
	export class RepoBindingRequest {
	    path: string;
	    alias: string;
	
	    static createFrom(source: any = {}) {
	        return new RepoBindingRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.alias = source["alias"];
	    }
	}
	
	
	export class RepositoriesResult {
	    repositories: RepositoryView[];
	    needAttention: number;
	    message?: string;
	
	    static createFrom(source: any = {}) {
	        return new RepositoriesResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.repositories = this.convertValues(source["repositories"], RepositoryView);
	        this.needAttention = source["needAttention"];
	        this.message = source["message"];
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
	export class RepositoryActionResult {
	    message: string;
	    repository?: RepositoryView;
	    status?: DesktopStatus;
	
	    static createFrom(source: any = {}) {
	        return new RepositoryActionResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.message = source["message"];
	        this.repository = this.convertValues(source["repository"], RepositoryView);
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
	
	export class UninstallRequest {
	    removeKeys: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UninstallRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.removeKeys = source["removeKeys"];
	    }
	}
	export class UninstallResult {
	    restoredRepos: string[];
	    failedRepos: string[];
	    removedKeys: string[];
	    gitIdentityNote?: string;
	    warnings: string[];
	    configRemoved: boolean;
	
	    static createFrom(source: any = {}) {
	        return new UninstallResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.restoredRepos = source["restoredRepos"];
	        this.failedRepos = source["failedRepos"];
	        this.removedKeys = source["removedKeys"];
	        this.gitIdentityNote = source["gitIdentityNote"];
	        this.warnings = source["warnings"];
	        this.configRemoved = source["configRemoved"];
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
	export class WorkspaceActionResult {
	    message: string;
	    status?: DesktopStatus;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceActionResult(source);
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
	export class WorkspaceRequest {
	    path: string;
	    alias: string;
	
	    static createFrom(source: any = {}) {
	        return new WorkspaceRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.alias = source["alias"];
	    }
	}

}

export namespace models {
	
	export class GitBranch {
	    name: string;
	    remote: boolean;
	    current: boolean;
	    upstream: string;
	    oid: string;
	    commitTime: number;
	
	    static createFrom(source: any = {}) {
	        return new GitBranch(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.remote = source["remote"];
	        this.current = source["current"];
	        this.upstream = source["upstream"];
	        this.oid = source["oid"];
	        this.commitTime = source["commitTime"];
	    }
	}
	export class GitCommit {
	    oid: string;
	    shortOid: string;
	    authorName: string;
	    authorEmail: string;
	    time: number;
	    summary: string;
	    body: string;
	    unpushed: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GitCommit(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.oid = source["oid"];
	        this.shortOid = source["shortOid"];
	        this.authorName = source["authorName"];
	        this.authorEmail = source["authorEmail"];
	        this.time = source["time"];
	        this.summary = source["summary"];
	        this.body = source["body"];
	        this.unpushed = source["unpushed"];
	    }
	}
	export class GitDiffLine {
	    kind: string;
	    text: string;
	    oldLine: number;
	    newLine: number;
	
	    static createFrom(source: any = {}) {
	        return new GitDiffLine(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.kind = source["kind"];
	        this.text = source["text"];
	        this.oldLine = source["oldLine"];
	        this.newLine = source["newLine"];
	    }
	}
	export class GitDiff {
	    path: string;
	    binary: boolean;
	    truncated: boolean;
	    additions: number;
	    deletions: number;
	    lines: GitDiffLine[];
	
	    static createFrom(source: any = {}) {
	        return new GitDiff(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.binary = source["binary"];
	        this.truncated = source["truncated"];
	        this.additions = source["additions"];
	        this.deletions = source["deletions"];
	        this.lines = this.convertValues(source["lines"], GitDiffLine);
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
	
	export class GitFileChange {
	    path: string;
	    origPath: string;
	    kind: string;
	    staged: boolean;
	    unstaged: boolean;
	
	    static createFrom(source: any = {}) {
	        return new GitFileChange(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.path = source["path"];
	        this.origPath = source["origPath"];
	        this.kind = source["kind"];
	        this.staged = source["staged"];
	        this.unstaged = source["unstaged"];
	    }
	}
	export class GitStatus {
	    branch: string;
	    detached: boolean;
	    headOid: string;
	    unborn: boolean;
	    upstream: string;
	    hasUpstream: boolean;
	    ahead: number;
	    behind: number;
	    files: GitFileChange[];
	    truncated: boolean;
	    mergeInProgress: boolean;
	    rebaseInProgress: boolean;
	    conflicts: number;
	    unresolvedConflicts: number;
	
	    static createFrom(source: any = {}) {
	        return new GitStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.branch = source["branch"];
	        this.detached = source["detached"];
	        this.headOid = source["headOid"];
	        this.unborn = source["unborn"];
	        this.upstream = source["upstream"];
	        this.hasUpstream = source["hasUpstream"];
	        this.ahead = source["ahead"];
	        this.behind = source["behind"];
	        this.files = this.convertValues(source["files"], GitFileChange);
	        this.truncated = source["truncated"];
	        this.mergeInProgress = source["mergeInProgress"];
	        this.rebaseInProgress = source["rebaseInProgress"];
	        this.conflicts = source["conflicts"];
	        this.unresolvedConflicts = source["unresolvedConflicts"];
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

