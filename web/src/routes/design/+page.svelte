<script lang="ts">
	type Priority = 'urgent' | 'high' | 'medium' | 'low';
	type Energy = 'high' | 'medium' | 'low';
	type ThemeName = 'bento' | 'bento-alt' | 'clay' | 'clay-alt';
	type ViewName = 'inbox' | 'today' | 'habits';

	interface MockTask {
		id: string;
		title: string;
		priority: Priority;
		energy: Energy;
		due_date?: string;
		tags: string[];
		duration: number;
	}

	interface ScoreBreakdown {
		total: number;
		urgency: number;
		priority: number;
		energy_match: number;
		streak_risk: number;
		age_boost: number;
		dependency_unlock: number;
	}

	interface MockSuggestion {
		title: string;
		duration: number;
		priority: Priority;
		energy: Energy;
		score: ScoreBreakdown;
	}

	interface MockHabit {
		id: string;
		title: string;
		streak: number;
		longest: number;
		total: number;
		rate: number;
		completedToday: boolean;
		freq: string;
	}

	const mockTasks: MockTask[] = [
		{ id: '01ABC', title: 'Write quarterly report', priority: 'urgent', energy: 'high', due_date: '2026-04-10', tags: ['work'], duration: 60 },
		{ id: '02DEF', title: 'Buy groceries', priority: 'medium', energy: 'low', tags: ['errands', 'home'], duration: 30 },
		{ id: '03GHI', title: 'Review PR #42', priority: 'high', energy: 'medium', tags: ['work'], duration: 15 },
		{ id: '04JKL', title: 'Read chapter 5', priority: 'low', energy: 'low', tags: ['learning'], duration: 20 },
		{ id: '05MNO', title: 'Fix CI pipeline', priority: 'high', energy: 'high', due_date: '2026-04-09', tags: ['work', 'devops'], duration: 45 },
	];

	const mockSuggestions: MockSuggestion[] = [
		{ title: 'Fix CI pipeline', duration: 45, priority: 'high', energy: 'high', score: { total: 0.72, urgency: 0.8, priority: 0.75, energy_match: 1.0, streak_risk: 0, age_boost: 0.15, dependency_unlock: 0.3 } },
		{ title: 'Write quarterly report', duration: 60, priority: 'urgent', energy: 'high', score: { total: 0.65, urgency: 1.0, priority: 1.0, energy_match: 0.5, streak_risk: 0, age_boost: 0.1, dependency_unlock: 0 } },
		{ title: 'Review PR #42', duration: 15, priority: 'high', energy: 'medium', score: { total: 0.48, urgency: 0, priority: 0.75, energy_match: 1.0, streak_risk: 0, age_boost: 0.08, dependency_unlock: 0.5 } },
	];

	const mockHabits: MockHabit[] = [
		{ id: 'h1', title: 'Meditate', streak: 15, longest: 30, total: 145, rate: 0.87, completedToday: true, freq: 'daily' },
		{ id: 'h2', title: 'Exercise', streak: 3, longest: 12, total: 48, rate: 0.65, completedToday: false, freq: 'daily' },
		{ id: 'h3', title: 'Read 30 pages', streak: 0, longest: 7, total: 22, rate: 0.42, completedToday: false, freq: 'daily' },
		{ id: 'h4', title: 'Review weekly goals', streak: 8, longest: 15, total: 32, rate: 0.92, completedToday: true, freq: 'weekly' },
	];

	const themes: { name: ThemeName; label: string }[] = [
		{ name: 'bento', label: 'Bento' },
		{ name: 'bento-alt', label: 'Bento Alt' },
		{ name: 'clay', label: 'Clay' },
		{ name: 'clay-alt', label: 'Clay Alt' },
	];

	const views: { name: ViewName; label: string; icon: string }[] = [
		{ name: 'inbox', label: 'Inbox', icon: '📥' },
		{ name: 'today', label: 'Today', icon: '📅' },
		{ name: 'habits', label: 'Habits', icon: '🔥' },
	];

	let activeTheme: ThemeName = $state('bento');
	let activeView: ViewName = $state('inbox');
	let expandedScoreIdx: number | null = $state(null);

	const scoreLabels: Record<string, string> = {
		urgency: 'Urgency',
		priority: 'Priority',
		energy_match: 'Energy Match',
		streak_risk: 'Streak Risk',
		age_boost: 'Age Boost',
		dependency_unlock: 'Dep. Unlock',
	};

	function scorePercent(score: number): number {
		return Math.min(100, Math.round((score / 0.95) * 100));
	}

	function toggleScore(idx: number): void {
		expandedScoreIdx = expandedScoreIdx === idx ? null : idx;
	}
</script>
<svelte:head>
	<link rel="preconnect" href="https://fonts.googleapis.com" />
	<link rel="preconnect" href="https://fonts.gstatic.com" crossorigin="anonymous" />
	<link href="https://fonts.googleapis.com/css2?family=Playfair+Display:ital,wght@0,400;0,600;0,700;1,400&family=JetBrains+Mono:wght@300;400;500;600&family=Nunito:wght@400;600;700;800&family=IBM+Plex+Sans:wght@300;400;500;600&display=swap" rel="stylesheet" />
</svelte:head>

