<script lang="ts">
	type Priority = 'urgent' | 'high' | 'medium' | 'low';
	type Energy = 'high' | 'medium' | 'low';
	type ThemeName = 'ink' | 'neon' | 'clay' | 'blueprint' | 'bento';
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
		{ name: 'ink', label: 'Ink' },
		{ name: 'neon', label: 'Neon' },
		{ name: 'clay', label: 'Clay' },
		{ name: 'blueprint', label: 'Blueprint' },
		{ name: 'bento', label: 'Bento' },
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

	.theme-btn[data-theme="ink"].active { background: #c0392b; border-color: #c0392b; }
	.theme-btn[data-theme="neon"].active { background: #00fff2; border-color: #00fff2; color: #000; }
	.theme-btn[data-theme="clay"].active { background: #c45b3a; border-color: #c45b3a; }
	.theme-btn[data-theme="blueprint"].active { background: #4a90d9; border-color: #4a90d9; }

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
	   THEME: INK — Editorial Minimalism
	   Feels like: a beautifully typeset Moleskine to-do list
	   ====================================================== */
	.theme-ink {
		background: #fafaf8;
		color: #1a1a1a;
		font-family: 'Playfair Display', Georgia, 'Times New Roman', serif;
	}

	.theme-ink .themed-content { max-width: 620px; }

	.theme-ink .view-title {
		font-family: 'Playfair Display', Georgia, serif;
		font-weight: 700;
		font-size: 2rem;
		letter-spacing: -0.02em;
		color: #1a1a1a;
	}

	.theme-ink .view-count,
	.theme-ink .view-subtitle {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.1em;
		color: #999;
	}

	.theme-ink .view-header {
		padding-bottom: 1rem;
		border-bottom: 0.5px solid #d0cfc8;
		margin-bottom: 2rem;
	}

	.theme-ink .add-input-bar {
		background: transparent;
		border: none;
		border-bottom: 0.5px solid #d0cfc8;
		border-radius: 0;
		padding: 0.8rem 0;
		margin-bottom: 1.5rem;
	}

	.theme-ink .add-icon {
		color: #c0392b;
		font-family: 'Playfair Display', serif;
	}

	.theme-ink .add-placeholder {
		font-family: 'Playfair Display', serif;
		font-style: italic;
		color: #b0afa8;
	}

	.theme-ink .task-item {
		border-bottom: 0.5px solid #e8e7e0;
		padding: 1.1rem 0;
	}

	.theme-ink .checkbox-inner {
		border: 1.5px solid #c0392b;
		border-radius: 2px;
		width: 16px;
		height: 16px;
	}

	.theme-ink .task-title {
		font-family: 'Playfair Display', serif;
		font-size: 1rem;
		font-weight: 400;
		color: #1a1a1a;
		line-height: 1.4;
	}

	.theme-ink .badge {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.55rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		font-variant: small-caps;
		background: transparent;
		color: #888;
		padding: 0;
		margin-right: 0.6rem;
	}

	.theme-ink .badge-priority { color: #c0392b; }
	.theme-ink .badge-urgent { color: #c0392b; font-weight: 600; }
	.theme-ink .badge-due { color: #c0392b; }
	.theme-ink .badge-tag { color: #666; }

	/* Ink: suggestions */
	.theme-ink .suggestion-item {
		border-bottom: 0.5px solid #e8e7e0;
		padding: 1.2rem 0;
	}

	.theme-ink .suggestion-rank {
		font-family: 'Playfair Display', serif;
		font-size: 1.1rem;
		font-weight: 700;
		color: #c0392b;
	}

	.theme-ink .suggestion-title {
		font-family: 'Playfair Display', serif;
		color: #1a1a1a;
		font-size: 1.05rem;
	}

	.theme-ink .suggestion-title:hover { color: #c0392b; }

	.theme-ink .suggestion-duration {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.65rem;
		color: #999;
	}

	.theme-ink .score-track {
		background: #e8e7e0;
		height: 1px;
	}

	.theme-ink .score-fill {
		background: #c0392b;
		height: 1px;
	}

	.theme-ink .score-value {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.6rem;
		color: #999;
	}

	.theme-ink .score-breakdown {
		border-top: 0.5px solid #e8e7e0;
	}

	.theme-ink .breakdown-label {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.55rem;
		letter-spacing: 0.1em;
		color: #b0afa8;
	}

	.theme-ink .factor-name {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.6rem;
		color: #888;
	}

	.theme-ink .factor-track {
		background: #e8e7e0;
		height: 1px;
	}

	.theme-ink .factor-fill {
		background: #c0392b;
		height: 1px;
	}

	.theme-ink .factor-value {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.55rem;
		color: #b0afa8;
	}

	/* Ink: habits */
	.theme-ink .habit-item {
		border-bottom: 0.5px solid #e8e7e0;
		padding: 1rem 0;
	}

	.theme-ink .habit-status-indicator {
		border: 1.5px solid #c0392b;
		border-radius: 2px;
		width: 22px;
		height: 22px;
	}

	.theme-ink .habit-check { color: #c0392b; }
	.theme-ink .habit-warning { color: #c0392b; }
	.theme-ink .habit-empty { color: #d0cfc8; }

	.theme-ink .habit-title {
		font-family: 'Playfair Display', serif;
		font-weight: 400;
	}

	.theme-ink .habit-freq {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.55rem;
		letter-spacing: 0.1em;
		color: #b0afa8;
	}

	.theme-ink .habit-stat {
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.6rem;
		color: #999;
	}

	.theme-ink .habit-stat-sep { color: #d0cfc8; }

	.theme-ink .habit-rate-bar {
		background: #e8e7e0;
		height: 1px;
	}

	.theme-ink .habit-rate-fill {
		background: #c0392b;
		height: 1px;
	}

	.theme-ink .habit-log-btn {
		font-family: 'JetBrains Mono', monospace;
		background: transparent;
		border: 1px solid #d0cfc8;
		color: #888;
		border-radius: 2px;
		font-size: 0.6rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		padding: 0.35rem 0.7rem;
	}

	.theme-ink .habit-log-btn:hover { border-color: #c0392b; color: #c0392b; }
	.theme-ink .habit-log-btn.logged { color: #c0392b; border-color: #c0392b; }

	.theme-ink .habit-done { opacity: 0.55; }

	/* ======================================================
	   THEME: NEON — Cyberpunk Dashboard
	   Feels like: hacker terminal meets Tron
	   ====================================================== */
	.theme-neon {
		background: #000;
		color: #00fff2;
		font-family: 'JetBrains Mono', 'Fira Code', monospace;
		background-image: repeating-linear-gradient(
			0deg,
			transparent,
			transparent 2px,
			rgba(0, 255, 242, 0.015) 2px,
			rgba(0, 255, 242, 0.015) 4px
		);
	}

	.theme-neon .themed-content { max-width: 750px; }

	.theme-neon .view-title {
		font-family: 'JetBrains Mono', monospace;
		font-weight: 300;
		font-size: 1.3rem;
		text-transform: uppercase;
		letter-spacing: 0.2em;
		color: #00fff2;
		text-shadow: 0 0 10px rgba(0, 255, 242, 0.5), 0 0 30px rgba(0, 255, 242, 0.2);
	}

	.theme-neon .view-count,
	.theme-neon .view-subtitle {
		color: #ff00aa;
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.15em;
		text-shadow: 0 0 8px rgba(255, 0, 170, 0.4);
	}

	.theme-neon .view-header {
		border-bottom: 1px solid rgba(0, 255, 242, 0.15);
		padding-bottom: 0.75rem;
		margin-bottom: 1.5rem;
	}

	.theme-neon .add-input-bar {
		background: rgba(0, 255, 242, 0.03);
		border: 1px solid rgba(0, 255, 242, 0.2);
		border-radius: 2px;
		box-shadow: 0 0 8px rgba(0, 255, 242, 0.1), inset 0 0 8px rgba(0, 255, 242, 0.03);
	}

	.theme-neon .add-icon { color: #00fff2; }
	.theme-neon .add-placeholder { color: rgba(0, 255, 242, 0.3); }

	.theme-neon .task-item {
		border-bottom: 1px solid rgba(0, 255, 242, 0.08);
		padding: 0.7rem 0;
	}

	.theme-neon .checkbox-inner {
		border: 1px solid #00fff2;
		border-radius: 2px;
		width: 16px;
		height: 16px;
		box-shadow: 0 0 4px rgba(0, 255, 242, 0.3);
	}

	.theme-neon .task-title {
		color: #e0e0e0;
		font-size: 0.85rem;
		font-weight: 400;
	}

	.theme-neon .badge {
		font-size: 0.55rem;
		text-transform: uppercase;
		letter-spacing: 0.06em;
		border: 1px solid rgba(0, 255, 242, 0.25);
		background: rgba(0, 255, 242, 0.05);
		color: #00fff2;
		border-radius: 2px;
		padding: 0.12rem 0.4rem;
	}

	.theme-neon .badge-priority {
		border-color: rgba(255, 0, 170, 0.4);
		background: rgba(255, 0, 170, 0.08);
		color: #ff00aa;
	}

	.theme-neon .badge-urgent {
		border-color: rgba(255, 0, 170, 0.6);
		background: rgba(255, 0, 170, 0.15);
		color: #ff00aa;
		box-shadow: 0 0 6px rgba(255, 0, 170, 0.3);
	}

	.theme-neon .badge-due {
		border-color: rgba(255, 0, 170, 0.3);
		color: #ff00aa;
	}

	.theme-neon .badge-energy {
		border-color: rgba(0, 255, 136, 0.3);
		background: rgba(0, 255, 136, 0.05);
		color: #00ff88;
	}

	.theme-neon .badge-tag {
		color: rgba(0, 255, 242, 0.6);
		border-color: rgba(0, 255, 242, 0.15);
	}

	/* Neon: suggestions */
	.theme-neon .suggestion-item {
		border-bottom: 1px solid rgba(0, 255, 242, 0.08);
		padding: 0.8rem 0;
	}

	.theme-neon .suggestion-rank {
		color: #ff00aa;
		font-weight: 600;
		text-shadow: 0 0 8px rgba(255, 0, 170, 0.4);
	}

	.theme-neon .suggestion-title {
		color: #e0e0e0;
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.85rem;
		font-weight: 400;
	}

	.theme-neon .suggestion-title:hover {
		color: #00fff2;
		text-shadow: 0 0 6px rgba(0, 255, 242, 0.4);
	}

	.theme-neon .suggestion-duration {
		color: rgba(0, 255, 242, 0.5);
		font-size: 0.65rem;
	}

	/* Neon: LED-style segmented score bars */
	.theme-neon .score-track {
		background: rgba(0, 255, 242, 0.06);
		height: 8px;
		border-radius: 1px;
		border: 1px solid rgba(0, 255, 242, 0.15);
	}

	.theme-neon .score-fill {
		background: repeating-linear-gradient(
			90deg,
			#00fff2 0px,
			#00fff2 6px,
			transparent 6px,
			transparent 8px
		);
		border-radius: 0;
		box-shadow: 0 0 8px rgba(0, 255, 242, 0.4);
	}

	.theme-neon .score-value {
		color: #00fff2;
		font-size: 0.65rem;
		text-shadow: 0 0 4px rgba(0, 255, 242, 0.3);
	}

	.theme-neon .score-breakdown {
		border-top: 1px solid rgba(0, 255, 242, 0.1);
	}

	.theme-neon .breakdown-label {
		color: #ff00aa;
		letter-spacing: 0.15em;
		text-shadow: 0 0 4px rgba(255, 0, 170, 0.3);
	}

	.theme-neon .factor-name { color: rgba(0, 255, 242, 0.5); font-size: 0.6rem; }

	.theme-neon .factor-track {
		background: rgba(0, 255, 136, 0.06);
		height: 6px;
		border: 1px solid rgba(0, 255, 136, 0.15);
		border-radius: 1px;
	}

	.theme-neon .factor-fill {
		background: repeating-linear-gradient(
			90deg,
			#00ff88 0px,
			#00ff88 4px,
			transparent 4px,
			transparent 6px
		);
		border-radius: 0;
		box-shadow: 0 0 6px rgba(0, 255, 136, 0.4);
	}

	.theme-neon .factor-value { color: #00ff88; }

	/* Neon: habits */
	.theme-neon .habit-item {
		border-bottom: 1px solid rgba(0, 255, 242, 0.08);
	}

	.theme-neon .habit-status-indicator {
		border: 1px solid rgba(0, 255, 242, 0.4);
		border-radius: 2px;
		width: 22px;
		height: 22px;
	}

	.theme-neon .habit-check {
		color: #00ff88;
		text-shadow: 0 0 6px rgba(0, 255, 136, 0.5);
	}

	.theme-neon .habit-warning {
		color: #ff00aa;
		text-shadow: 0 0 6px rgba(255, 0, 170, 0.5);
	}

	.theme-neon .habit-empty { color: rgba(0, 255, 242, 0.2); }

	.theme-neon .habit-title { color: #e0e0e0; font-size: 0.85rem; }

	.theme-neon .habit-freq {
		color: rgba(255, 0, 170, 0.5);
		font-size: 0.55rem;
		letter-spacing: 0.1em;
	}

	.theme-neon .habit-stat {
		color: rgba(0, 255, 242, 0.5);
		font-size: 0.6rem;
	}

	.theme-neon .habit-stat-sep { color: rgba(0, 255, 242, 0.15); }

	.theme-neon .habit-rate-bar {
		background: rgba(0, 255, 136, 0.06);
		height: 6px;
		border: 1px solid rgba(0, 255, 136, 0.15);
		border-radius: 1px;
	}

	.theme-neon .habit-rate-fill {
		background: repeating-linear-gradient(
			90deg,
			#00ff88 0px,
			#00ff88 6px,
			transparent 6px,
			transparent 8px
		);
		border-radius: 0;
		box-shadow: 0 0 6px rgba(0, 255, 136, 0.4);
	}

	.theme-neon .habit-log-btn {
		background: rgba(0, 255, 242, 0.05);
		border: 1px solid rgba(0, 255, 242, 0.3);
		color: #00fff2;
		border-radius: 2px;
		font-family: 'JetBrains Mono', monospace;
		font-size: 0.6rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		box-shadow: 0 0 6px rgba(0, 255, 242, 0.15);
	}

	.theme-neon .habit-log-btn:hover {
		background: rgba(0, 255, 242, 0.1);
		box-shadow: 0 0 12px rgba(0, 255, 242, 0.3);
	}

	.theme-neon .habit-log-btn.logged {
		border-color: rgba(0, 255, 136, 0.4);
		color: #00ff88;
		box-shadow: 0 0 6px rgba(0, 255, 136, 0.2);
	}

	.theme-neon .habit-done { opacity: 0.6; }

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
	   THEME: BLUEPRINT — Technical Precision
	   Feels like: mission control dashboard
	   ====================================================== */
	.theme-blueprint {
		background: #0a1628;
		color: #e8ecf1;
		font-family: 'IBM Plex Sans', -apple-system, sans-serif;
		background-image:
			linear-gradient(rgba(74, 144, 217, 0.04) 1px, transparent 1px),
			linear-gradient(90deg, rgba(74, 144, 217, 0.04) 1px, transparent 1px);
		background-size: 20px 20px;
	}

	.theme-blueprint .themed-content { max-width: 760px; }

	.theme-blueprint .view-title {
		font-family: 'IBM Plex Sans', sans-serif;
		font-weight: 600;
		font-size: 1.2rem;
		text-transform: uppercase;
		letter-spacing: 0.08em;
		color: #e8ecf1;
	}

	.theme-blueprint .view-count,
	.theme-blueprint .view-subtitle {
		font-size: 0.7rem;
		color: #4a90d9;
		font-weight: 500;
		letter-spacing: 0.04em;
	}

	.theme-blueprint .view-header {
		border-bottom: 1px solid rgba(74, 144, 217, 0.2);
		padding-bottom: 0.6rem;
		margin-bottom: 1.25rem;
	}

	.theme-blueprint .add-input-bar {
		background: rgba(74, 144, 217, 0.05);
		border: 1px solid rgba(74, 144, 217, 0.2);
		border-radius: 2px;
		padding: 0.55rem 0.8rem;
	}

	.theme-blueprint .add-icon { color: #4a90d9; font-weight: 500; }
	.theme-blueprint .add-placeholder { color: rgba(232, 236, 241, 0.25); font-size: 0.8rem; }

	.theme-blueprint .task-item {
		border-bottom: 1px solid rgba(74, 144, 217, 0.1);
		padding: 0.55rem 0;
	}

	.theme-blueprint .checkbox-inner {
		border: 1px solid #4a90d9;
		border-radius: 2px;
		width: 14px;
		height: 14px;
	}

	.theme-blueprint .task-title {
		color: #e8ecf1;
		font-size: 0.85rem;
		font-weight: 400;
	}

	.theme-blueprint .badge {
		font-size: 0.55rem;
		font-weight: 500;
		border: 1px solid rgba(74, 144, 217, 0.25);
		background: rgba(74, 144, 217, 0.08);
		color: #4a90d9;
		border-radius: 2px;
		padding: 0.1rem 0.35rem;
		letter-spacing: 0.02em;
	}

	.theme-blueprint .badge-priority {
		border-color: rgba(240, 160, 48, 0.4);
		background: rgba(240, 160, 48, 0.1);
		color: #f0a030;
	}

	.theme-blueprint .badge-urgent {
		border-color: rgba(240, 80, 48, 0.5);
		background: rgba(240, 80, 48, 0.12);
		color: #f05030;
	}

	.theme-blueprint .badge-energy {
		border-color: rgba(74, 144, 217, 0.3);
		color: rgba(74, 144, 217, 0.8);
	}

	.theme-blueprint .badge-due {
		border-color: rgba(240, 160, 48, 0.3);
		color: #f0a030;
	}

	.theme-blueprint .badge-tag {
		color: rgba(232, 236, 241, 0.5);
		border-color: rgba(232, 236, 241, 0.12);
		background: rgba(232, 236, 241, 0.04);
	}

	/* Blueprint: suggestions */
	.theme-blueprint .suggestion-item {
		border-bottom: 1px solid rgba(74, 144, 217, 0.1);
		padding: 0.65rem 0;
	}

	.theme-blueprint .suggestion-rank {
		font-family: 'JetBrains Mono', monospace;
		color: #4a90d9;
		font-weight: 500;
		font-size: 0.75rem;
	}

	.theme-blueprint .suggestion-title {
		font-family: 'IBM Plex Sans', sans-serif;
		color: #e8ecf1;
		font-size: 0.88rem;
		font-weight: 500;
	}

	.theme-blueprint .suggestion-title:hover { color: #4a90d9; }

	.theme-blueprint .suggestion-duration {
		font-family: 'JetBrains Mono', monospace;
		color: rgba(74, 144, 217, 0.6);
		font-size: 0.6rem;
	}

	/* Blueprint: gauge-style score bars with numerical labels */
	.theme-blueprint .score-track {
		background: rgba(74, 144, 217, 0.1);
		height: 6px;
		border-radius: 1px;
		border: 1px solid rgba(74, 144, 217, 0.15);
	}

	.theme-blueprint .score-fill {
		background: #4a90d9;
		border-radius: 0;
	}

	.theme-blueprint .score-value {
		font-family: 'JetBrains Mono', monospace;
		color: #4a90d9;
		font-weight: 500;
		font-size: 0.65rem;
	}

	.theme-blueprint .score-breakdown {
		border-top: 1px solid rgba(74, 144, 217, 0.12);
	}

	.theme-blueprint .breakdown-label {
		color: rgba(74, 144, 217, 0.5);
		font-weight: 500;
		letter-spacing: 0.08em;
	}

	.theme-blueprint .factor-name {
		font-size: 0.6rem;
		color: rgba(232, 236, 241, 0.45);
		font-weight: 400;
	}

	.theme-blueprint .factor-track {
		background: rgba(74, 144, 217, 0.08);
		height: 4px;
		border-radius: 0;
		border: 1px solid rgba(74, 144, 217, 0.1);
	}

	.theme-blueprint .factor-fill {
		background: #4a90d9;
		border-radius: 0;
	}

	.theme-blueprint .factor-value {
		font-family: 'JetBrains Mono', monospace;
		color: rgba(74, 144, 217, 0.6);
	}

	/* Blueprint: habits */
	.theme-blueprint .habit-item {
		border-bottom: 1px solid rgba(74, 144, 217, 0.1);
		padding: 0.6rem 0;
	}

	.theme-blueprint .habit-status-indicator {
		border: 1px solid rgba(74, 144, 217, 0.3);
		border-radius: 2px;
		width: 20px;
		height: 20px;
	}

	.theme-blueprint .habit-check { color: #4a90d9; }
	.theme-blueprint .habit-warning { color: #f0a030; }
	.theme-blueprint .habit-empty { color: rgba(74, 144, 217, 0.2); }

	.theme-blueprint .habit-title {
		color: #e8ecf1;
		font-size: 0.85rem;
		font-weight: 500;
	}

	.theme-blueprint .habit-freq {
		font-family: 'JetBrains Mono', monospace;
		color: rgba(74, 144, 217, 0.4);
		font-size: 0.55rem;
		letter-spacing: 0.05em;
	}

	.theme-blueprint .habit-stat {
		font-family: 'JetBrains Mono', monospace;
		color: rgba(232, 236, 241, 0.45);
		font-size: 0.65rem;
	}

	.theme-blueprint .habit-stat-sep { color: rgba(74, 144, 217, 0.15); }

	.theme-blueprint .habit-rate-bar {
		background: rgba(74, 144, 217, 0.1);
		height: 4px;
		border-radius: 0;
		border: 1px solid rgba(74, 144, 217, 0.1);
	}

	.theme-blueprint .habit-rate-fill {
		background: #4a90d9;
		border-radius: 0;
	}

	.theme-blueprint .habit-log-btn {
		background: rgba(74, 144, 217, 0.1);
		border: 1px solid rgba(74, 144, 217, 0.3);
		color: #4a90d9;
		border-radius: 2px;
		font-family: 'IBM Plex Sans', sans-serif;
		font-weight: 500;
		font-size: 0.65rem;
		text-transform: uppercase;
		letter-spacing: 0.04em;
	}

	.theme-blueprint .habit-log-btn:hover {
		background: rgba(74, 144, 217, 0.2);
	}

	.theme-blueprint .habit-log-btn.logged {
		border-color: rgba(74, 144, 217, 0.2);
		color: rgba(74, 144, 217, 0.5);
	}

	.theme-blueprint .habit-done { opacity: 0.55; }

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
		background: rgba(16, 185, 129, 0.1);
		color: #10b981;
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

	/* Bento: smooth gradient score bars */
	.theme-bento .score-track {
		background: #252528;
		height: 6px;
		border-radius: 3px;
	}

	.theme-bento .score-fill {
		background: linear-gradient(90deg, #6366f1, #10b981);
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
		background: #10b981;
		border-radius: 2px;
	}

	.theme-bento .factor-value { color: #10b981; font-weight: 500; }

	/* Bento: habits — bento-box grid style */
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
		background: #10b981;
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
		background: linear-gradient(90deg, #6366f1, #10b981);
		border-radius: 2px;
	}

	.theme-bento .habit-log-btn {
		background: rgba(16, 185, 129, 0.1);
		border: 1px solid rgba(16, 185, 129, 0.25);
		color: #10b981;
		border-radius: 8px;
		font-weight: 500;
		font-size: 0.72rem;
	}

	.theme-bento .habit-log-btn:hover {
		background: rgba(16, 185, 129, 0.18);
	}

	.theme-bento .habit-log-btn.logged {
		background: #252528;
		border-color: #3a3a3c;
		color: #666;
	}

	.theme-bento .habit-done { opacity: 0.6; }
</style>
