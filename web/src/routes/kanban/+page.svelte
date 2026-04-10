<script lang="ts">
	import { onMount } from 'svelte';
	import { tasks, type TaskJSON, type UpdateTaskRequest } from '$lib/api';

	interface Column {
		status: string;
		label: string;
		items: TaskJSON[];
	}

	const statusOrder = ['pending', 'active', 'paused', 'waiting', 'done', 'cancelled'];
	const statusLabels: Record<string, string> = {
		pending: 'Pending', active: 'Active', paused: 'Paused',
		waiting: 'Waiting', done: 'Done', cancelled: 'Cancelled'
	};

	let allTasks: TaskJSON[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	let columns = $derived.by((): Column[] => {
		const grouped: Record<string, TaskJSON[]> = {};
		for (const t of allTasks) {
			if (t.type === 'routine' || t.type === 'habit') continue;
			const s = t.status || 'pending';
			if (!grouped[s]) grouped[s] = [];
			grouped[s].push(t);
		}
		return statusOrder
			.filter((s) => grouped[s] && grouped[s].length > 0)
			.map((s) => ({ status: s, label: statusLabels[s] || s, items: grouped[s] }));
	});

	async function loadTasks() {
		error = '';
		try {
			const res = await tasks.list();
			allTasks = res.items;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load tasks';
		} finally {
			loading = false;
		}
	}

	async function changeStatus(taskId: string, newStatus: string) {
		error = '';
		try {
			const changes: UpdateTaskRequest = { status: newStatus };
			await tasks.update(taskId, changes);
			await loadTasks();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to update status';
		}
	}

	function priorityClass(p?: string): string {
		if (!p || p === 'none') return '';
		return `priority-${p}`;
	}

	onMount(() => { loadTasks(); });
</script>

<div class="view">
	<h1>Kanban</h1>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if columns.length === 0}
		<p class="empty">No tasks found. Add some in the Inbox!</p>
	{:else}
		<div class="board">
			{#each columns as col}
				<div class="column">
					<div class="column-header">
						<span class="column-title">{col.label}</span>
						<span class="column-count">{col.items.length}</span>
					</div>
					<ul class="card-list">
						{#each col.items as task (task.id)}
							<li class="card">
								<span class="card-title">{task.title}</span>
								<div class="card-meta">
									{#if task.priority && task.priority !== 'none'}
										<span class="badge {priorityClass(task.priority)}">{task.priority}</span>
									{/if}
									{#if task.energy}
										<span class="badge energy">{task.energy}</span>
									{/if}
									{#if task.due_date}
										<span class="card-due">due {task.due_date}</span>
									{/if}
								</div>
								<select
									class="status-select"
									value={task.status}
									onchange={(e) => changeStatus(task.id, (e.target as HTMLSelectElement).value)}
								>
									{#each statusOrder as s}
										<option value={s} selected={task.status === s}>{statusLabels[s] || s}</option>
									{/each}
								</select>
							</li>
						{/each}
					</ul>
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.view { max-width: 100%; }
	h1 { margin-bottom: 1.5rem; font-size: 1.5rem; color: var(--text-primary); }

	.error { padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning); border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem; }
	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.board { display: flex; gap: 1rem; overflow-x: auto; padding-bottom: 1rem; }

	.column { min-width: 220px; flex: 1; max-width: 300px; display: flex; flex-direction: column; }

	.column-header { display: flex; justify-content: space-between; align-items: center; padding: 0.5rem 0.75rem; margin-bottom: 0.5rem; border-bottom: 2px solid var(--border-default); }
	.column-title { font-size: 0.85rem; font-weight: 600; color: var(--text-primary); text-transform: capitalize; }
	.column-count { font-size: 0.7rem; background: var(--bg-elevated); color: var(--text-secondary); padding: 0.1rem 0.4rem; border-radius: 10px; }

	.card-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; }

	.card { background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-card); padding: 0.65rem 0.75rem; box-shadow: var(--shadow-card); }

	.card-title { font-size: 0.85rem; font-weight: 500; color: var(--text-primary); display: block; margin-bottom: 0.3rem; }

	.card-meta { display: flex; flex-wrap: wrap; gap: 0.25rem; margin-bottom: 0.35rem; }

	.card-due { font-size: 0.6rem; color: var(--warning-text); }

	.status-select { width: 100%; padding: 0.25rem 0.4rem; background: var(--bg-base); border: 1px solid var(--border-default); border-radius: var(--radius-badge); color: var(--text-secondary); font-size: 0.7rem; cursor: pointer; }
	.status-select:focus { outline: none; border-color: var(--accent-primary); }
</style>
