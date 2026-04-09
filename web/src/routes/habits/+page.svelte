<script lang="ts">
	import { onMount } from 'svelte';
	import { habits, type TaskJSON } from '$lib/api';

	let habitList: TaskJSON[] = $state([]);
	let loading = $state(true);
	let error = $state('');
	let newTitle = $state('');

	async function loadHabits() {
		try {
			const res = await habits.list();
			habitList = res.items;
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
			await habits.create({ title: newTitle.trim(), freq_type: 'daily', freq_target: 1 });
			newTitle = '';
			await loadHabits();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to add habit';
		}
	}

	async function logHabit(id: string) {
		error = '';
		try {
			await habits.log(id);
			await loadHabits();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to log habit';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') addHabit();
	}

	onMount(() => {
		loadHabits();
	});
</script>

<div class="view">
	<h1>🔥 Habits</h1>

	<div class="add-bar">
		<input
			type="text"
			placeholder="Add a daily habit..."
			bind:value={newTitle}
			onkeydown={handleKeydown}
		/>
		<button onclick={addHabit} disabled={!newTitle.trim()}>Add</button>
	</div>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if habitList.length === 0}
		<p class="empty">No habits yet. Start building a streak!</p>
	{:else}
		<ul class="habit-list">
			{#each habitList as habit (habit.id)}
				<li class="habit-item">
					<div class="habit-body">
						<span class="habit-title">{habit.title}</span>
						<div class="habit-meta">
							{#if habit.priority && habit.priority !== 'none'}
								<span class="badge priority-{habit.priority}">{habit.priority}</span>
							{/if}
							{#if habit.energy}
								<span class="badge energy">{habit.energy}</span>
							{/if}
						</div>
					</div>
					<button class="log-btn" onclick={() => logHabit(habit.id)} title="Log completion">
						Log
					</button>
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

	.empty { color: #666; text-align: center; padding: 3rem; }

	.habit-list { list-style: none; }

	.habit-item {
		display: flex;
		align-items: center;
		gap: 0.75rem;
		padding: 0.75rem 0;
		border-bottom: 1px solid #1f1f1f;
	}

	.habit-body { flex: 1; }

	.habit-title {
		display: block;
		font-size: 0.95rem;
	}

	.habit-meta {
		display: flex;
		gap: 0.35rem;
		margin-top: 0.25rem;
	}

	.log-btn {
		padding: 0.4rem 0.8rem;
		background: #1a3a1a;
		border: 1px solid #2a5a2a;
		border-radius: 6px;
		color: #4ade80;
		cursor: pointer;
		font-size: 0.8rem;
		flex-shrink: 0;
	}

	.log-btn:hover {
		background: #2a5a2a;
	}
</style>