<!-- Control Bar -->
<div class="design-controls">
	<div class="controls-inner">
		<div class="control-group">
			<span class="control-label">Theme</span>
			<div class="theme-buttons">
				{#each themes as t}
					<button
						class="theme-btn"
						class:active={activeTheme === t.name}
						data-theme={t.name}
						onclick={() => activeTheme = t.name}
					>
						{t.label}
					</button>
				{/each}
			</div>
		</div>
		<div class="control-group">
			<span class="control-label">View</span>
			<div class="view-buttons">
				{#each views as v}
					<button
						class="view-btn"
						class:active={activeView === v.name}
						onclick={() => { activeView = v.name; expandedScoreIdx = null; }}
					>
						<span class="view-icon">{v.icon}</span>
						{v.label}
					</button>
				{/each}
			</div>
		</div>
	</div>
</div>

<!-- Themed Content Area -->
<div class="theme-wrapper theme-{activeTheme}">
	<div class="themed-content">

		{#if activeView === 'inbox'}
			<!-- INBOX VIEW -->
			<div class="view-header">
				<h2 class="view-title">Inbox</h2>
				<span class="view-count">{mockTasks.length} tasks</span>
			</div>

			<div class="add-input-bar">
				<span class="add-icon">+</span>
				<span class="add-placeholder">Add a task...</span>
			</div>

			<ul class="task-list">
				{#each mockTasks as task (task.id)}
					<li class="task-item">
						<button class="task-checkbox" aria-label="Complete {task.title}">
							<span class="checkbox-inner"></span>
						</button>
						<div class="task-content">
							<span class="task-title">{task.title}</span>
							<div class="task-badges">
								<span class="badge badge-priority badge-{task.priority}">{task.priority}</span>
								<span class="badge badge-energy">{task.energy} energy</span>
								{#each task.tags as tag}
									<span class="badge badge-tag">#{tag}</span>
								{/each}
								{#if task.due_date}
									<span class="badge badge-due">due {task.due_date}</span>
								{/if}
								<span class="badge badge-duration">{task.duration}m</span>
							</div>
						</div>
					</li>
				{/each}
			</ul>

		{:else if activeView === 'today'}
			<!-- TODAY VIEW -->
			<div class="view-header">
				<h2 class="view-title">What Now?</h2>
				<span class="view-subtitle">Ranked by scheduling score</span>
			</div>

			<ul class="suggestion-list">
				{#each mockSuggestions as s, i (s.title)}
					<li class="suggestion-item">
						<span class="suggestion-rank">#{i + 1}</span>
						<div class="suggestion-content">
							<div class="suggestion-header">
								<button class="suggestion-title" onclick={() => toggleScore(i)}>
									{s.title}
								</button>
								<span class="suggestion-duration">{s.duration}m</span>
							</div>
							<div class="suggestion-badges">
								<span class="badge badge-priority badge-{s.priority}">{s.priority}</span>
								<span class="badge badge-energy">{s.energy} energy</span>
							</div>
							<div class="score-row">
								<div class="score-track">
									<div class="score-fill" style="width: {scorePercent(s.score.total)}%"></div>
								</div>
								<span class="score-value">{s.score.total.toFixed(2)}</span>
							</div>

							{#if expandedScoreIdx === i}
								<div class="score-breakdown">
									<span class="breakdown-label">Score Breakdown</span>
									{#each Object.entries(scoreLabels) as [key, label]}
										{@const val = s.score[key as keyof ScoreBreakdown]}
										{#if typeof val === 'number' && val > 0}
											<div class="factor-row">
												<span class="factor-name">{label}</span>
												<div class="factor-track">
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

		{:else}
			<!-- HABITS VIEW -->
			<div class="view-header">
				<h2 class="view-title">Habits</h2>
				<span class="view-count">{mockHabits.length} tracked</span>
			</div>

			<ul class="habit-list">
				{#each mockHabits as habit (habit.id)}
					<li class="habit-item" class:habit-done={habit.completedToday}>
						<div class="habit-status-indicator">
							{#if habit.completedToday}
								<span class="habit-check">&#10003;</span>
							{:else if habit.streak > 0}
								<span class="habit-warning">!</span>
							{:else}
								<span class="habit-empty">&#9675;</span>
							{/if}
						</div>
						<div class="habit-content">
							<div class="habit-header-row">
								<span class="habit-title">{habit.title}</span>
								<span class="habit-freq">{habit.freq}</span>
							</div>
							<div class="habit-stats">
								<span class="habit-stat" title="Current streak">🔥 {habit.streak}</span>
								<span class="habit-stat-sep">|</span>
								<span class="habit-stat" title="Longest streak">Best: {habit.longest}</span>
								<span class="habit-stat-sep">|</span>
								<span class="habit-stat" title="Total completions">Total: {habit.total}</span>
								<span class="habit-stat-sep">|</span>
								<span class="habit-stat" title="Completion rate">{Math.round(habit.rate * 100)}%</span>
							</div>
							<div class="habit-rate-bar">
								<div class="habit-rate-fill" style="width: {Math.round(habit.rate * 100)}%"></div>
							</div>
						</div>
						<button class="habit-log-btn" class:logged={habit.completedToday}>
							{habit.completedToday ? '✓ Done' : 'Log'}
						</button>
					</li>
				{/each}
			</ul>
		{/if}
	</div>
</div>
<style>
	/* ======================================================
	   CONTROL BAR — neutral, always the same regardless of theme
	   ====================================================== */
	.design-controls {
		position: sticky;
		top: 0;
		z-index: 100;
		background: #111113;
		border-bottom: 1px solid #2a2a2c;
		padding: 0.75rem 1.5rem;
		margin: -2rem -2rem 0;
	}

	.controls-inner {
		display: flex;
		align-items: center;
		gap: 2rem;
		max-width: 800px;
	}

	.control-group {
		display: flex;
		align-items: center;
		gap: 0.6rem;
	}

	.control-label {
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #666;
		font-weight: 500;
	}

	.theme-buttons, .view-buttons {
		display: flex;
		gap: 0.25rem;
	}

	.theme-btn, .view-btn {
		padding: 0.35rem 0.7rem;
		background: #1c1c1e;
		border: 1px solid #333;
		border-radius: 5px;
		color: #888;
		cursor: pointer;
		font-size: 0.75rem;
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		transition: all 0.15s ease;
	}

	.theme-btn:hover, .view-btn:hover {
		color: #ccc;
		border-color: #555;
	}

	.theme-btn.active {
		background: #6366f1;
		border-color: #6366f1;
		color: #fff;
	}

	.theme-btn[data-theme="bento"].active { background: #6366f1; border-color: #6366f1; }
	.theme-btn[data-theme="bento-alt"].active { background: #9b7ede; border-color: #9b7ede; }
	.theme-btn[data-theme="clay"].active { background: #c45b3a; border-color: #c45b3a; }
	.theme-btn[data-theme="clay-alt"].active { background: #9d6381; border-color: #9d6381; }

	.view-btn.active {
		background: #252530;
		border-color: #6366f1;
		color: #e5e5e7;
	}

	.view-icon { margin-right: 0.2rem; }

	/* ======================================================
	   THEME WRAPPER — the canvas for themed content
	   ====================================================== */
	.theme-wrapper {
		margin: 0 -2rem;
		padding: 2rem;
		min-height: calc(100vh - 60px);
		transition: background-color 0.3s ease, color 0.3s ease;
	}

	.themed-content {
		max-width: 700px;
		transition: all 0.3s ease;
	}

	/* ======================================================
	   BASE COMPONENT STYLES (overridden per-theme)
	   ====================================================== */

	/* -- View Header -- */
	.view-header {
		display: flex;
		align-items: baseline;
		gap: 0.75rem;
		margin-bottom: 1.5rem;
		transition: all 0.3s ease;
	}

	.view-title {
		font-size: 1.5rem;
		font-weight: 600;
		transition: all 0.3s ease;
	}

	.view-count, .view-subtitle {
		font-size: 0.8rem;
		opacity: 0.5;
		transition: all 0.3s ease;
	}

	/* -- Add Input Bar -- */
	.add-input-bar {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.7rem 1rem;
		margin-bottom: 1.25rem;
		border-radius: 6px;
		cursor: text;
		transition: all 0.3s ease;
	}

	.add-icon {
		font-size: 1rem;
		font-weight: 600;
		opacity: 0.4;
		transition: all 0.3s ease;
	}

	.add-placeholder {
		font-size: 0.9rem;
		opacity: 0.35;
		transition: all 0.3s ease;
	}

	/* -- Task List -- */
	.task-list, .suggestion-list, .habit-list {
		list-style: none;
		padding: 0;
		margin: 0;
	}

	.task-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.8rem 0;
		transition: all 0.3s ease;
	}

	.task-checkbox {
		flex-shrink: 0;
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.15rem;
		margin-top: 0.1rem;
	}

	.checkbox-inner {
		display: block;
		width: 18px;
		height: 18px;
		border-radius: 50%;
		transition: all 0.3s ease;
	}

	.task-content {
		flex: 1;
		min-width: 0;
	}

	.task-title {
		display: block;
		font-size: 0.95rem;
		margin-bottom: 0.3rem;
		transition: all 0.3s ease;
	}

	.task-badges, .suggestion-badges {
		display: flex;
		flex-wrap: wrap;
		gap: 0.3rem;
		transition: all 0.3s ease;
	}

	/* -- Badges -- */
	.badge {
		display: inline-block;
		padding: 0.15rem 0.45rem;
		font-size: 0.65rem;
		font-weight: 500;
		border-radius: 3px;
		transition: all 0.3s ease;
	}

	/* -- Suggestion List -- */
	.suggestion-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 1rem 0;
		transition: all 0.3s ease;
	}

	.suggestion-rank {
		font-weight: 700;
		font-size: 0.85rem;
		min-width: 2rem;
		text-align: center;
		margin-top: 0.15rem;
		transition: all 0.3s ease;
	}

	.suggestion-content { flex: 1; }

	.suggestion-header {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 0.35rem;
	}

	.suggestion-title {
		background: none;
		border: none;
		padding: 0;
		cursor: pointer;
		font-size: 1rem;
		font-weight: 500;
		text-align: left;
		transition: all 0.3s ease;
	}

	.suggestion-duration {
		font-size: 0.75rem;
		opacity: 0.6;
		white-space: nowrap;
		transition: all 0.3s ease;
	}

	/* -- Score Bars -- */
	.score-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-top: 0.5rem;
	}

	.score-track {
		flex: 1;
		height: 4px;
		border-radius: 2px;
		overflow: hidden;
		transition: all 0.3s ease;
	}

	.score-fill {
		height: 100%;
		border-radius: 2px;
		transition: width 0.4s ease, background 0.3s ease;
	}

	.score-value {
		font-size: 0.7rem;
		font-weight: 500;
		min-width: 2rem;
		text-align: right;
		opacity: 0.6;
		transition: all 0.3s ease;
	}

	/* -- Score Breakdown -- */
	.score-breakdown {
		margin-top: 0.75rem;
		padding-top: 0.75rem;
		transition: all 0.3s ease;
	}

	.breakdown-label {
		display: block;
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		margin-bottom: 0.5rem;
		opacity: 0.5;
		transition: all 0.3s ease;
	}

	.factor-row {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		margin-bottom: 0.35rem;
	}

	.factor-name {
		font-size: 0.7rem;
		width: 90px;
		opacity: 0.7;
		transition: all 0.3s ease;
	}

	.factor-track {
		flex: 1;
		height: 3px;
		border-radius: 2px;
		max-width: 100px;
		overflow: hidden;
		transition: all 0.3s ease;
	}

	.factor-fill {
		height: 100%;
		border-radius: 2px;
		transition: width 0.4s ease, background 0.3s ease;
	}

	.factor-value {
		font-size: 0.65rem;
		width: 2rem;
		text-align: right;
		opacity: 0.6;
		font-variant-numeric: tabular-nums;
		transition: all 0.3s ease;
	}

	/* -- Habit List -- */
	.habit-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.85rem 0;
		transition: all 0.3s ease;
	}

	.habit-status-indicator {
		flex-shrink: 0;
		width: 24px;
		height: 24px;
		display: flex;
		align-items: center;
		justify-content: center;
		border-radius: 50%;
		font-size: 0.7rem;
		font-weight: 700;
		margin-top: 0.1rem;
		transition: all 0.3s ease;
	}

	.habit-content { flex: 1; min-width: 0; }

	.habit-header-row {
		display: flex;
		justify-content: space-between;
		align-items: baseline;
		margin-bottom: 0.25rem;
	}

	.habit-title {
		font-size: 0.95rem;
		font-weight: 500;
		transition: all 0.3s ease;
	}

	.habit-freq {
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
		opacity: 0.45;
		transition: all 0.3s ease;
	}

	.habit-stats {
		display: flex;
		gap: 0.4rem;
		align-items: center;
		margin-bottom: 0.4rem;
	}

	.habit-stat {
		font-size: 0.72rem;
		opacity: 0.65;
		transition: all 0.3s ease;
	}

	.habit-stat-sep {
		font-size: 0.6rem;
		opacity: 0.25;
	}

	.habit-rate-bar {
		height: 3px;
		border-radius: 2px;
		overflow: hidden;
		transition: all 0.3s ease;
	}

	.habit-rate-fill {
		height: 100%;
		border-radius: 2px;
		transition: width 0.4s ease, background 0.3s ease;
	}

	.habit-log-btn {
		flex-shrink: 0;
		padding: 0.4rem 0.8rem;
		border-radius: 6px;
		font-size: 0.75rem;
		font-weight: 500;
		cursor: pointer;
		border: 1px solid transparent;
		transition: all 0.3s ease;
	}

	/* ======================================================
	   THEME: BENTO — The One We Ship
	   Feels like: Arc browser meets Things 3
	   ====================================================== */
	.theme-bento {
		background: #111113;
		color: #e5e5e7;
		font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', Roboto, sans-serif;
	}

	.theme-bento .view-title {
		font-weight: 600;
		font-size: 1.5rem;
		color: #e5e5e7;
		letter-spacing: -0.01em;
	}

	.theme-bento .view-count,
	.theme-bento .view-subtitle {
		color: #6366f1;
		font-weight: 500;
		font-size: 0.8rem;
	}

	.theme-bento .view-header {
		margin-bottom: 1.25rem;
	}

	.theme-bento .add-input-bar {
		background: #1c1c1e;
		border: 1px solid #2a2a2c;
		border-radius: 10px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}

	.theme-bento .add-icon { color: #6366f1; }
	.theme-bento .add-placeholder { color: #555; }

	.theme-bento .task-item {
		background: #1c1c1e;
		border: 1px solid #2a2a2c;
		border-radius: 10px;
		padding: 0.8rem 1rem;
		margin-bottom: 0.4rem;
		border-bottom: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
	}

	.theme-bento .checkbox-inner {
		border: 2px solid #6366f1;
		border-radius: 50%;
		width: 18px;
		height: 18px;
	}

	.theme-bento .task-title {
		color: #e5e5e7;
		font-weight: 500;
		font-size: 0.93rem;
	}

	.theme-bento .badge {
		background: #252528;
		color: #999;
		border-radius: 5px;
		font-size: 0.6rem;
		font-weight: 500;
		padding: 0.15rem 0.45rem;
	}

	.theme-bento .badge-priority {
		background: rgba(99, 102, 241, 0.12);
		color: #818cf8;
	}

	.theme-bento .badge-urgent {
		background: rgba(239, 68, 68, 0.12);
		color: #f87171;
	}

	.theme-bento .badge-high {
		background: rgba(245, 158, 11, 0.12);
		color: #fbbf24;
	}

	.theme-bento .badge-energy {
		background: rgba(134, 203, 146, 0.1);
		color: #86cb92;
	}

	.theme-bento .badge-due {
		background: rgba(245, 158, 11, 0.1);
		color: #f59e0b;
	}

	.theme-bento .badge-tag {
		color: #777;
	}

	/* Bento: suggestions */
	.theme-bento .suggestion-item {
		background: #1c1c1e;
		border: 1px solid #2a2a2c;
		border-radius: 10px;
		padding: 1rem;
		margin-bottom: 0.4rem;
		border-bottom: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
	}

	.theme-bento .suggestion-rank {
		color: #6366f1;
		font-weight: 700;
		font-size: 0.9rem;
	}

	.theme-bento .suggestion-title {
		color: #e5e5e7;
		font-weight: 500;
		font-size: 0.95rem;
	}

	.theme-bento .suggestion-title:hover { color: #818cf8; }

	.theme-bento .suggestion-duration {
		color: #666;
		font-weight: 500;
	}

	.theme-bento .score-track {
		background: #252528;
		height: 6px;
		border-radius: 3px;
	}

	.theme-bento .score-fill {
		background: #6366f1;
		border-radius: 3px;
	}

	.theme-bento .score-value {
		color: #818cf8;
		font-weight: 600;
		font-size: 0.7rem;
	}

	.theme-bento .score-breakdown {
		border-top: 1px solid #2a2a2c;
	}

	.theme-bento .breakdown-label {
		color: #555;
		font-weight: 500;
	}

	.theme-bento .factor-name {
		color: #888;
		font-size: 0.68rem;
	}

	.theme-bento .factor-track {
		background: #252528;
		height: 4px;
		border-radius: 2px;
	}

	.theme-bento .factor-fill {
		background: #6366f1;
		border-radius: 2px;
	}

	.theme-bento .factor-value { color: #6366f1; font-weight: 500; }

	/* Bento: habits */
	.theme-bento .habit-item {
		background: #1c1c1e;
		border: 1px solid #2a2a2c;
		border-radius: 10px;
		padding: 0.85rem 1rem;
		margin-bottom: 0.4rem;
		border-bottom: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
	}

	.theme-bento .habit-status-indicator {
		border-radius: 50%;
		width: 24px;
		height: 24px;
	}

	.theme-bento .habit-check { color: #fff; }

	.theme-bento .habit-item:has(.habit-check) .habit-status-indicator {
		background: #86cb92;
	}

	.theme-bento .habit-item:has(.habit-warning) .habit-status-indicator {
		background: #f59e0b;
	}

	.theme-bento .habit-warning { color: #fff; }

	.theme-bento .habit-item:has(.habit-empty) .habit-status-indicator {
		background: #252528;
		border: 1px solid #3a3a3c;
	}

	.theme-bento .habit-empty { color: #555; }

	.theme-bento .habit-title {
		color: #e5e5e7;
		font-weight: 500;
	}

	.theme-bento .habit-freq { color: #555; font-weight: 500; }

	.theme-bento .habit-stat {
		color: #888;
		font-size: 0.7rem;
	}

	.theme-bento .habit-stat-sep { color: #333; }

	.theme-bento .habit-rate-bar {
		background: #252528;
		height: 4px;
		border-radius: 2px;
	}

	.theme-bento .habit-rate-fill {
		background: #6366f1;
		border-radius: 2px;
	}

	.theme-bento .habit-log-btn {
		background: rgba(134, 203, 146, 0.1);
		border: 1px solid rgba(134, 203, 146, 0.25);
		color: #86cb92;
		border-radius: 8px;
		font-weight: 500;
		font-size: 0.72rem;
	}

	.theme-bento .habit-log-btn:hover {
		background: rgba(134, 203, 146, 0.18);
	}

	.theme-bento .habit-log-btn.logged {
		background: #252528;
		border-color: #3a3a3c;
		color: #666;
	}

	.theme-bento .habit-done { opacity: 0.6; }

	/* ======================================================
	   THEME: BENTO ALT — Dark with Periwinkle Accent
	   Feels like: a cooler, moodier Bento variant
	   ====================================================== */
	.theme-bento-alt {
		background: #0f0f12;
		color: #e5e5e7;
		font-family: -apple-system, BlinkMacSystemFont, 'SF Pro Display', 'Segoe UI', Roboto, sans-serif;
	}

	.theme-bento-alt .view-title {
		font-weight: 600;
		font-size: 1.5rem;
		color: #e5e5e7;
		letter-spacing: -0.01em;
	}

	.theme-bento-alt .view-count,
	.theme-bento-alt .view-subtitle {
		color: #9b7ede;
		font-weight: 500;
		font-size: 0.8rem;
	}

	.theme-bento-alt .view-header {
		margin-bottom: 1.25rem;
	}

	.theme-bento-alt .add-input-bar {
		background: #1a1a22;
		border: 1px solid #2e2d4d;
		border-radius: 10px;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.3);
	}

	.theme-bento-alt .add-icon { color: #9b7ede; }
	.theme-bento-alt .add-placeholder { color: #555; }

	.theme-bento-alt .task-item {
		background: #1a1a22;
		border: 1px solid #2e2d4d;
		border-radius: 10px;
		padding: 0.8rem 1rem;
		margin-bottom: 0.4rem;
		border-bottom: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
	}

	.theme-bento-alt .checkbox-inner {
		border: 2px solid #9b7ede;
		border-radius: 50%;
		width: 18px;
		height: 18px;
	}

	.theme-bento-alt .task-title {
		color: #e5e5e7;
		font-weight: 500;
		font-size: 0.93rem;
	}

	.theme-bento-alt .badge {
		background: #22222e;
		color: #999;
		border-radius: 5px;
		font-size: 0.6rem;
		font-weight: 500;
		padding: 0.15rem 0.45rem;
	}

	.theme-bento-alt .badge-priority {
		background: rgba(155, 126, 222, 0.12);
		color: #b9a4e8;
	}

	.theme-bento-alt .badge-urgent {
		background: rgba(182, 23, 75, 0.12);
		color: #e04878;
	}

	.theme-bento-alt .badge-high {
		background: rgba(155, 126, 222, 0.12);
		color: #9b7ede;
	}

	.theme-bento-alt .badge-energy {
		background: rgba(9, 129, 74, 0.1);
		color: #09814a;
	}

	.theme-bento-alt .badge-due {
		background: rgba(182, 23, 75, 0.1);
		color: #b6174b;
	}

	.theme-bento-alt .badge-tag {
		color: #777;
	}

	/* Bento Alt: suggestions */
	.theme-bento-alt .suggestion-item {
		background: #1a1a22;
		border: 1px solid #2e2d4d;
		border-radius: 10px;
		padding: 1rem;
		margin-bottom: 0.4rem;
		border-bottom: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
	}

	.theme-bento-alt .suggestion-rank {
		color: #9b7ede;
		font-weight: 700;
		font-size: 0.9rem;
	}

	.theme-bento-alt .suggestion-title {
		color: #e5e5e7;
		font-weight: 500;
		font-size: 0.95rem;
	}

	.theme-bento-alt .suggestion-title:hover { color: #b9a4e8; }

	.theme-bento-alt .suggestion-duration {
		color: #666;
		font-weight: 500;
	}

	.theme-bento-alt .score-track {
		background: #22222e;
		height: 6px;
		border-radius: 3px;
	}

	.theme-bento-alt .score-fill {
		background: #9b7ede;
		border-radius: 3px;
	}

	.theme-bento-alt .score-value {
		color: #b9a4e8;
		font-weight: 600;
		font-size: 0.7rem;
	}

	.theme-bento-alt .score-breakdown {
		border-top: 1px solid #2e2d4d;
	}

	.theme-bento-alt .breakdown-label {
		color: #555;
		font-weight: 500;
	}

	.theme-bento-alt .factor-name {
		color: #888;
		font-size: 0.68rem;
	}

	.theme-bento-alt .factor-track {
		background: #22222e;
		height: 4px;
		border-radius: 2px;
	}

	.theme-bento-alt .factor-fill {
		background: #09814a;
		border-radius: 2px;
	}

	.theme-bento-alt .factor-value { color: #09814a; font-weight: 500; }

	/* Bento Alt: habits */
	.theme-bento-alt .habit-item {
		background: #1a1a22;
		border: 1px solid #2e2d4d;
		border-radius: 10px;
		padding: 0.85rem 1rem;
		margin-bottom: 0.4rem;
		border-bottom: none;
		box-shadow: 0 1px 3px rgba(0, 0, 0, 0.2);
	}

	.theme-bento-alt .habit-status-indicator {
		border-radius: 50%;
		width: 24px;
		height: 24px;
	}

	.theme-bento-alt .habit-check { color: #fff; }

	.theme-bento-alt .habit-item:has(.habit-check) .habit-status-indicator {
		background: #09814a;
	}

	.theme-bento-alt .habit-item:has(.habit-warning) .habit-status-indicator {
		background: #b6174b;
	}

	.theme-bento-alt .habit-warning { color: #fff; }

	.theme-bento-alt .habit-item:has(.habit-empty) .habit-status-indicator {
		background: #22222e;
		border: 1px solid #3a3a4c;
	}

	.theme-bento-alt .habit-empty { color: #555; }

	.theme-bento-alt .habit-title {
		color: #e5e5e7;
		font-weight: 500;
	}

	.theme-bento-alt .habit-freq { color: #555; font-weight: 500; }

	.theme-bento-alt .habit-stat {
		color: #888;
		font-size: 0.7rem;
	}

	.theme-bento-alt .habit-stat-sep { color: #333; }

	.theme-bento-alt .habit-rate-bar {
		background: #22222e;
		height: 4px;
		border-radius: 2px;
	}

	.theme-bento-alt .habit-rate-fill {
		background: #09814a;
		border-radius: 2px;
	}

	.theme-bento-alt .habit-log-btn {
		background: rgba(9, 129, 74, 0.1);
		border: 1px solid rgba(9, 129, 74, 0.25);
		color: #09814a;
		border-radius: 8px;
		font-weight: 500;
		font-size: 0.72rem;
	}

	.theme-bento-alt .habit-log-btn:hover {
		background: rgba(9, 129, 74, 0.18);
	}

	.theme-bento-alt .habit-log-btn.logged {
		background: #22222e;
		border-color: #3a3a4c;
		color: #666;
	}

	.theme-bento-alt .habit-done { opacity: 0.6; }

	/* ======================================================
	   THEME: CLAY — Warm Organic
	   Feels like: cozy indie productivity app
	   ====================================================== */
	.theme-clay {
		background: #f5f0e8;
		color: #2d3a4a;
		font-family: 'Nunito', -apple-system, sans-serif;
	}

	.theme-clay .view-title {
		font-family: 'Nunito', sans-serif;
		font-weight: 800;
		font-size: 1.6rem;
		color: #2d3a4a;
	}

	.theme-clay .view-count,
	.theme-clay .view-subtitle {
		font-weight: 600;
		color: #6b8f71;
		font-size: 0.8rem;
	}

	.theme-clay .view-header {
		margin-bottom: 1.25rem;
	}

	.theme-clay .add-input-bar {
		background: #fff;
		border: 2px solid #e5ddd0;
		border-radius: 14px;
		padding: 0.85rem 1.1rem;
		box-shadow: 0 2px 8px rgba(44, 58, 74, 0.06);
	}

	.theme-clay .add-icon { color: #c45b3a; font-weight: 800; }
	.theme-clay .add-placeholder { color: #b8ad9e; }

	.theme-clay .task-item {
		background: #fff;
		border-radius: 14px;
		padding: 0.9rem 1rem;
		margin-bottom: 0.5rem;
		border-bottom: none;
		box-shadow: 0 2px 8px rgba(44, 58, 74, 0.06);
	}

	.theme-clay .checkbox-inner {
		border: 2px solid #c45b3a;
		border-radius: 50%;
		width: 20px;
		height: 20px;
	}

	.theme-clay .task-title {
		font-weight: 600;
		color: #2d3a4a;
		font-size: 0.95rem;
	}

	.theme-clay .badge {
		background: #f0ebe3;
		color: #8a7e6f;
		border-radius: 10px;
		font-size: 0.6rem;
		font-weight: 700;
		padding: 0.2rem 0.55rem;
	}

	.theme-clay .badge-priority {
		background: #fce8e2;
		color: #c45b3a;
	}

	.theme-clay .badge-urgent {
		background: #c45b3a;
		color: #fff;
	}

	.theme-clay .badge-energy {
		background: #e5f0e7;
		color: #6b8f71;
	}

	.theme-clay .badge-due {
		background: #fce8e2;
		color: #c45b3a;
	}

	.theme-clay .badge-tag {
		background: #e8e3db;
		color: #7a6f60;
	}

	/* Clay: suggestions */
	.theme-clay .suggestion-item {
		background: #fff;
		border-radius: 14px;
		padding: 1rem;
		margin-bottom: 0.5rem;
		border-bottom: none;
		box-shadow: 0 2px 8px rgba(44, 58, 74, 0.06);
	}

	.theme-clay .suggestion-rank {
		background: #c45b3a;
		color: #fff;
		border-radius: 50%;
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.7rem;
		font-weight: 800;
		min-width: 28px;
	}

	.theme-clay .suggestion-title {
		font-family: 'Nunito', sans-serif;
		color: #2d3a4a;
		font-weight: 700;
		font-size: 1rem;
	}

	.theme-clay .suggestion-title:hover { color: #c45b3a; }

	.theme-clay .suggestion-duration {
		color: #b8ad9e;
		font-weight: 600;
	}

	.theme-clay .score-track {
		background: #e5ddd0;
		height: 8px;
		border-radius: 4px;
	}

	.theme-clay .score-fill {
		background: linear-gradient(90deg, #c45b3a, #6b8f71);
		border-radius: 4px;
	}

	.theme-clay .score-value {
		color: #8a7e6f;
		font-weight: 700;
		font-size: 0.7rem;
	}

	.theme-clay .score-breakdown {
		border-top: 2px solid #f0ebe3;
	}

	.theme-clay .breakdown-label {
		color: #b8ad9e;
		font-weight: 700;
	}

	.theme-clay .factor-name { color: #8a7e6f; font-weight: 600; }

	.theme-clay .factor-track {
		background: #e5ddd0;
		height: 6px;
		border-radius: 3px;
	}

	.theme-clay .factor-fill {
		background: #6b8f71;
		border-radius: 3px;
	}

	.theme-clay .factor-value { color: #8a7e6f; font-weight: 600; }

	/* Clay: habits */
	.theme-clay .habit-item {
		background: #fff;
		border-radius: 14px;
		padding: 0.9rem 1rem;
		margin-bottom: 0.5rem;
		border-bottom: none;
		box-shadow: 0 2px 8px rgba(44, 58, 74, 0.06);
	}

	.theme-clay .habit-status-indicator {
		border-radius: 50%;
		width: 26px;
		height: 26px;
	}

	.theme-clay .habit-check {
		color: #fff;
	}

	.theme-clay .habit-item:has(.habit-check) .habit-status-indicator {
		background: #6b8f71;
	}

	.theme-clay .habit-item:has(.habit-warning) .habit-status-indicator {
		background: #f0c27a;
	}

	.theme-clay .habit-warning { color: #fff; }

	.theme-clay .habit-item:has(.habit-empty) .habit-status-indicator {
		background: #e5ddd0;
	}

	.theme-clay .habit-empty { color: #b8ad9e; }

	.theme-clay .habit-title {
		font-weight: 700;
		color: #2d3a4a;
	}

	.theme-clay .habit-freq {
		color: #b8ad9e;
		font-weight: 700;
		font-size: 0.6rem;
	}

	.theme-clay .habit-stat {
		color: #8a7e6f;
		font-weight: 600;
		font-size: 0.7rem;
	}

	.theme-clay .habit-stat-sep { color: #e5ddd0; }

	.theme-clay .habit-rate-bar {
		background: #e5ddd0;
		height: 6px;
		border-radius: 3px;
	}

	.theme-clay .habit-rate-fill {
		background: #6b8f71;
		border-radius: 3px;
	}

	.theme-clay .habit-log-btn {
		background: #6b8f71;
		border: none;
		color: #fff;
		border-radius: 12px;
		font-family: 'Nunito', sans-serif;
		font-weight: 700;
		font-size: 0.7rem;
		padding: 0.4rem 1rem;
	}

	.theme-clay .habit-log-btn:hover {
		background: #5a7d60;
	}

	.theme-clay .habit-log-btn.logged {
		background: #e5ddd0;
		color: #8a7e6f;
	}

	.theme-clay .habit-done { opacity: 0.7; }

	/* ======================================================
	   THEME: CLAY ALT — Rosewood Organic
	   Feels like: warm blush productivity with a romantic edge
	   ====================================================== */
	.theme-clay-alt {
		background: #fdecef;
		color: #0f110c;
		font-family: 'Nunito', -apple-system, sans-serif;
	}

	.theme-clay-alt .view-title {
		font-family: 'Nunito', sans-serif;
		font-weight: 800;
		font-size: 1.6rem;
		color: #0f110c;
	}

	.theme-clay-alt .view-count,
	.theme-clay-alt .view-subtitle {
		font-weight: 600;
		color: #9d6381;
		font-size: 0.8rem;
	}

	.theme-clay-alt .view-header {
		margin-bottom: 1.25rem;
	}

	.theme-clay-alt .add-input-bar {
		background: #fff5f6;
		border: 2px solid #e8d0d6;
		border-radius: 14px;
		padding: 0.85rem 1.1rem;
		box-shadow: 0 2px 8px rgba(157, 99, 129, 0.08);
	}

	.theme-clay-alt .add-icon { color: #9d6381; font-weight: 800; }
	.theme-clay-alt .add-placeholder { color: #84828f; }

	.theme-clay-alt .task-item {
		background: #fff5f6;
		border-radius: 14px;
		padding: 0.9rem 1rem;
		margin-bottom: 0.5rem;
		border-bottom: none;
		box-shadow: 0 2px 8px rgba(157, 99, 129, 0.08);
	}

	.theme-clay-alt .checkbox-inner {
		border: 2px solid #9d6381;
		border-radius: 50%;
		width: 20px;
		height: 20px;
	}

	.theme-clay-alt .task-title {
		font-weight: 600;
		color: #0f110c;
		font-size: 0.95rem;
	}

	.theme-clay-alt .badge {
		background: #f3e4e8;
		color: #84828f;
		border-radius: 10px;
		font-size: 0.6rem;
		font-weight: 700;
		padding: 0.2rem 0.55rem;
	}

	.theme-clay-alt .badge-priority {
		background: #f5dde3;
		color: #9d6381;
	}

	.theme-clay-alt .badge-urgent {
		background: #9d6381;
		color: #fff;
	}

	.theme-clay-alt .badge-energy {
		background: #f0e4ec;
		color: #612940;
	}

	.theme-clay-alt .badge-due {
		background: #f5dde3;
		color: #9d6381;
	}

	.theme-clay-alt .badge-tag {
		background: #ede0e4;
		color: #84828f;
	}

	/* Clay Alt: suggestions */
	.theme-clay-alt .suggestion-item {
		background: #fff5f6;
		border-radius: 14px;
		padding: 1rem;
		margin-bottom: 0.5rem;
		border-bottom: none;
		box-shadow: 0 2px 8px rgba(157, 99, 129, 0.08);
	}

	.theme-clay-alt .suggestion-rank {
		background: #9d6381;
		color: #fff;
		border-radius: 50%;
		width: 28px;
		height: 28px;
		display: flex;
		align-items: center;
		justify-content: center;
		font-size: 0.7rem;
		font-weight: 800;
		min-width: 28px;
	}

	.theme-clay-alt .suggestion-title {
		font-family: 'Nunito', sans-serif;
		color: #0f110c;
		font-weight: 700;
		font-size: 1rem;
	}

	.theme-clay-alt .suggestion-title:hover { color: #9d6381; }

	.theme-clay-alt .suggestion-duration {
		color: #84828f;
		font-weight: 600;
	}

	.theme-clay-alt .score-track {
		background: #e8d0d6;
		height: 8px;
		border-radius: 4px;
	}

	.theme-clay-alt .score-fill {
		background: #9d6381;
		border-radius: 4px;
	}

	.theme-clay-alt .score-value {
		color: #9d6381;
		font-weight: 700;
		font-size: 0.7rem;
	}

	.theme-clay-alt .score-breakdown {
		border-top: 2px solid #f3e4e8;
	}

	.theme-clay-alt .breakdown-label {
		color: #84828f;
		font-weight: 700;
	}

	.theme-clay-alt .factor-name { color: #84828f; font-weight: 600; }

	.theme-clay-alt .factor-track {
		background: #e8d0d6;
		height: 6px;
		border-radius: 3px;
	}

	.theme-clay-alt .factor-fill {
		background: #612940;
		border-radius: 3px;
	}

	.theme-clay-alt .factor-value { color: #612940; font-weight: 600; }

	/* Clay Alt: habits */
	.theme-clay-alt .habit-item {
		background: #fff5f6;
		border-radius: 14px;
		padding: 0.9rem 1rem;
		margin-bottom: 0.5rem;
		border-bottom: none;
		box-shadow: 0 2px 8px rgba(157, 99, 129, 0.08);
	}

	.theme-clay-alt .habit-status-indicator {
		border-radius: 50%;
		width: 26px;
		height: 26px;
	}

	.theme-clay-alt .habit-check {
		color: #fff;
	}

	.theme-clay-alt .habit-item:has(.habit-check) .habit-status-indicator {
		background: #612940;
	}

	.theme-clay-alt .habit-item:has(.habit-warning) .habit-status-indicator {
		background: #9d6381;
	}

	.theme-clay-alt .habit-warning { color: #fff; }

	.theme-clay-alt .habit-item:has(.habit-empty) .habit-status-indicator {
		background: #e8d0d6;
	}

	.theme-clay-alt .habit-empty { color: #84828f; }

	.theme-clay-alt .habit-title {
		font-weight: 700;
		color: #0f110c;
	}

	.theme-clay-alt .habit-freq {
		color: #84828f;
		font-weight: 700;
		font-size: 0.6rem;
	}

	.theme-clay-alt .habit-stat {
		color: #84828f;
		font-weight: 600;
		font-size: 0.7rem;
	}

	.theme-clay-alt .habit-stat-sep { color: #e8d0d6; }

	.theme-clay-alt .habit-rate-bar {
		background: #e8d0d6;
		height: 6px;
		border-radius: 3px;
	}

	.theme-clay-alt .habit-rate-fill {
		background: #612940;
		border-radius: 3px;
	}

	.theme-clay-alt .habit-log-btn {
		background: #612940;
		border: none;
		color: #fff;
		border-radius: 12px;
		font-family: 'Nunito', sans-serif;
		font-weight: 700;
		font-size: 0.7rem;
		padding: 0.4rem 1rem;
	}

	.theme-clay-alt .habit-log-btn:hover {
		background: #4e1f33;
	}

	.theme-clay-alt .habit-log-btn.logged {
		background: #e8d0d6;
		color: #84828f;
	}

	.theme-clay-alt .habit-done { opacity: 0.7; }
</style>
