export namespace main {
	
	export class Project {
	    id: number;
	    name: string;
	    description: string;
	    color: string;
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
	    status: string;
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
	        this.status = source["status"];
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

