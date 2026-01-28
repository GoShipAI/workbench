export namespace main {
	
	export class Agent {
	    id: number;
	    name: string;
	    description: string;
	    prompt: string;
	    provider_id?: number;
	    model: string;
	    enabled: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Agent(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.prompt = source["prompt"];
	        this.provider_id = source["provider_id"];
	        this.model = source["model"];
	        this.enabled = source["enabled"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class AgentInput {
	    id: number;
	    name: string;
	    description: string;
	    prompt: string;
	    provider_id?: number;
	    model: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new AgentInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.prompt = source["prompt"];
	        this.provider_id = source["provider_id"];
	        this.model = source["model"];
	        this.enabled = source["enabled"];
	    }
	}
	export class CompleteTaskInput {
	    id: number;
	    actual_start?: string;
	    actual_hours: number;
	
	    static createFrom(source: any = {}) {
	        return new CompleteTaskInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.actual_start = source["actual_start"];
	        this.actual_hours = source["actual_hours"];
	    }
	}
	export class DailyTaskStats {
	    date: string;
	    total_count: number;
	    completed_count: number;
	    completion_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new DailyTaskStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.total_count = source["total_count"];
	        this.completed_count = source["completed_count"];
	        this.completion_rate = source["completion_rate"];
	    }
	}
	export class ModelProvider {
	    id: number;
	    name: string;
	    label: string;
	    api_key: string;
	    base_url: string;
	    enabled: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new ModelProvider(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.label = source["label"];
	        this.api_key = source["api_key"];
	        this.base_url = source["base_url"];
	        this.enabled = source["enabled"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class ModelProviderInput {
	    id: number;
	    name: string;
	    label: string;
	    api_key: string;
	    base_url: string;
	    enabled: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ModelProviderInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.label = source["label"];
	        this.api_key = source["api_key"];
	        this.base_url = source["base_url"];
	        this.enabled = source["enabled"];
	    }
	}
	export class Project {
	    id: number;
	    name: string;
	    description: string;
	    color: string;
	    archived: boolean;
	    // Go type: time
	    created_at: any;
	    task_count: number;
	
	    static createFrom(source: any = {}) {
	        return new Project(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.color = source["color"];
	        this.archived = source["archived"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.task_count = source["task_count"];
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
	export class ProjectTimeStats {
	    project_id: number;
	    project_name: string;
	    color: string;
	    total_hours: number;
	    task_count: number;
	    percentage: number;
	
	    static createFrom(source: any = {}) {
	        return new ProjectTimeStats(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_id = source["project_id"];
	        this.project_name = source["project_name"];
	        this.color = source["color"];
	        this.total_hours = source["total_hours"];
	        this.task_count = source["task_count"];
	        this.percentage = source["percentage"];
	    }
	}
	export class ReportSummary {
	    total_tasks: number;
	    completed_tasks: number;
	    total_hours: number;
	    completed_hours: number;
	    average_rate: number;
	
	    static createFrom(source: any = {}) {
	        return new ReportSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.total_tasks = source["total_tasks"];
	        this.completed_tasks = source["completed_tasks"];
	        this.total_hours = source["total_hours"];
	        this.completed_hours = source["completed_hours"];
	        this.average_rate = source["average_rate"];
	    }
	}
	export class ReportData {
	    project_stats: ProjectTimeStats[];
	    daily_stats: DailyTaskStats[];
	    summary: ReportSummary;
	
	    static createFrom(source: any = {}) {
	        return new ReportData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.project_stats = this.convertValues(source["project_stats"], ProjectTimeStats);
	        this.daily_stats = this.convertValues(source["daily_stats"], DailyTaskStats);
	        this.summary = this.convertValues(source["summary"], ReportSummary);
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
	
	export class Task {
	    id: number;
	    project_id?: number;
	    project_name: string;
	    name: string;
	    description: string;
	    date?: string;
	    start_time?: string;
	    end_time?: string;
	    hours: number;
	    deadline?: string;
	    priority: string;
	    urgency: string;
	    status: string;
	    actual_start?: string;
	    actual_hours: number;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Task(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.project_name = source["project_name"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.date = source["date"];
	        this.start_time = source["start_time"];
	        this.end_time = source["end_time"];
	        this.hours = source["hours"];
	        this.deadline = source["deadline"];
	        this.priority = source["priority"];
	        this.urgency = source["urgency"];
	        this.status = source["status"];
	        this.actual_start = source["actual_start"];
	        this.actual_hours = source["actual_hours"];
	        this.created_at = this.convertValues(source["created_at"], null);
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
	export class TaskInput {
	    id: number;
	    project_id?: number;
	    name: string;
	    description: string;
	    date?: string;
	    start_time?: string;
	    end_time?: string;
	    hours: number;
	    deadline?: string;
	    priority: string;
	    urgency: string;
	    status: string;
	
	    static createFrom(source: any = {}) {
	        return new TaskInput(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.project_id = source["project_id"];
	        this.name = source["name"];
	        this.description = source["description"];
	        this.date = source["date"];
	        this.start_time = source["start_time"];
	        this.end_time = source["end_time"];
	        this.hours = source["hours"];
	        this.deadline = source["deadline"];
	        this.priority = source["priority"];
	        this.urgency = source["urgency"];
	        this.status = source["status"];
	    }
	}
	export class WorkbenchData {
	    today_tasks: Task[];
	    total_count: number;
	    completed_count: number;
	    planned_hours: number;
	    completed_hours: number;
	    pending_count: number;
	
	    static createFrom(source: any = {}) {
	        return new WorkbenchData(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.today_tasks = this.convertValues(source["today_tasks"], Task);
	        this.total_count = source["total_count"];
	        this.completed_count = source["completed_count"];
	        this.planned_hours = source["planned_hours"];
	        this.completed_hours = source["completed_hours"];
	        this.pending_count = source["pending_count"];
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

