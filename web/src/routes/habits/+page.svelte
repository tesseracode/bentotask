<script lang="ts">
	import { onMount } from 'svelte';
	import { habits, tasks, type TaskJSON, type HabitStats, type UpdateTaskRequest } from '$lib/api';

	interface HabitWithStats {
		habit: TaskJSON;
		stats: HabitStats | null;
		loadingStats: boolean;
	}

	let items: HabitWithStats[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	let newTitle = $state('');
	let newFreqType: 'daily' | 'weekly' = $state('daily');
	let newFreqTarget = $state(1);
	let newPriority = $state('');
	let newEnergy = $state('');
	let showAddForm = $state(false);

	let editingId: string | null = $state(null);
	let editTitle = $state('');
	let editPriority = $state('');
	let editEnergy = $state('');

	async function loadHabits() {
		error = '';
		try {
			const res = await habits.list();
			items = res.items.map((h) => ({ habit: h, stats: null, loadingStats: true }));
			const statsPromises = items.map(async (item, idx) => {
				try {
					const res = await habits.stats(item.habit.id);
					items[idx] = { ...items[idx], stats: res.stats, loadingStats: false };
				} catch {
					items[idx] = { ...items[idx], loadingStats: false };
				}
			});
			await Promise.all(statsPromises);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load habits';
		} finally {
			loading = false;
		}
	}

	async function addHabit() {
		if (!newTitle.trim()) return;
		error = '';
		try {
			await habits.create({
				title: newTitle.trim(), freq_type: newFreqType, freq_target: newFreqTarget,
				priority: newPriority || undefined, energy: newEnergy || undefined
			});
			newTitle = ''; newFreqType = 'daily'; newFreqTarget = 1; newPriority = ''; newEnergy = '';
			showAddForm = false; loading = true;
			await loadHabits();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to add habit';
		}
	}

	async function logHabit(id: string) {
		error = '';
		try {
			await habits.log(id);
			const idx = items.findIndex((i) => i.habit.id === id);
			if (idx >= 0) {
				items[idx] = { ...items[idx], loadingStats: true };
				try {
					const res = await habits.stats(id);
					items[idx] = { ...items[idx], stats: res.stats, loadingStats: false };
				} catch { items[idx] = { ...items[idx], loadingStats: false }; }
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to log habit';
		}
	}

	async function deleteHabit(id: string) {
		error = '';
		try {
			await tasks.delete(id);
			items = items.filter((i) => i.habit.id !== id);
			if (editingId === id) editingId = null;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete habit';
		}
	}

	function startEdit(habit: TaskJSON) {
		editingId = habit.id;
		editTitle = habit.title;
		editPriority = habit.priority ?? '';
		editEnergy = habit.energy ?? '';
	}

	function cancelEdit() { editingId = null; }

	async function saveEdit() {
		if (!editingId) return;
		error = '';
		const item = items.find((i) => i.habit.id === editingId);
		if (!item) return;
		const changes: UpdateTaskRequest = {};
		if (editTitle !== item.habit.title) changes.title = editTitle;
		if (editPriority !== (item.habit.priority ?? '')) changes.priority = editPriority;
		if (editEnergy !== (item.habit.energy ?? '')) changes.energy = editEnergy;
		try {
			await tasks.update(editingId, changes);
			editingId = null; loading = true;
			await loadHabits();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to update habit';
		}
	}

	function isCompletedToday(stats: HabitStats | null): boolean {
		if (!stats) return false;
		return stats.completed_today;
	}

	function formatRate(stats: HabitStats | null): string {
		if (!stats || stats.total_completions === 0) return '0%';
		return `${Math.round((stats.completion_rate || 0) * 100)}%`;
	}

	function handleKeydown(e: KeyboardEvent) { if (e.key === 'Enter') addHabit(); }

	onMount(() => { loadHabits(); });
</script>

<div class="view">
	<h1>Habits</h1>

	{#if !showAddForm}
		<div class="add-bar">
			<input type="text" placeholder="New habit name..." bind:value={newTitle} onkeydown={handleKeydown} />
			<button onclick={() => { if (newTitle.trim()) showAddForm = true; }} disabled={!newTitle.trim()}>+</button>
		</div>
	{:else}
		<div class="add-form">
			<label>Title <input type="text" bind:value={newTitle} onkeydown={handleKeydown} /></label>
			<div class="form-row">
				<label>
					Frequency
					<div class="toggle-group">
						<button class:active={newFreqType === 'daily'} onclick={() => newFreqType = 'daily'}>Daily</button>
						<button class:active={newFreqType === 'weekly'} onclick={() => newFreqType = 'weekly'}>Weekly</button>
					</div>
				</label>
				<label>Target <input type="number" min="1" max="7" bind:value={newFreqTarget} /></label>
				<label>Priority
					<select bind:value={newPriority}>
						<option value="">None</option><option value="low">Low</option><option value="medium">Medium</option>
						<option value="high">High</option><option value="urgent">Urgent</option>
					</select>
				</label>
				<label>Energy
					<select bind:value={newEnergy}>
						<option value="">None</option><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option>
					</select>
				</label>
			</div>
			<div class="form-actions">
				<button class="save-btn" onclick={addHabit} disabled={!newTitle.trim()}>Create Habit</button>
				<button class="cancel-btn" onclick={() => showAddForm = false}>Cancel</button>
			</div>
		</div>
	{/if}

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if items.length === 0}
		<p class="empty">No habits yet. Start building a streak!</p>
	{:else}
		<ul class="habit-list">
			{#each items as { habit, stats, loadingStats } (habit.id)}
				{@const done = isCompletedToday(stats)}
				<li class="habit-item" class:completed={done}>
					<div class="habit-status">
						{#if done}
							<span class="status-icon done-icon">✓</span>
						{:else if stats && stats.current_streak > 0}
							<span class="status-icon at-risk-icon">!</span>
						{:else}
							<span class="status-icon neutral-icon">○</span>
						{/if}
					</div>
					<div class="habit-body">
						{#if editingId === habit.id}
							<div class="edit-inline">
								<input type="text" bind:value={editTitle} class="edit-title" />
								<div class="edit-row">
									<select bind:value={editPriority}>
										<option value="">No priority</option><option value="low">Low</option><option value="medium">Medium</option>
										<option value="high">High</option><option value="urgent">Urgent</option>
									</select>
									<select bind:value={editEnergy}>
										<option value="">No energy</option><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option>
									</select>
								</div>
								<div class="edit-actions">
									<button class="save-btn" onclick={saveEdit}>Save</button>
									<button class="cancel-btn" onclick={cancelEdit}>Cancel</button>
								</div>
							</div>
						{:else}
							<span class="habit-title">{habit.title}</span>
							<div class="habit-meta">
								{#if habit.priority && habit.priority !== 'none'}
									<span class="badge priority-{habit.priority}">{habit.priority}</span>
								{/if}
								{#if habit.energy}
									<span class="badge energy">{habit.energy}</span>
								{/if}
							</div>
							{#if loadingStats}
								<div class="stats-row"><span class="stat-loading">Loading stats...</span></div>
							{:else if stats}
								<div class="stats-row">
									<span class="stat" title="Current streak">🔥 {stats.current_streak}</span>
									<span class="stat" title="Longest streak">🏆 {stats.longest_streak}</span>
									<span class="stat" title="Total completions">✅ {stats.total_completions}</span>
									<span class="stat" title="Completion rate ({stats.rate_period_days}d)">{formatRate(stats)}</span>
								</div>
								<div class="rate-bar"><div class="rate-fill" style="width: {Math.round((stats.completion_rate || 0) * 100)}%"></div></div>
							{/if}
						{/if}
					</div>
					<div class="habit-actions">
						<button class="log-btn" class:log-done={done} onclick={() => logHabit(habit.id)} title={done ? 'Already logged today' : 'Log completion'}>
							{done ? '✓' : 'Log'}
						</button>
						{#if editingId !== habit.id}
							<button class="action-btn" onclick={() => startEdit(habit)} title="Edit">✎</button>
							<button class="action-btn delete" onclick={() => deleteHabit(habit.id)} title="Delete">&times;</button>
						{/if}
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.view { max-width: 700px; }
	h1 { margin-bottom: 1.5rem; font-size: 1.5rem; color: var(--text-primary); }

	.add-bar { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; }

	.add-bar input {
		flex: 1; padding: 0.6rem 0.8rem; background: var(--bg-surface); border: 1px solid var(--border-default);
		border-radius: var(--radius-input); color: var(--text-primary); font-size: 0.9rem;
	}

	.add-bar input:focus { outline: none; border-color: var(--accent-primary); }

	.add-bar button {
		padding: 0.6rem 1rem; background: var(--accent-primary); color: var(--text-on-accent); border: none;
		border-radius: var(--radius-button); cursor: pointer; font-size: 1rem;
	}

	.add-bar button:disabled { opacity: 0.5; cursor: not-allowed; }

	.add-form {
		background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-card);
		padding: 1rem; margin-bottom: 1.5rem; display: flex; flex-direction: column; gap: 0.6rem;
	}

	.add-form label {
		display: flex; flex-direction: column; gap: 0.2rem; font-size: 0.75rem; color: var(--text-secondary);
	}

	.add-form input, .add-form select {
		padding: 0.4rem 0.6rem; background: var(--bg-base); border: 1px solid var(--border-default);
		border-radius: var(--radius-badge); color: var(--text-primary); font-size: 0.85rem;
	}

	.add-form input[type="number"] { width: 60px; }
	.add-form input:focus, .add-form select:focus { outline: none; border-color: var(--accent-primary); }

	.form-row { display: flex; gap: 0.6rem; flex-wrap: wrap; }
	.form-row label { flex: 1; min-width: 80px; }

	.toggle-group { display: flex; border: 1px solid var(--border-default); border-radius: var(--radius-badge); overflow: hidden; }

	.toggle-group button {
		padding: 0.35rem 0.6rem; background: var(--bg-surface); border: none; border-right: 1px solid var(--border-default);
		color: var(--text-secondary); cursor: pointer; font-size: 0.8rem;
	}

	.toggle-group button:last-child { border-right: none; }
	.toggle-group button.active { background: var(--accent-primary); color: var(--text-on-accent); }

	.form-actions, .edit-actions { display: flex; gap: 0.5rem; margin-top: 0.25rem; }

	.save-btn {
		padding: 0.4rem 0.8rem; background: var(--accent-primary); color: var(--text-on-accent); border: none;
		border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem;
	}

	.save-btn:disabled { opacity: 0.5; cursor: not-allowed; }

	.cancel-btn {
		padding: 0.4rem 0.8rem; background: var(--bg-elevated); color: var(--text-secondary);
		border: 1px solid var(--border-default); border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem;
	}

	.error {
		padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning);
		border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem;
	}

	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.habit-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; }

	.habit-item {
		display: flex; align-items: flex-start; gap: 0.75rem; padding: 0.85rem 1rem;
		background: var(--bg-surface); border: 1px solid var(--border-default);
		border-radius: var(--radius-card); box-shadow: var(--shadow-card);
	}

	.habit-item.completed { opacity: 0.6; }

	.habit-status { flex-shrink: 0; margin-top: 0.1rem; }

	.status-icon {
		display: block; width: 24px; height: 24px; text-align: center; line-height: 24px;
		font-size: 0.75rem; border-radius: 50%;
	}

	.done-icon { background: var(--success); color: var(--text-on-accent); }
	.at-risk-icon { background: var(--warning); color: var(--text-on-accent); }
	.neutral-icon { background: var(--bg-elevated); color: var(--text-tertiary); border: 1px solid var(--border-subtle); }

	.habit-body { flex: 1; min-width: 0; }
	.habit-title { display: block; font-size: 0.95rem; font-weight: 500; color: var(--text-primary); }
	.habit-meta { display: flex; gap: 0.35rem; margin-top: 0.25rem; }

	.stats-row { display: flex; gap: 0.75rem; margin-top: 0.35rem; }
	.stat { font-size: 0.75rem; color: var(--text-secondary); }
	.stat-loading { font-size: 0.7rem; color: var(--text-tertiary); }

	.rate-bar { height: 4px; background: var(--score-track); border-radius: 2px; margin-top: 0.3rem; overflow: hidden; }
	.rate-fill { height: 100%; background: var(--success); border-radius: 2px; transition: width 0.3s; }

	.edit-inline { display: flex; flex-direction: column; gap: 0.4rem; }

	.edit-title {
		padding: 0.35rem 0.5rem; background: var(--bg-base); border: 1px solid var(--border-default);
		border-radius: var(--radius-badge); color: var(--text-primary); font-size: 0.9rem;
	}

	.edit-title:focus { outline: none; border-color: var(--accent-primary); }

	.edit-row { display: flex; gap: 0.4rem; }

	.edit-row select {
		padding: 0.3rem 0.5rem; background: var(--bg-base); border: 1px solid var(--border-default);
		border-radius: var(--radius-badge); color: var(--text-secondary); font-size: 0.8rem;
	}

	.edit-row select:focus { outline: none; border-color: var(--accent-primary); }

	.habit-actions { display: flex; gap: 0.3rem; align-items: center; flex-shrink: 0; }

	.log-btn {
		padding: 0.4rem 0.8rem; background: var(--success-subtle); border: 1px solid var(--success-border);
		border-radius: var(--radius-button); color: var(--success); cursor: pointer; font-size: 0.8rem;
	}

	.log-btn:hover { background: var(--success-hover); }

	.log-btn.log-done { background: var(--bg-elevated); border-color: var(--border-subtle); color: var(--text-tertiary); }

	.action-btn {
		background: none; border: none; color: var(--text-tertiary); font-size: 1rem;
		cursor: pointer; padding: 0.2rem 0.3rem; line-height: 1;
	}

	.action-btn:hover { color: var(--text-secondary); }
	.action-btn.delete:hover { color: var(--warning-text); }
</style>
