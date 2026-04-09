<script lang="ts">
	import { scheduling, type SuggestionJSON, type PlanJSON } from '$lib/api';

	let suggestions: SuggestionJSON[] = $state([]);
	let plan: PlanJSON | null = $state(null);
	let loading = $state(true);
	let error = $state('');
	let viewMode: 'suggest' | 'plan' = $state('suggest');

	async function loadSuggestions() {
		loading = true;
		error = '';
		try {
			const res = await scheduling.suggest({ time: 60, energy: 'medium', count: 10 });
			suggestions = res.items;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load suggestions';
		} finally {
			loading = false;
		}
	}

	async function loadPlan() {
		loading = true;
		error = '';
		try {
			plan = await scheduling.planToday({ time: 480, energy: 'medium' });
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load plan';
		} finally {
			loading = false;
		}
	}

	function switchView(mode: 'suggest' | 'plan') {
		viewMode = mode;
		if (mode === 'suggest') loadSuggestions();
		else loadPlan();
	}

	function formatScore(n: number): string {
		return n.toFixed(2);
	}

	function formatDuration(min: number): string {
		if (min < 60) return `${min}m`;
		const h = Math.floor(min / 60);
		const m = min % 60;
		return m === 0 ? `${h}h` : `${h}h${m}m`;
	}

	function scorePercent(score: number): number {
		return Math.min(100, Math.round((score / 0.95) * 100));
	}

	function formatClock(minutes: number): string {
		return `${Math.floor(minutes / 60)}:${String(minutes % 60).padStart(2, '0')}`;
	}

	function planSlots(items: SuggestionJSON[]): { s: SuggestionJSON; start: number; end: number }[] {
		let elapsed = 0;
		return items.map((s) => {
			const start = elapsed;
			const end = elapsed + s.duration;
			elapsed = end;
			return { s, start, end };
		});
	}

	$effect(() => {
		loadSuggestions();
	});
</script>

<div class="view">
	<h1>📅 Today</h1>

	<div class="tabs">
		<button class:active={viewMode === 'suggest'} onclick={() => switchView('suggest')}>
			What Now?
		</button>
		<button class:active={viewMode === 'plan'} onclick={() => switchView('plan')}>
			Day Plan
		</button>
	</div>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if viewMode === 'suggest'}
		{#if suggestions.length === 0}
			<p class="empty">No suggestions. Add some tasks first!</p>
		{:else}
			<ul class="suggestion-list">
				{#each suggestions as s, i}
					<li class="suggestion-item">
						<span class="rank">{i + 1}</span>
						<div class="suggestion-body">
							<div class="suggestion-title">{s.title}</div>
							<div class="suggestion-meta">
								<span class="duration">~{s.duration}m</span>
								{#if s.priority && s.priority !== 'none'}
									<span class="badge priority-{s.priority}">{s.priority}</span>
								{/if}
								{#if s.energy}
									<span class="badge energy">{s.energy}</span>
								{/if}
							</div>
							<div class="score-bar">
								<div class="score-fill" style="width: {scorePercent(s.score.total)}%"></div>
							</div>
							<span class="score-label">{formatScore(s.score.total)}</span>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	{:else if plan}
		{#if plan.suggestions.length === 0}
			<p class="empty">Nothing to plan. Add some tasks first!</p>
		{:else}
			<div class="plan-header">
				<span>{formatDuration(plan.total_duration)} packed</span>
				<span class="plan-free">{formatDuration(plan.time_remaining)} free</span>
			</div>
			<ul class="plan-list">
				{#each planSlots(plan.suggestions) as { s, start, end }}
					<li class="plan-item">
						<span class="plan-time">{formatClock(start)} &ndash; {formatClock(end)}</span>
						<div class="plan-body">
							<span class="plan-title">{s.title}</span>
							<span class="plan-duration">{s.duration}m</span>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	{/if}
</div>

<style>
	.view { max-width: 700px; }
	h1 { margin-bottom: 1rem; font-size: 1.5rem; }

	.tabs {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
	}

	.tabs button {
		padding: 0.5rem 1rem;
		background: #1a1a1a;
		border: 1px solid #333;
		border-radius: 6px;
		color: #999;
		cursor: pointer;
		font-size: 0.85rem;
	}

	.tabs button.active {
		background: #2563eb;
		border-color: #2563eb;
		color: white;
	}

	.error {
		padding: 0.6rem;
		background: #3b1111;
		border: 1px solid #5c2020;
		border-radius: 6px;
		color: #ff6b6b;
		margin-bottom: 1rem;
		font-size: 0.85rem;
	}

	.empty { color: #666; text-align: center; padding: 3rem; }

	.suggestion-list { list-style: none; }

	.suggestion-item {
		display: flex;
		gap: 0.75rem;
		padding: 0.75rem 0;
		border-bottom: 1px solid #1f1f1f;
		align-items: flex-start;
	}

	.rank {
		font-size: 0.8rem;
		color: #555;
		font-weight: 600;
		min-width: 1.5rem;
		text-align: right;
		margin-top: 0.15rem;
	}

	.suggestion-body { flex: 1; }

	.suggestion-title { font-size: 0.95rem; margin-bottom: 0.25rem; }

	.suggestion-meta {
		display: flex;
		gap: 0.35rem;
		align-items: center;
		margin-bottom: 0.35rem;
	}

	.duration { font-size: 0.75rem; color: #888; }

	.badge {
		font-size: 0.65rem;
		padding: 0.1rem 0.35rem;
		border-radius: 3px;
		background: #252525;
		color: #888;
	}

	.priority-urgent { background: #5c1a1a; color: #ff6b6b; }
	.priority-high { background: #5c3a1a; color: #ffaa6b; }
	.priority-medium { background: #4a4a1a; color: #e0d06b; }
	.priority-low { background: #1a3a1a; color: #6bcc6b; }
	.energy { background: #1a2a4a; color: #6b9bff; }

	.score-bar {
		height: 4px;
		background: #252525;
		border-radius: 2px;
		width: 100px;
		display: inline-block;
		vertical-align: middle;
	}

	.score-fill {
		height: 100%;
		background: #2563eb;
		border-radius: 2px;
		transition: width 0.3s;
	}

	.score-label {
		font-size: 0.7rem;
		color: #666;
		margin-left: 0.4rem;
	}

	.plan-header {
		display: flex;
		justify-content: space-between;
		margin-bottom: 1rem;
		font-size: 0.85rem;
		color: #888;
	}

	.plan-free { color: #4ade80; }

	.plan-list { list-style: none; }

	.plan-item {
		display: flex;
		gap: 1rem;
		padding: 0.6rem 0;
		border-bottom: 1px solid #1f1f1f;
		align-items: center;
	}

	.plan-time {
		font-size: 0.8rem;
		color: #666;
		font-family: monospace;
		min-width: 90px;
	}

	.plan-body {
		flex: 1;
		display: flex;
		justify-content: space-between;
		align-items: center;
	}

	.plan-title { font-size: 0.9rem; }
	.plan-duration { font-size: 0.75rem; color: #666; }
</style>
