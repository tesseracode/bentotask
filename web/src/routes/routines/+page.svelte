<script lang="ts">
	import { onMount } from 'svelte';
	import { routines, type TaskJSON } from '$lib/api';

	let routineList: TaskJSON[] = $state([]);
	let loading = $state(true);
	let error = $state('');

	// Create form
	let showCreate = $state(false);
	let newTitle = $state('');
	let newPriority = $state('');
	let newEnergy = $state('');
	let newSteps: { title: string; duration: number; optional: boolean }[] = $state([]);

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

	function addNewStep() {
		newSteps = [...newSteps, { title: '', duration: 5, optional: false }];
	}

	function removeNewStep(idx: number) {
		newSteps = newSteps.filter((_, i) => i !== idx);
	}

	async function createRoutine() {
		if (!newTitle.trim()) return;
		error = '';
		try {
			const steps = newSteps.filter((s) => s.title.trim()).map((s) => ({
				title: s.title.trim(),
				duration: s.duration || undefined,
				optional: s.optional || undefined
			}));
			await routines.create({
				title: newTitle.trim(),
				steps,
				priority: newPriority || undefined,
				energy: newEnergy || undefined
			});
			newTitle = '';
			newPriority = '';
			newEnergy = '';
			newSteps = [];
			showCreate = false;
			loading = true;
			await loadRoutines();
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to create routine';
		}
	}

	function handleKeydown(e: KeyboardEvent) {
		if (e.key === 'Enter') {
			if (!showCreate) return;
			createRoutine();
		}
	}

	onMount(() => { loadRoutines(); });
</script>

<div class="view">
	<h1>Routines</h1>

	{#if !showCreate}
		<div class="add-bar">
			<input type="text" placeholder="New routine name..." bind:value={newTitle} onkeydown={handleKeydown} />
			<button onclick={() => { if (newTitle.trim()) { showCreate = true; addNewStep(); } }} disabled={!newTitle.trim()}>+</button>
		</div>
	{:else}
		<div class="create-form">
			<label>Title <input type="text" bind:value={newTitle} /></label>
			<div class="form-row">
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
			<div class="steps-editor">
				<span class="steps-label">Steps</span>
				{#each newSteps as step, i}
					<div class="step-edit-row">
						<span class="step-edit-num">{i + 1}</span>
						<input type="text" placeholder="Step title..." bind:value={step.title} class="step-edit-title" />
						<input type="number" min="0" max="120" bind:value={step.duration} class="step-edit-dur" />
						<label class="step-edit-opt"><input type="checkbox" bind:checked={step.optional} /> opt</label>
						<button class="step-edit-del" onclick={() => removeNewStep(i)}>&times;</button>
					</div>
				{/each}
				<button class="add-step-btn" onclick={addNewStep}>+ Add Step</button>
			</div>
			<div class="form-actions">
				<button class="save-btn" onclick={createRoutine} disabled={!newTitle.trim()}>Create Routine</button>
				<button class="cancel-btn" onclick={() => { showCreate = false; newSteps = []; }}>Cancel</button>
			</div>
		</div>
	{/if}

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else if routineList.length === 0}
		<p class="empty">No routines yet. Create one above!</p>
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

	.add-bar { display: flex; gap: 0.5rem; margin-bottom: 1.5rem; }
	.add-bar input { flex: 1; padding: 0.6rem 0.8rem; background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-input); color: var(--text-primary); font-size: 0.9rem; }
	.add-bar input:focus { outline: none; border-color: var(--accent-primary); }
	.add-bar button { padding: 0.6rem 1rem; background: var(--accent-primary); color: var(--text-on-accent); border: none; border-radius: var(--radius-button); cursor: pointer; font-size: 1rem; }
	.add-bar button:disabled { opacity: 0.5; cursor: not-allowed; }

	.create-form { background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-card); padding: 1rem; margin-bottom: 1.5rem; display: flex; flex-direction: column; gap: 0.6rem; }
	.create-form label { display: flex; flex-direction: column; gap: 0.2rem; font-size: 0.75rem; color: var(--text-secondary); }
	.create-form input, .create-form select { padding: 0.4rem 0.6rem; background: var(--bg-base); border: 1px solid var(--border-default); border-radius: var(--radius-badge); color: var(--text-primary); font-size: 0.85rem; }
	.create-form input:focus, .create-form select:focus { outline: none; border-color: var(--accent-primary); }
	.form-row { display: flex; gap: 0.6rem; }
	.form-row label { flex: 1; }

	.steps-editor { display: flex; flex-direction: column; gap: 0.35rem; }
	.steps-label { font-size: 0.75rem; color: var(--text-secondary); margin-bottom: 0.2rem; }
	.step-edit-row { display: flex; gap: 0.35rem; align-items: center; }
	.step-edit-num { font-size: 0.75rem; color: var(--text-tertiary); width: 1.2rem; text-align: center; }
	.step-edit-title { flex: 1 !important; }
	.step-edit-dur { width: 55px !important; }
	.step-edit-opt { flex-direction: row !important; align-items: center; gap: 0.25rem !important; font-size: 0.7rem !important; white-space: nowrap; }
	.step-edit-opt input { width: auto; }
	.step-edit-del { background: none; border: none; color: var(--text-tertiary); cursor: pointer; font-size: 1rem; padding: 0 0.3rem; }
	.step-edit-del:hover { color: var(--warning-text); }
	.add-step-btn { background: none; border: 1px dashed var(--border-default); border-radius: var(--radius-badge); color: var(--text-secondary); cursor: pointer; padding: 0.35rem; font-size: 0.8rem; margin-top: 0.2rem; }
	.add-step-btn:hover { border-color: var(--accent-primary); color: var(--accent-primary); }

	.form-actions { display: flex; gap: 0.5rem; margin-top: 0.25rem; }
	.save-btn { padding: 0.4rem 0.8rem; background: var(--accent-primary); color: var(--text-on-accent); border: none; border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem; }
	.save-btn:disabled { opacity: 0.5; cursor: not-allowed; }
	.cancel-btn { padding: 0.4rem 0.8rem; background: var(--bg-elevated); color: var(--text-secondary); border: 1px solid var(--border-default); border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem; }

	.error { padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning); border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem; }
	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.routine-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; }
	.routine-card { background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-card); box-shadow: var(--shadow-card); transition: border-color 0.15s; }
	.routine-card:hover { border-color: var(--accent-primary); }
	.routine-link { display: flex; justify-content: space-between; align-items: center; padding: 1rem 1.25rem; text-decoration: none; color: var(--text-primary); }
	.routine-title { font-size: 0.95rem; font-weight: 500; }
	.routine-meta { display: flex; gap: 0.5rem; align-items: center; }
	.meta-item { font-size: 0.75rem; color: var(--text-secondary); }
</style>
