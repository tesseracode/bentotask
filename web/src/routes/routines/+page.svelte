<script lang="ts">
	import { onMount } from 'svelte';
	import { routines, type TaskJSON } from '$lib/api';

	let routineList: TaskJSON[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	async function loadRoutines() {
		error = '';
		try {
			const res = await routines.list();
			routineList = res.items;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load routines';
		} finally {
			loading = false;
		}
	}

	onMount(() => { loadRoutines(); });
</script>

<div class="view">
	<h1>Routines</h1>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if routineList.length === 0}
		<p class="empty">No routines yet. Create one with <code>bt routine create</code>.</p>
	{:else}
		<ul class="routine-list">
			{#each routineList as routine (routine.id)}
				<li class="routine-card">
					<a href="/routines/{routine.id}" class="routine-link">
						<span class="routine-title">{routine.title}</span>
						<div class="routine-meta">
							{#if routine.estimated_duration}
								<span class="meta-item">~{routine.estimated_duration}m</span>
							{/if}
							{#if routine.priority && routine.priority !== 'none'}
								<span class="badge priority-{routine.priority}">{routine.priority}</span>
							{/if}
						</div>
					</a>
				</li>
			{/each}
		</ul>
	{/if}
</div>

<style>
	.view { max-width: 700px; }
	h1 { margin-bottom: 1.5rem; font-size: 1.5rem; color: var(--text-primary); }

	.error {
		padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning);
		border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem;
	}

	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }
	.empty code { background: var(--bg-elevated); padding: 0.15rem 0.4rem; border-radius: 3px; font-size: 0.85em; }

	.routine-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; }

	.routine-card {
		background: var(--bg-surface); border: 1px solid var(--border-default);
		border-radius: var(--radius-card); box-shadow: var(--shadow-card);
		transition: border-color 0.15s;
	}

	.routine-card:hover { border-color: var(--accent-primary); }

	.routine-link {
		display: flex; justify-content: space-between; align-items: center;
		padding: 1rem 1.25rem; text-decoration: none; color: var(--text-primary);
	}

	.routine-title { font-size: 0.95rem; font-weight: 500; }

	.routine-meta { display: flex; gap: 0.5rem; align-items: center; }

	.meta-item { font-size: 0.75rem; color: var(--text-secondary); }
</style>
