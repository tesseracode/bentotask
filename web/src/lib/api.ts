// API client for BentoTask REST API.
// All functions call /api/v1/* which is proxied to the Go backend in dev.

const BASE = '/api/v1';

export interface TaskJSON {
	id: string;
	title: string;
	type: string;
	status: string;
	priority?: string;
	energy?: string;
	estimated_duration?: number;
	due_date?: string;
	due_start?: string;
	due_end?: string;
	box?: string;
	recurrence?: string;
	completed_at?: string;
	created_at: string;
	updated_at: string;
	tags: string[];
	contexts: string[];
	file_path: string;
	body?: string;
	steps?: StepJSON[];
	schedule?: ScheduleJSON;
	links?: Record<string, string>[];
}

export interface StepJSON {
	title: string;
	duration?: number;
	ref?: string;
	optional?: boolean;
}

export interface ScheduleJSON {
	time?: string;
	days?: string[];
}

export interface SuggestionJSON {
	task_id: string;
	title: string;
	duration: number;
	score: ScoreBreakdown;
	priority?: string;
	energy?: string;
	due_date?: string;
	tags: string[];
	contexts: string[];
}

export interface ScoreBreakdown {
	urgency: number;
	priority: number;
	energy_match: number;
	streak_risk: number;
	age_boost: number;
	dependency_unlock: number;
	total: number;
}

export interface PlanJSON {
	suggestions: SuggestionJSON[];
	total_duration: number;
	time_remaining: number;
	available_time: number;
}

export interface HabitStats {
	current_streak: number;
	longest_streak: number;
	total_completions: number;
	completion_rate: number;
	rate_period_days: number;
	completed_today: boolean;
}

export interface Collection<T> {
	items: T[];
	count: number;
}

export interface APIError {
	error: {
		code: string;
		message: string;
	};
}

export interface CreateTaskRequest {
	title: string;
	priority?: string;
	energy?: string;
	duration?: number;
	due_date?: string;
	tags?: string[];
	contexts?: string[];
	box?: string;
}

export interface UpdateTaskRequest {
	title?: string;
	priority?: string;
	energy?: string;
	duration?: number;
	due_date?: string;
	tags?: string[];
	contexts?: string[];
	box?: string;
	status?: string;
	steps?: { title: string; duration?: number; optional?: boolean }[];
	schedule?: { time?: string; days?: string[] };
}

async function request<T>(method: string, path: string, body?: unknown): Promise<T> {
	const opts: RequestInit = {
		method,
		headers: { 'Content-Type': 'application/json' }
	};
	if (body !== undefined) {
		opts.body = JSON.stringify(body);
	}

	const res = await fetch(`${BASE}${path}`, opts);
	const data = await res.json();

	if (!res.ok) {
		const err = data as APIError;
		throw new Error(err.error?.message ?? `API error: ${res.status}`);
	}

	return data as T;
}

// --- Tasks ---

export const tasks = {
	create: (req: CreateTaskRequest) => request<TaskJSON>('POST', '/tasks', req),

	list: (params?: Record<string, string>) => {
		const qs = params ? '?' + new URLSearchParams(params).toString() : '';
		return request<Collection<TaskJSON>>('GET', `/tasks${qs}`);
	},

	get: (id: string) => request<TaskJSON>('GET', `/tasks/${id}`),

	update: (id: string, req: UpdateTaskRequest) => request<TaskJSON>('PATCH', `/tasks/${id}`, req),

	delete: (id: string) => request<TaskJSON>('DELETE', `/tasks/${id}`),

	done: (id: string) => request<TaskJSON>('POST', `/tasks/${id}/done`),

	search: (q: string) => request<Collection<TaskJSON>>('GET', `/tasks/search?q=${encodeURIComponent(q)}`)
};

// --- Habits ---

export const habits = {
	create: (req: { title: string; freq_type?: string; freq_target?: number; priority?: string; energy?: string }) =>
		request<TaskJSON>('POST', '/habits', req),

	list: () => request<Collection<TaskJSON>>('GET', '/habits'),

	log: (id: string, req?: { duration?: number; note?: string }) =>
		request<TaskJSON>('POST', `/habits/${id}/log`, req ?? {}),

	stats: (id: string) => request<{ task: TaskJSON; stats: HabitStats }>('GET', `/habits/${id}/stats`)
};

// --- Routines ---

export const routines = {
	create: (req: { title: string; steps?: { title: string; duration?: number; optional?: boolean }[]; priority?: string; energy?: string }) =>
		request<TaskJSON>('POST', '/routines', req),

	list: () => request<Collection<TaskJSON>>('GET', '/routines'),
	get: (id: string) => request<TaskJSON>('GET', `/routines/${id}`)
};

// --- Scheduling ---

export const scheduling = {
	suggest: (params?: { time?: number; energy?: string; context?: string; count?: number }) => {
		const qs = params ? '?' + new URLSearchParams(
			Object.entries(params)
				.filter(([, v]) => v !== undefined)
				.map(([k, v]) => [k, String(v)])
		).toString() : '';
		return request<Collection<SuggestionJSON>>('GET', `/suggest${qs}`);
	},

	planToday: (params?: { time?: number; energy?: string; context?: string }) => {
		const qs = params ? '?' + new URLSearchParams(
			Object.entries(params)
				.filter(([, v]) => v !== undefined)
				.map(([k, v]) => [k, String(v)])
		).toString() : '';
		return request<PlanJSON>('GET', `/plan/today${qs}`);
	}
};

// --- Meta ---

export const meta = {
	tags: () => request<Collection<string>>('GET', '/meta/tags'),
	boxes: () => request<Collection<string>>('GET', '/meta/boxes'),
	contexts: () => request<Collection<string>>('GET', '/meta/contexts')
};
