export namespace models {
	
	export class AICharacter {
	    id: number;
	    nickname: string;
	    gender: string;
	    age: number;
	    birthdate: string;
	    region: string;
	    hobby: string;
	    job_category: string;
	    mbti: string;
	    aggression_level: number;
	    formality_level: number;
	    roleplay_level: number;
	    assigned_model_index: number;
	    post_count: number;
	    comment_count: number;
	    persona_summary: string;
	    is_active: boolean;
	    // Go type: time
	    created_at: any;
	
	    static createFrom(source: any = {}) {
	        return new AICharacter(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.nickname = source["nickname"];
	        this.gender = source["gender"];
	        this.age = source["age"];
	        this.birthdate = source["birthdate"];
	        this.region = source["region"];
	        this.hobby = source["hobby"];
	        this.job_category = source["job_category"];
	        this.mbti = source["mbti"];
	        this.aggression_level = source["aggression_level"];
	        this.formality_level = source["formality_level"];
	        this.roleplay_level = source["roleplay_level"];
	        this.assigned_model_index = source["assigned_model_index"];
	        this.post_count = source["post_count"];
	        this.comment_count = source["comment_count"];
	        this.persona_summary = source["persona_summary"];
	        this.is_active = source["is_active"];
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
	export class BBSConfig {
	    title: string;
	    footer: string;
	    theme: string;
	    font: string;
	    posts_per_page: number;
	
	    static createFrom(source: any = {}) {
	        return new BBSConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.title = source["title"];
	        this.footer = source["footer"];
	        this.theme = source["theme"];
	        this.font = source["font"];
	        this.posts_per_page = source["posts_per_page"];
	    }
	}
	export class Comment {
	    id: number;
	    post_id: number;
	    parent_id?: number;
	    author_type: string;
	    author_id: number;
	    author_nickname: string;
	    content: string;
	    // Go type: time
	    created_at: any;
	    replies?: Comment[];
	
	    static createFrom(source: any = {}) {
	        return new Comment(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.post_id = source["post_id"];
	        this.parent_id = source["parent_id"];
	        this.author_type = source["author_type"];
	        this.author_id = source["author_id"];
	        this.author_nickname = source["author_nickname"];
	        this.content = source["content"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.replies = this.convertValues(source["replies"], Comment);
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
	export class LLMConfig {
	    host: string;
	    port: string;
	    model1: string;
	    model2: string;
	    model3: string;
	    posts_per_hour: number;
	    comments_per_hour: number;
	    max_tokens: number;
	    temperature: number;
	
	    static createFrom(source: any = {}) {
	        return new LLMConfig(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.host = source["host"];
	        this.port = source["port"];
	        this.model1 = source["model1"];
	        this.model2 = source["model2"];
	        this.model3 = source["model3"];
	        this.posts_per_hour = source["posts_per_hour"];
	        this.comments_per_hour = source["comments_per_hour"];
	        this.max_tokens = source["max_tokens"];
	        this.temperature = source["temperature"];
	    }
	}
	export class Post {
	    id: number;
	    author_type: string;
	    author_id: number;
	    author_nickname: string;
	    title: string;
	    content: string;
	    view_count: number;
	    recommend_count: number;
	    comment_count: number;
	    is_pinned: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new Post(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.author_type = source["author_type"];
	        this.author_id = source["author_id"];
	        this.author_nickname = source["author_nickname"];
	        this.title = source["title"];
	        this.content = source["content"];
	        this.view_count = source["view_count"];
	        this.recommend_count = source["recommend_count"];
	        this.comment_count = source["comment_count"];
	        this.is_pinned = source["is_pinned"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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
	export class PostList {
	    posts: Post[];
	    total_count: number;
	    page: number;
	    per_page: number;
	    total_pages: number;
	
	    static createFrom(source: any = {}) {
	        return new PostList(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.posts = this.convertValues(source["posts"], Post);
	        this.total_count = source["total_count"];
	        this.page = source["page"];
	        this.per_page = source["per_page"];
	        this.total_pages = source["total_pages"];
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
	export class User {
	    id: number;
	    username: string;
	    nickname: string;
	    is_admin: boolean;
	    // Go type: time
	    created_at: any;
	    // Go type: time
	    updated_at: any;
	
	    static createFrom(source: any = {}) {
	        return new User(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.username = source["username"];
	        this.nickname = source["nickname"];
	        this.is_admin = source["is_admin"];
	        this.created_at = this.convertValues(source["created_at"], null);
	        this.updated_at = this.convertValues(source["updated_at"], null);
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

