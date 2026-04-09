<script lang="ts">
	import { onMount } from 'svelte';
	import { scheduling, habits, type SuggestionJSON, type HabitStats, type TaskJSON } from '$lib/api';

	interface HabitDisplay {
		title: string;
		streak: number;
		completedToday: boolean;
	}

	let now = $state(new Date());
	let suggestions: SuggestionJSON[] = $state([]);
	let habitDisplays: HabitDisplay[] = $state([]);
	let loadError = $state('');

	function formatTime(d: Date): string {
		return d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', hour12: false });
	}

	function formatDate(d: Date): string {
		return d.toLocaleDateString([], { weekday: 'long', month: 'long', day: 'numeric' });
	}

	async function loadData() {
		loadError = '';
		try {
			const [suggestRes, habitsRes] = await Promise.all([
				scheduling.suggest({ count: 3, energy: 'medium', time: 120 }),
				habits.list()
			]);
			suggestions = suggestRes.items;

			const displays: HabitDisplay[] = [];
			for (const h of habitsRes.items) {
				try {
					const statsRes = await habits.stats(h.id);
					displays.push({
						title: h.title,
						streak: statsRes.stats.current_streak,
						completedToday: statsRes.stats.completed_today
					});
				} catch {
					displays.push({ title: h.title, streak: 0, completedToday: false });
				}
			}
			habitDisplays = displays;
		} catch (e) {
			loadError = e instanceof Error ? e.message : 'Failed to load';
		}
	}

	function scorePercent(score: number): number {
		return Math.min(100, Math.round((score / 0.95) * 100));
	}

	onMount(() => {
		loadData();

		// Clock tick every second
		const clockInterval = setInterval(() => { now = new Date(); }, 1000);

		// Data refresh every 60 seconds
		const dataInterval = setInterval(loadData, 60000);

		return () => {
			clearInterval(clockInterval);
			clearInterval(dataInterval);
		};
	});
</script>

<div class="mirror">
	<!-- Clock -->
	<section class="clock-section">
		<div class="clock-time">{formatTime(now)}</div>
		<div class="clock-date">{formatDate(now)}</div>
	</section>

	<!-- Next Tasks -->
	<section class="tasks-section">
		<h2 class="section-title">Next Up</h2>
		{#if suggestions.length === 0}
			<p class="section-empty">No tasks scheduled</p>
		{:else}
			<ul class="task-list">
				{#each suggestions as s, i}
					<li class="task-item">
						<span class="task-rank">{i + 1}</span>
						<div class="task-body">
							<span class="task-title">{s.title}</span>
							<div class="task-meta">
								<span class="task-duration">{s.duration}m</span>
								<div class="score-bar">
									<div class="score-fill" style="width: {scorePercent(s.score.total)}%"></div>
								</div>
							</div>
						</div>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	<!-- Habit Streaks -->
	<section class="habits-section">
		<h2 class="section-title">Streaks</h2>
		{#if habitDisplays.length === 0}
			<p class="section-empty">No habits tracked</p>
		{:else}
			<ul class="habit-list">
				{#each habitDisplays as h}
					<li class="habit-item" class:done={h.completedToday} class:at-risk={!h.completedToday && h.streak > 0}>
						<span class="habit-status">
							{#if h.completedToday}✓{:else if h.streak > 0}!{:else}○{/if}
						</span>
						<span class="habit-title">{h.title}</span>
						<span class="habit-streak">🔥 {h.streak}</span>
					</li>
				{/each}
			</ul>
		{/if}
	</section>

	{#if loadError}
		<div class="mirror-error">{loadError}</div>
	{/if}
</div>

<style>
	.mirror {
		display: flex;
		flex-direction: column;
		gap: 3rem;
		max-width: 600px;
		margin: 0 auto;
	}

	/* Clock */
	.clock-section { text-align: center; }

	.clock-time {
		font-size: 5rem;
		font-weight: 200;
		letter-spacing: 0.05em;
		line-height: 1;
		color: #fff;
	}

	.clock-date {
		font-size: 1.2rem;
		color: #888;
		margin-top: 0.5rem;
	}

	/* Sections */
	.section-title {
		font-size: 1rem;
		font-weight: 500;
		color: #666;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		margin-bottom: 1rem;
		border-bottom: 1px solid #222;
		padding-bottom: 0.5rem;
	}

	.section-empty {
		color: #444;
		font-size: 1.2rem;
	}

	/* Tasks */
	.task-list { list-style: none; }

	.task-item {
		display: flex;
		align-items: center;
		gap: 1rem;
		padding: 0.8rem 0;
		border-bottom: 1px solid #1a1a1a;
	}

	.task-rank {
		font-size: 1.4rem;
		font-weight: 300;
		color: #9b7ede;
		min-width: 2rem;
		text-align: center;
	}

	.task-body { flex: 1; }

	.task-title {
		font-size: 1.3rem;
		font-weight: 400;
		display: block;
		margin-bottom: 0.3rem;
	}

	.task-meta {
		display: flex;
		align-items: center;
		gap: 0.75rem;
	}

	.task-duration {
		font-size: 0.9rem;
		color: #666;
	}

	.score-bar {
		width: 80px;
		height: 4px;
		background: #222;
		border-radius: 2px;
		overflow: hidden;
	}

	.score-fill {
		height: 100%;
		background: #9b7ede;
		border-radius: 2px;
	}

	/* Habits */
	.habit-list { list-style: none; }

	.habit-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.6rem 0;
		font-size: 1.2rem;
	}

	.habit-item.done { opacity: 0.5; }

	.habit-status {
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		font-size: 0.8rem;
		font-weight: 700;
	}

	.habit-item.done .habit-status { background: #09814a; color: #fff; }
	.habit-item.at-risk .habit-status { background: #b6174b; color: #fff; }
	.habit-item:not(.done):not(.at-risk) .habit-status { border: 1px solid #333; color: #555; }

	.habit-title { flex: 1; }
	.habit-streak { color: #888; font-size: 1rem; }

	.mirror-error {
		color: #b6174b;
		font-size: 0.9rem;
		text-align: center;
		margin-top: 1rem;
	}
</style>
