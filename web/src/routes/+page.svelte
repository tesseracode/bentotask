<script lang="ts">
	import { onMount } from 'svelte';
	import { tasks, type TaskJSON } from '$lib/api';

	let taskList: TaskJSON[] = $state([]);
	let newTitle = $state('');
	let loading = $state(true);
	let error = $state('');

	async function loadTasks() {
		try {
			const res = await tasks.list({ status: 'pending' });
			taskList = res.items;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load tasks';
		} finally {
			loading = false;
		}
	}

	async function addTask() {
		if (!newTitle.trim()) return;
		error = '';
		try {
			await tasks.create({ title: newTitle.trim() });
			newTitle = '';
			await loadTasks();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to add task';
		}
	}

	async function completeTask(id: string) {
		error = '';
		try {
			await tasks.done(id);
			taskList = taskList.filter((t) => t.id !== id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to complete task';
		}
	}

	async function deleteTask(id: string) {
		error = '';
		try {
			await tasks.delete(id);
			taskList = taskList.filter((t) => t.id !== id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete task';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') addTask();
	}

	function priorityClass(p?: string): string {
		if (!p || p === 'none') return '';
		return `priority-${p}`;
	}

	onMount(() => {
		loadTasks();
	});
</script>

<div class="view">
	<h1>📥 Inbox</h1>

	<div class="add-bar">
		<input
			type="text"
			placeholder="Add a task..."
			bind:value={newTitle}
			onkeydown={handleKeydown}
		/>
		<button onclick={addTask} disabled={!newTitle.trim()}>Add</button>
	</div>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if taskList.length === 0}
		<p class="empty">No pending tasks. Add one above!</p>
	{:else}
		<ul class="task-list">
			{#each taskList as task (task.id)}
				<li class="task-item">
					<button class="check-btn" onclick={() => completeTask(task.id)} title="Complete">
						<span class="check-circle"></span>
					</button>
					<div class="task-body">
						<span class="task-title">{task.title}</span>
						<div class="task-meta">
							{#if task.priority && task.priority !== 'none'}
								<span class="badge {priorityClass(task.priority)}">{task.priority}</span>
							{/if}
							{#if task.energy}
								<span class="badge energy">{task.energy}</span>
							{/if}
							{#if task.due_date}
								<span class="badge due">due {task.due_date}</span>
							{/if}
							{#each task.tags as tag}
								<span class="badge tag">#{tag}</span>
							{/each}
						</div>
					</div>
					<button class="delete-btn" onclick={() => deleteTask(task.id)} title="Delete">&times;</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.view { max-width: 700px; }
	h1 { margin-bottom: 1.5rem; font-size: 1.5rem; }

	.add-bar {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
	}

	.add-bar input {
		flex: 1;
		padding: 0.6rem 0.8rem;
		background: #1a1a1a;
		border: 1px solid #333;
		border-radius: 6px;
		color: #e0e0e0;
		font-size: 0.9rem;
	}

	.add-bar input:focus {
		outline: none;
		border-color: #555;
	}

	.add-bar button {
		padding: 0.6rem 1.2rem;
		background: #2563eb;
		color: white;
		border: none;
		border-radius: 6px;
		cursor: pointer;
		font-size: 0.9rem;
	}

	.add-bar button:disabled {
		opacity: 0.5;
		cursor: not-allowed;
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

	.empty {
		color: #666;
		text-align: center;
		padding: 3rem;
	}

	.task-list {
		list-style: none;
	}

	.task-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.75rem 0;
		border-bottom: 1px solid #1f1f1f;
	}

	.check-btn {
		background: none;
		border: none;
		cursor: pointer;
		padding: 0.2rem;
		flex-shrink: 0;
		margin-top: 0.1rem;
	}

	.check-circle {
		display: block;
		width: 18px;
		height: 18px;
		border: 2px solid #555;
		border-radius: 50%;
		transition: border-color 0.15s;
	}

	.check-btn:hover .check-circle {
		border-color: #4ade80;
	}

	.task-body {
		flex: 1;
		min-width: 0;
	}

	.task-title {
		display: block;
		font-size: 0.95rem;
	}

	.task-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
		margin-top: 0.3rem;
	}

	.delete-btn {
		background: none;
		border: none;
		color: #555;
		font-size: 1.2rem;
		cursor: pointer;
		padding: 0 0.3rem;
		line-height: 1;
		flex-shrink: 0;
	}

	.delete-btn:hover {
		color: #ff6b6b;
	}
</style>
