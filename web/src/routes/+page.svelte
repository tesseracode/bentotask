<script lang="ts">
	import { onMount } from 'svelte';
	import { tasks, type TaskJSON, type UpdateTaskRequest } from '$lib/api';
	import { renderWikilinks } from '$lib/wikilink';

	let taskList: TaskJSON[] = $state([]);
	let newTitle = $state('');
	let loading = $state(true);
	let error = $state('');

	// Filters
	let filterPriority = $state('');
	let filterEnergy = $state('');
	let filterTag = $state('');
	let sortBy: 'created' | 'priority' = $state('created');

	// Expanded / Edit state
	let expandedId: string | null = $state(null);
	let expandedTask: TaskJSON | null = $state(null);
	let editing = $state(false);
	let editTitle = $state('');
	let editPriority = $state('');
	let editEnergy = $state('');
	let editDue = $state('');
	let editTags = $state('');

	const priorityOrder: Record<string, number> = { urgent: 0, high: 1, medium: 2, low: 3, none: 4, '': 5 };

	let sortedTasks = $derived.by(() => {
		if (sortBy === 'priority') {
			return [...taskList].sort((a, b) => {
				const pa = priorityOrder[a.priority ?? ''] ?? 5;
				const pb = priorityOrder[b.priority ?? ''] ?? 5;
				return pa - pb;
			});
		}
		return taskList;
	});

	async function loadTasks() {
		error = '';
		try {
			const params: Record<string, string> = { status: 'pending' };
			if (filterPriority) params.priority = filterPriority;
			if (filterEnergy) params.energy = filterEnergy;
			if (filterTag) params.tag = filterTag;
			const res = await tasks.list(params);
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
			if (expandedId === id) { expandedId = null; expandedTask = null; }
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to complete task';
		}
	}

	async function deleteTask(id: string) {
		error = '';
		try {
			await tasks.delete(id);
			taskList = taskList.filter((t) => t.id !== id);
			if (expandedId === id) { expandedId = null; expandedTask = null; }
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete task';
		}
	}

	async function toggleExpand(id: string) {
		if (expandedId === id) {
			expandedId = null;
			expandedTask = null;
			editing = false;
			return;
		}
		error = '';
		expandedId = id;
		editing = false;
		try {
			expandedTask = await tasks.get(id);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load task details';
			expandedId = null;
			expandedTask = null;
		}
	}

	function startEdit() {
		if (!expandedTask) return;
		editing = true;
		editTitle = expandedTask.title;
		editPriority = expandedTask.priority ?? '';
		editEnergy = expandedTask.energy ?? '';
		editDue = expandedTask.due_date ?? '';
		editTags = expandedTask.tags.join(', ');
	}

	function cancelEdit() {
		editing = false;
	}

	async function saveEdit() {
		if (!expandedTask) return;
		error = '';
		const changes: UpdateTaskRequest = {};
		if (editTitle !== expandedTask.title) changes.title = editTitle;
		if (editPriority !== (expandedTask.priority ?? '')) changes.priority = editPriority;
		if (editEnergy !== (expandedTask.energy ?? '')) changes.energy = editEnergy;
		if (editDue !== (expandedTask.due_date ?? '')) changes.due_date = editDue;
		const newTags = editTags.split(',').map((t) => t.trim()).filter(Boolean);
		if (JSON.stringify(newTags) !== JSON.stringify(expandedTask.tags)) changes.tags = newTags;

		try {
			expandedTask = await tasks.update(expandedTask.id, changes);
			editing = false;
			await loadTasks();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to update task';
		}
	}

	function applyFilters() {
		loading = true;
		loadTasks();
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
	<h1>Inbox</h1>

	<div class="add-bar">
		<input
			type="text"
			placeholder="Add a task..."
			bind:value={newTitle}
			onkeydown={handleKeydown}
		/>
		<button onclick={addTask} disabled={!newTitle.trim()}>Add</button>
	</div>

	<div class="filter-bar">
		<select bind:value={filterPriority} onchange={applyFilters}>
			<option value="">All priorities</option>
			<option value="urgent">Urgent</option>
			<option value="high">High</option>
			<option value="medium">Medium</option>
			<option value="low">Low</option>
		</select>
		<select bind:value={filterEnergy} onchange={applyFilters}>
			<option value="">All energy</option>
			<option value="low">Low</option>
			<option value="medium">Medium</option>
			<option value="high">High</option>
		</select>
		<input
			type="text"
			placeholder="Filter by tag..."
			class="tag-filter"
			bind:value={filterTag}
			onchange={applyFilters}
		/>
		<button class="sort-btn" onclick={() => { sortBy = sortBy === 'created' ? 'priority' : 'created'; }} title="Toggle sort">
			Sort: {sortBy === 'created' ? 'Date' : 'Priority'}
		</button>
	</div>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if sortedTasks.length === 0}
		<p class="empty">No pending tasks. Add one above!</p>
	{:else}
		<ul class="task-list">
			{#each sortedTasks as task (task.id)}
				<li class="task-item" class:expanded={expandedId === task.id}>
					<button class="check-btn" onclick={() => completeTask(task.id)} title="Complete">
						<span class="check-circle"></span>
					</button>
					<div class="task-body">
						<button class="task-title-btn" onclick={() => toggleExpand(task.id)}>
							{task.title}
						</button>
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

						{#if expandedId === task.id && expandedTask}
							<div class="detail-panel">
								{#if !editing}
									<div class="detail-grid">
										<span class="detail-label">ID</span>
										<span class="detail-value mono">{expandedTask.id}</span>
										{#if expandedTask.due_date}
											<span class="detail-label">Due</span>
											<span class="detail-value">{expandedTask.due_date}</span>
										{/if}
										{#if expandedTask.contexts.length > 0}
											<span class="detail-label">Contexts</span>
											<span class="detail-value">{expandedTask.contexts.join(', ')}</span>
										{/if}
										<span class="detail-label">File</span>
										<span class="detail-value mono">{expandedTask.file_path}</span>
										<span class="detail-label">Created</span>
										<span class="detail-value">{new Date(expandedTask.created_at).toLocaleString()}</span>
										<span class="detail-label">Updated</span>
										<span class="detail-value">{new Date(expandedTask.updated_at).toLocaleString()}</span>
									</div>
									{#if expandedTask.body}
										<div class="detail-body">{@html renderWikilinks(expandedTask.body)}</div>
									{/if}
									<button class="edit-btn" onclick={startEdit}>Edit</button>
								{:else}
									<div class="edit-form">
										<label>
											Title
											<input type="text" bind:value={editTitle} />
										</label>
										<div class="edit-row">
											<label>
												Priority
												<select bind:value={editPriority}>
													<option value="">None</option>
													<option value="low">Low</option>
													<option value="medium">Medium</option>
													<option value="high">High</option>
													<option value="urgent">Urgent</option>
												</select>
											</label>
											<label>
												Energy
												<select bind:value={editEnergy}>
													<option value="">None</option>
													<option value="low">Low</option>
													<option value="medium">Medium</option>
													<option value="high">High</option>
												</select>
											</label>
											<label>
												Due date
												<input type="date" bind:value={editDue} />
											</label>
										</div>
										<label>
											Tags <span class="hint">(comma-separated)</span>
											<input type="text" bind:value={editTags} />
										</label>
										<div class="edit-actions">
											<button class="save-btn" onclick={saveEdit}>Save</button>
											<button class="cancel-btn" onclick={cancelEdit}>Cancel</button>
										</div>
									</div>
								{/if}
							</div>
						{/if}
					</div>
					<button class="delete-btn" onclick={() => deleteTask(task.id)} title="Delete">&times;</button>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.view { max-width: 700px; }
	h1 { margin-bottom: 1.5rem; font-size: 1.5rem; color: var(--text-primary); }

	.add-bar {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 0.75rem;
	}

	.add-bar input {
		flex: 1;
		padding: 0.6rem 0.8rem;
		background: var(--bg-surface);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-input);
		color: var(--text-primary);
		font-size: 0.9rem;
	}

	.add-bar input:focus { outline: none; border-color: var(--accent-primary); }

	.add-bar button {
		padding: 0.6rem 1.2rem;
		background: var(--accent-primary);
		color: var(--text-on-accent);
		border: none;
		border-radius: var(--radius-button);
		cursor: pointer;
		font-size: 0.9rem;
	}

	.add-bar button:disabled { opacity: 0.5; cursor: not-allowed; }

	.filter-bar {
		display: flex;
		gap: 0.5rem;
		margin-bottom: 1.5rem;
		flex-wrap: wrap;
	}

	.filter-bar select,
	.filter-bar .tag-filter {
		padding: 0.4rem 0.6rem;
		background: var(--bg-surface);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-badge);
		color: var(--text-secondary);
		font-size: 0.8rem;
	}

	.filter-bar .tag-filter { width: 120px; }
	.filter-bar select:focus, .filter-bar .tag-filter:focus { outline: none; border-color: var(--accent-primary); }

	.sort-btn {
		padding: 0.4rem 0.6rem;
		background: var(--bg-elevated);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-badge);
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.8rem;
		margin-left: auto;
	}

	.sort-btn:hover { color: var(--text-primary); border-color: var(--accent-primary); }

	.error {
		padding: 0.6rem;
		background: var(--warning-subtle);
		border: 1px solid var(--warning);
		border-radius: var(--radius-badge);
		color: var(--warning-text);
		margin-bottom: 1rem;
		font-size: 0.85rem;
	}

	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.task-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; }

	.task-item {
		display: flex;
		align-items: flex-start;
		gap: 0.75rem;
		padding: 0.8rem 1rem;
		background: var(--bg-surface);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-card);
		box-shadow: var(--shadow-card);
	}

	.task-item.expanded { border-color: var(--accent-primary); }

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
		border: 2px solid var(--accent-primary);
		border-radius: 50%;
		transition: border-color 0.15s, background 0.15s;
	}

	.check-btn:hover .check-circle { border-color: var(--success); background: var(--success-subtle); }

	.task-body { flex: 1; min-width: 0; }

	.task-title-btn {
		display: block;
		background: none;
		border: none;
		color: var(--text-primary);
		font-size: 0.93rem;
		font-weight: 500;
		cursor: pointer;
		text-align: left;
		padding: 0;
		width: 100%;
	}

	.task-title-btn:hover { color: var(--accent-hover); }

	.task-meta {
		display: flex;
		flex-wrap: wrap;
		gap: 0.35rem;
		margin-top: 0.3rem;
	}

	.delete-btn {
		background: none;
		border: none;
		color: var(--text-tertiary);
		font-size: 1.2rem;
		cursor: pointer;
		padding: 0 0.3rem;
		line-height: 1;
		flex-shrink: 0;
	}

	.delete-btn:hover { color: var(--warning-text); }

	.detail-panel {
		margin-top: 0.75rem;
		padding-top: 0.75rem;
		border-top: 1px solid var(--border-default);
	}

	.detail-grid {
		display: grid;
		grid-template-columns: 80px 1fr;
		gap: 0.3rem 0.75rem;
		font-size: 0.8rem;
		margin-bottom: 0.75rem;
	}

	.detail-label { color: var(--text-tertiary); }
	.detail-value { color: var(--text-secondary); }
	.mono { font-family: monospace; font-size: 0.75rem; }

	.detail-body {
		background: var(--bg-base);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-badge);
		padding: 0.6rem;
		font-size: 0.8rem;
		color: var(--text-secondary);
		white-space: pre-wrap;
		margin-bottom: 0.75rem;
		font-family: inherit;
	}

	.edit-btn {
		padding: 0.35rem 0.7rem;
		background: var(--bg-elevated);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-badge);
		color: var(--text-secondary);
		cursor: pointer;
		font-size: 0.8rem;
	}

	.edit-btn:hover { border-color: var(--accent-primary); color: var(--text-primary); }

	.edit-form {
		display: flex;
		flex-direction: column;
		gap: 0.6rem;
	}

	.edit-form label {
		display: flex;
		flex-direction: column;
		gap: 0.2rem;
		font-size: 0.75rem;
		color: var(--text-secondary);
	}

	.hint { color: var(--text-tertiary); }

	.edit-form input, .edit-form select {
		padding: 0.4rem 0.6rem;
		background: var(--bg-base);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-badge);
		color: var(--text-primary);
		font-size: 0.85rem;
	}

	.edit-form input:focus, .edit-form select:focus { outline: none; border-color: var(--accent-primary); }

	.edit-row { display: flex; gap: 0.5rem; }
	.edit-row label { flex: 1; }

	.edit-actions { display: flex; gap: 0.5rem; margin-top: 0.25rem; }

	.save-btn {
		padding: 0.4rem 0.8rem;
		background: var(--accent-primary);
		color: var(--text-on-accent);
		border: none;
		border-radius: var(--radius-badge);
		cursor: pointer;
		font-size: 0.8rem;
	}

	.cancel-btn {
		padding: 0.4rem 0.8rem;
		background: var(--bg-elevated);
		color: var(--text-secondary);
		border: 1px solid var(--border-default);
		border-radius: var(--radius-badge);
		cursor: pointer;
		font-size: 0.8rem;
	}
</style>
