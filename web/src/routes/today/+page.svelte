<script lang="ts">
	import { onMount } from 'svelte';
	import { scheduling, meta, type SuggestionJSON, type PlanJSON } from '$lib/api';

	let suggestions: SuggestionJSON[] = $state([]);
	let plan: PlanJSON | null = $state(null);
	let loading = $state(true);
	let error = $state('');
	let viewMode: 'suggest' | 'plan' = $state('suggest');

	// Controls
	let availTime = $state(60);
	let planTime = $state(480);
	let energy: 'low' | 'medium' | 'high' = $state('medium');
	let context = $state('');
	let count = $state(5);
	let contexts: string[] = $state([]);

	// Score expansion
	let expandedIdx: number | null = $state(null);

	async function loadContexts() {
		try {
			const res = await meta.contexts();
			contexts = res.items;
		} catch {
			// non-critical
		}
	}

	async function loadSuggestions() {
		loading = true;
		error = '';
		try {
			const res = await scheduling.suggest({
				time: availTime,
				energy,
				context: context || undefined,
				count
			});
			suggestions = res.items;
			expandedIdx = null;
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
			plan = await scheduling.planToday({
				time: planTime,
				energy,
				context: context || undefined
			});
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load plan';
		} finally {
			loading = false;
		}
	}

	function switchView(mode: 'suggest' | 'plan') {
		viewMode = mode;
		reload();
	}

	function reload() {
		if (viewMode === 'suggest') loadSuggestions();
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

	function toggleScoreExpand(idx: number) {
		expandedIdx = expandedIdx === idx ? null : idx;
	}

	const scoreLabels: Record<string, string> = {
		urgency: 'Urgency',
		priority: 'Priority',
		energy_match: 'Energy Match',
		streak_risk: 'Streak Risk',
		age_boost: 'Age Boost',
		dependency_unlock: 'Dep. Unlock'
	};

	onMount(() => {
		loadContexts();
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

	<div class="controls">
		<label>
			<span class="ctrl-label">{viewMode === 'suggest' ? 'Available' : 'Total'} time</span>
			{#if viewMode === 'suggest'}
				<input type="number" min="5" max="600" bind:value={availTime} onchange={reload} />
			{:else}
				<input type="number" min="5" max="960" bind:value={planTime} onchange={reload} />
			{/if}
			<span class="ctrl-unit">min</span>
		</label>

		<div class="energy-toggle">
			<span class="ctrl-label">Energy</span>
			<div class="toggle-group">
				{#each (['low', 'medium', 'high'] as const) as lvl}
					<button
						class:active={energy === lvl}
						onclick={() => { energy = lvl; reload(); }}
					>{lvl}</button>
				{/each}
			</div>
		</div>

		<label>
			<span class="ctrl-label">Context</span>
			<select bind:value={context} onchange={reload}>
				<option value="">Any</option>
				{#each contexts as ctx}
					<option value={ctx}>{ctx}</option>
				{/each}
			</select>
		</label>

		{#if viewMode === 'suggest'}
			<label>
				<span class="ctrl-label">Count</span>
				<input type="number" min="1" max="20" bind:value={count} onchange={reload} />
			</label>
		{/if}

		<button class="refresh-btn" onclick={reload}>Refresh</button>
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
							<button class="suggestion-title" onclick={() => toggleScoreExpand(i)}>
								{s.title}
							</button>
							<div class="suggestion-meta">
								<span class="duration">~{s.duration}m</span>
								{#if s.priority && s.priority !== 'none'}
									<span class="badge priority-{s.priority}">{s.priority}</span>
								{/if}
								{#if s.energy}
									<span class="badge energy">{s.energy}</span>
								{/if}
								{#if s.due_date}
									<span class="badge due">due {s.due_date}</span>
								{/if}
							</div>
							<div class="score-row">
								<div class="score-bar">
									<div class="score-fill" style="width: {scorePercent(s.score.total)}%"></div>
								</div>
								<span class="score-label">{formatScore(s.score.total)}</span>
							</div>

							{#if expandedIdx === i}
								<div class="score-breakdown">
									{#each Object.entries(scoreLabels) as [key, label]}
										{@const val = s.score[key as keyof typeof s.score]}
										{#if val > 0}
											<div class="score-factor">
												<span class="factor-label">{label}</span>
												<div class="factor-bar">
													<div class="factor-fill" style="width: {Math.round(val * 100)}%"></div>
												</div>
												<span class="factor-value">{val.toFixed(2)}</span>
											</div>
										{/if}
									{/each}
								</div>
							{/if}
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
				<span class="plan-util">{Math.round((plan.total_duration / plan.available_time) * 100)}% utilized</span>
				<span class="plan-free">{formatDuration(plan.time_remaining)} free</span>
			</div>
			<ul class="plan-list">
				{#each planSlots(plan.suggestions) as { s, start, end }}
					<li class="plan-item">
						<span class="plan-time">{formatClock(start)} &ndash; {formatClock(end)}</span>
						<div class="plan-body">
							<span class="plan-title">{s.title}</span>
							<div class="plan-score">
								<div class="score-bar">
									<div class="score-fill" style="width: {scorePercent(s.score.total)}%"></div>
								</div>
								<span class="score-label">{formatScore(s.score.total)}</span>
							</div>
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
		margin-bottom: 1rem;
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

	/* Controls */
	.controls {
		display: flex;
		gap: 0.75rem;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
		align-items: flex-end;
	}

	.controls label {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		font-size: 0.75rem;
	}

	.ctrl-label { color: #666; font-size: 0.7rem; }
	.ctrl-unit { color: #555; font-size: 0.7rem; }

	.controls input[type="number"] {
		width: 65px;
		padding: 0.35rem 0.5rem;
		background: #1a1a1a;
		border: 1px solid #333;
		border-radius: 4px;
		color: #e0e0e0;
		font-size: 0.8rem;
	}

	.controls select {
		padding: 0.35rem 0.5rem;
		background: #1a1a1a;
		border: 1px solid #333;
		border-radius: 4px;
		color: #ccc;
		font-size: 0.8rem;
	}

	.controls input:focus, .controls select:focus { outline: none; border-color: #555; }

	.energy-toggle {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
	}

	.toggle-group {
		display: flex;
		border: 1px solid #333;
		border-radius: 4px;
		overflow: hidden;
	}

	.toggle-group button {
		padding: 0.35rem 0.5rem;
		background: #1a1a1a;
		border: none;
		border-right: 1px solid #333;
		color: #999;
		cursor: pointer;
		font-size: 0.75rem;
		text-transform: capitalize;
	}

	.toggle-group button:last-child { border-right: none; }
	.toggle-group button.active { background: #2563eb; color: white; }

	.refresh-btn {
		padding: 0.35rem 0.7rem;
		background: #252525;
		border: 1px solid #333;
		border-radius: 4px;
		color: #ccc;
		cursor: pointer;
		font-size: 0.8rem;
		align-self: flex-end;
	}

	.refresh-btn:hover { border-color: #2563eb; color: #fff; }

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

	/* Suggestions */
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

	.suggestion-title {
		font-size: 0.95rem;
		margin-bottom: 0.25rem;
		background: none;
		border: none;
		color: #e0e0e0;
		cursor: pointer;
		padding: 0;
		text-align: left;
		width: 100%;
	}

	.suggestion-title:hover { color: #fff; }

	.suggestion-meta {
		display: flex;
		gap: 0.35rem;
		align-items: center;
		margin-bottom: 0.35rem;
	}

	.duration { font-size: 0.75rem; color: #888; }

	.score-row {
		display: flex;
		align-items: center;
		gap: 0.4rem;
	}

	.score-bar {
		height: 4px;
		background: #252525;
		border-radius: 2px;
		width: 100px;
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
	}

	/* Score breakdown */
	.score-breakdown {
		margin-top: 0.5rem;
		padding: 0.5rem 0;
		border-top: 1px solid #252525;
	}

	.score-factor {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.3rem;
	}

	.factor-label {
		font-size: 0.7rem;
		color: #888;
		width: 80px;
	}

	.factor-bar {
		flex: 1;
		height: 3px;
		background: #252525;
		border-radius: 2px;
		max-width: 80px;
	}

	.factor-fill {
		height: 100%;
		background: #4ade80;
		border-radius: 2px;
	}

	.factor-value {
		font-size: 0.65rem;
		color: #666;
		width: 30px;
		text-align: right;
		font-family: monospace;
	}

	/* Plan */
	.plan-header {
		display: flex;
		justify-content: space-between;
		margin-bottom: 1rem;
		font-size: 0.85rem;
		color: #888;
	}

	.plan-util { color: #2563eb; }
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
		align-items: center;
		gap: 0.75rem;
	}

	.plan-title { font-size: 0.9rem; flex: 1; }
	.plan-score { display: flex; align-items: center; gap: 0.3rem; }
	.plan-duration { font-size: 0.75rem; color: #666; min-width: 30px; text-align: right; }
</style>
