<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { goto } from '$app/navigation';
	import { routines, tasks, type TaskJSON, type StepJSON, type UpdateTaskRequest } from '$lib/api';

	let routine: TaskJSON | null = $state(null);
	let loading = $state(true);
	let error = $state('');

	// Play state
	let playing = $state(false);
	let currentStep = $state(0);
	let elapsed = $state(0);
	let stepCompleted: boolean[] = $state([]);
	let totalElapsed = $state(0);
	let finished = $state(false);
	let timerInterval: ReturnType<typeof setInterval> | null = $state(null);

	// Edit state
	let editing = $state(false);
	let editTitle = $state('');
	let editPriority = $state('');
	let editEnergy = $state('');
	let editSteps: { title: string; duration: number; optional: boolean }[] = $state([]);

	function routineId(): string {
		return page.params.id ?? '';
	}

	async function loadRoutine() {
		error = '';
		try {
			routine = await routines.get(routineId());
			if (routine.steps) {
				stepCompleted = routine.steps.map(() => false);
			}
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load routine';
		} finally {
			loading = false;
		}
	}

	// --- Play ---
	function startPlay() {
		if (!routine?.steps?.length) return;
		playing = true; editing = false;
		currentStep = 0; elapsed = 0; totalElapsed = 0; finished = false;
		stepCompleted = routine.steps.map(() => false);
		startTimer();
	}

	function startTimer() {
		stopTimer();
		timerInterval = setInterval(() => { elapsed += 1; totalElapsed += 1; }, 1000);
	}

	function stopTimer() {
		if (timerInterval !== null) { clearInterval(timerInterval); timerInterval = null; }
	}

	function nextStep() {
		if (!routine?.steps) return;
		stepCompleted[currentStep] = true;
		stepCompleted = [...stepCompleted];
		if (currentStep < routine.steps.length - 1) { currentStep += 1; elapsed = 0; }
		else { stopTimer(); finished = true; }
	}

	function skipStep() {
		if (!routine?.steps) return;
		if (currentStep < routine.steps.length - 1) { currentStep += 1; elapsed = 0; }
		else { stopTimer(); finished = true; }
	}

	function stopPlay() {
		stopTimer(); playing = false; finished = false;
		currentStep = 0; elapsed = 0; totalElapsed = 0;
	}

	function formatElapsed(seconds: number): string {
		const m = Math.floor(seconds / 60);
		const s = seconds % 60;
		return `${m}:${String(s).padStart(2, '0')}`;
	}

	function stepProgress(step: StepJSON, elapsedSec: number): number {
		if (!step.duration || step.duration === 0) return 0;
		return Math.min(100, Math.round((elapsedSec / (step.duration * 60)) * 100));
	}

	// --- Edit ---
	function startEdit() {
		if (!routine) return;
		editing = true;
		editTitle = routine.title;
		editPriority = routine.priority ?? '';
		editEnergy = routine.energy ?? '';
		editSteps = (routine.steps ?? []).map((s) => ({ title: s.title, duration: s.duration ?? 0, optional: s.optional ?? false }));
	}

	function cancelEdit() { editing = false; }

	function addEditStep() {
		editSteps = [...editSteps, { title: '', duration: 5, optional: false }];
	}

	function removeEditStep(idx: number) {
		editSteps = editSteps.filter((_, i) => i !== idx);
	}

	async function saveEdit() {
		if (!routine) return;
		error = '';
		const changes: UpdateTaskRequest = {};
		if (editTitle !== routine.title) changes.title = editTitle;
		if (editPriority !== (routine.priority ?? '')) changes.priority = editPriority;
		if (editEnergy !== (routine.energy ?? '')) changes.energy = editEnergy;
		changes.steps = editSteps.filter((s) => s.title.trim()).map((s) => ({
			title: s.title.trim(),
			duration: s.duration || undefined,
			optional: s.optional || undefined
		}));
		try {
			routine = await tasks.update(routine.id, changes);
			if (routine.steps) stepCompleted = routine.steps.map(() => false);
			editing = false;
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to update routine';
		}
	}

	async function deleteRoutine() {
		if (!routine) return;
		if (!confirm('Delete this routine? This cannot be undone.')) return;
		error = '';
		try {
			await tasks.delete(routine.id);
			goto('/routines');
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to delete routine';
		}
	}

	// --- Keyboard ---
	function handlePlayKeydown(e: KeyboardEvent) {
		if (!playing || finished) return;
		if (e.key === 'Enter' || e.key === 'ArrowRight') { e.preventDefault(); nextStep(); }
		else if (e.key === 's' || e.key === 'S') {
			if (routine?.steps?.[currentStep]?.optional) { e.preventDefault(); skipStep(); }
		}
		else if (e.key === 'Escape') { e.preventDefault(); stopPlay(); }
	}

	onMount(() => { loadRoutine(); return () => stopTimer(); });
</script>

<svelte:window onkeydown={handlePlayKeydown} />

<div class="view">
	{#if loading}
		<p class="empty">Loading...</p>
	{:else if error}
		<div class="error">{error}</div>
	{:else if routine}
		<div class="header">
			<a href="/routines" class="back-link">&larr; Routines</a>
			{#if !editing}
				<h1>{routine.title}</h1>
				{#if routine.steps}
					<span class="step-count">{routine.steps.length} steps</span>
				{/if}
			{/if}
		</div>

		{#if editing}
			<!-- Edit mode -->
			<div class="edit-form">
				<label>Title <input type="text" bind:value={editTitle} /></label>
				<div class="edit-row">
					<label>Priority
						<select bind:value={editPriority}>
							<option value="">None</option><option value="low">Low</option><option value="medium">Medium</option>
							<option value="high">High</option><option value="urgent">Urgent</option>
						</select>
					</label>
					<label>Energy
						<select bind:value={editEnergy}>
							<option value="">None</option><option value="low">Low</option><option value="medium">Medium</option><option value="high">High</option>
						</select>
					</label>
				</div>
				<div class="steps-editor">
					<span class="steps-label">Steps</span>
					{#each editSteps as step, i}
						<div class="step-edit-row">
							<span class="step-edit-num">{i + 1}</span>
							<input type="text" placeholder="Step title..." bind:value={step.title} class="step-edit-title" />
							<input type="number" min="0" max="120" bind:value={step.duration} class="step-edit-dur" />
							<label class="step-edit-opt"><input type="checkbox" bind:checked={step.optional} /> opt</label>
							<button class="step-edit-del" onclick={() => removeEditStep(i)}>&times;</button>
						</div>
					{/each}
					<button class="add-step-btn" onclick={addEditStep}>+ Add Step</button>
				</div>
				<div class="edit-actions">
					<button class="save-btn" onclick={saveEdit}>Save</button>
					<button class="cancel-btn" onclick={cancelEdit}>Cancel</button>
					<button class="delete-routine-btn" onclick={deleteRoutine}>Delete Routine</button>
				</div>
			</div>

		{:else if !playing && !finished}
			<!-- Static step list -->
			{#if routine.steps && routine.steps.length > 0}
				<ul class="step-list">
					{#each routine.steps as step, i}
						<li class="step-item">
							<span class="step-num">{i + 1}</span>
							<div class="step-body">
								<span class="step-title">{step.title}</span>
								<div class="step-meta">
									{#if step.duration}
										<span class="step-duration">~{step.duration}m</span>
									{/if}
									{#if step.optional}
										<span class="step-optional">optional</span>
									{/if}
								</div>
							</div>
						</li>
					{/each}
				</ul>
				<div class="action-row">
					<button class="play-btn" onclick={startPlay} title="Start stepping through the routine">▶ Play Routine</button>
					<button class="edit-btn" onclick={startEdit} title="Edit routine steps, title, and settings">✎ Edit</button>
				</div>
			{:else}
				<p class="empty">This routine has no steps.</p>
				<div class="action-row">
					<button class="edit-btn" onclick={startEdit} title="Add steps to this routine">✎ Edit</button>
				</div>
			{/if}

		{:else if playing && !finished && routine.steps}
			<!-- Play mode -->
			<p class="play-hint">Press <strong>Next</strong> when done with each step. Optional steps can be skipped.</p>
			<ul class="step-list">
				{#each routine.steps as step, i}
					<li class="step-item" class:active={i === currentStep} class:done={stepCompleted[i]}>
						<span class="step-num">{#if stepCompleted[i]}✓{:else}{i + 1}{/if}</span>
						<div class="step-body">
							<span class="step-title">{step.title}</span>
							{#if i === currentStep}
								<div class="step-timer">
									<span class="timer-value">{formatElapsed(elapsed)}</span>
									{#if step.duration}
										<span class="timer-target">/ {step.duration}:00</span>
										<div class="progress-bar">
											<div class="progress-fill" style="width: {stepProgress(step, elapsed)}%"></div>
										</div>
									{/if}
								</div>
								<div class="step-actions">
									<button class="next-btn" onclick={nextStep} title={i === routine.steps.length - 1 ? 'Mark the last step as done and finish' : 'Mark this step as done and move to the next'}>
										{i === routine.steps.length - 1 ? '✓ Finish' : 'Next →'}
									</button>
									{#if step.optional}
										<button class="skip-btn" onclick={skipStep} title="Skip this optional step">Skip</button>
									{/if}
								</div>
								<span class="key-hint">↵ Next · S Skip · Esc Stop</span>
							{:else}
								<div class="step-meta">
									{#if step.duration}
										<span class="step-duration">~{step.duration}m</span>
									{/if}
									{#if step.optional}
										<span class="step-optional">optional</span>
									{/if}
								</div>
							{/if}
						</div>
					</li>
				{/each}
			</ul>
			<div class="play-footer">
				<span class="total-time">Total: {formatElapsed(totalElapsed)}</span>
				<button class="stop-btn" onclick={stopPlay} title="Stop the routine and discard progress">Stop</button>
			</div>

		{:else if finished}
			<!-- Summary -->
			<div class="summary">
				<h2>Routine Complete!</h2>
				<p class="summary-time">Total time: {formatElapsed(totalElapsed)}</p>
				<ul class="summary-steps">
					{#if routine.steps}
						{#each routine.steps as step, i}
							<li class="summary-step">
								<span class="summary-check">{stepCompleted[i] ? '✓' : '—'}</span>
								<span>{step.title}</span>
							</li>
						{/each}
					{/if}
				</ul>
				<button class="play-btn" onclick={startPlay}>▶ Play Again</button>
				<a href="/routines" class="back-link summary-back">&larr; Back to Routines</a>
			</div>
		{/if}
	{/if}
</div>

<style>
	.view { max-width: 600px; }
	.header { margin-bottom: 1.5rem; }
	.back-link { font-size: 0.8rem; color: var(--text-secondary); text-decoration: none; display: inline-block; margin-bottom: 0.5rem; }
	.back-link:hover { color: var(--accent-primary); }
	h1 { font-size: 1.5rem; color: var(--text-primary); margin-bottom: 0.25rem; }
	h2 { font-size: 1.3rem; color: var(--text-primary); margin-bottom: 0.75rem; }
	.step-count { font-size: 0.8rem; color: var(--text-secondary); }
	.error { padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning); border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem; }
	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.step-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; margin-bottom: 1.5rem; }
	.step-item { display: flex; gap: 0.75rem; padding: 0.8rem 1rem; background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-card); box-shadow: var(--shadow-card); transition: border-color 0.2s, opacity 0.2s; }
	.step-item.active { border-color: var(--accent-primary); box-shadow: var(--shadow-elevated); }
	.step-item.done { opacity: 0.5; }
	.step-num { min-width: 1.5rem; text-align: center; font-weight: 600; color: var(--accent-primary); font-size: 0.9rem; margin-top: 0.1rem; }
	.step-item.done .step-num { color: var(--success); }
	.step-body { flex: 1; }
	.step-title { font-size: 0.95rem; font-weight: 500; color: var(--text-primary); display: block; }
	.step-meta { display: flex; gap: 0.5rem; margin-top: 0.25rem; }
	.step-duration { font-size: 0.75rem; color: var(--text-secondary); }
	.step-optional { font-size: 0.65rem; color: var(--text-tertiary); background: var(--bg-elevated); padding: 0.1rem 0.35rem; border-radius: 3px; }

	.step-timer { margin-top: 0.5rem; display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }
	.timer-value { font-size: 1.5rem; font-weight: 300; color: var(--accent-primary); font-variant-numeric: tabular-nums; }
	.timer-target { font-size: 0.85rem; color: var(--text-tertiary); }
	.progress-bar { width: 100%; height: 4px; background: var(--score-track); border-radius: 2px; overflow: hidden; margin-top: 0.3rem; }
	.progress-fill { height: 100%; background: var(--accent-primary); border-radius: 2px; transition: width 1s linear; }
	.step-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
	.key-hint { font-size: 0.65rem; color: var(--text-tertiary); margin-top: 0.3rem; display: block; }
	.play-hint { color: var(--text-tertiary); font-size: 0.8rem; margin-bottom: 1rem; }

	.next-btn { padding: 0.4rem 0.8rem; background: var(--accent-primary); color: var(--text-on-accent); border: none; border-radius: var(--radius-button); cursor: pointer; font-size: 0.85rem; font-weight: 500; }
	.skip-btn { padding: 0.4rem 0.8rem; background: var(--bg-elevated); color: var(--text-secondary); border: 1px solid var(--border-default); border-radius: var(--radius-button); cursor: pointer; font-size: 0.85rem; }

	.action-row { display: flex; gap: 0.75rem; align-items: center; }
	.play-btn { padding: 0.6rem 1.5rem; background: var(--accent-primary); color: var(--text-on-accent); border: none; border-radius: var(--radius-button); cursor: pointer; font-size: 0.95rem; font-weight: 500; }
	.edit-btn { padding: 0.6rem 1rem; background: var(--bg-elevated); color: var(--text-secondary); border: 1px solid var(--border-default); border-radius: var(--radius-button); cursor: pointer; font-size: 0.85rem; }
	.edit-btn:hover { border-color: var(--accent-primary); color: var(--text-primary); }

	.play-footer { display: flex; justify-content: space-between; align-items: center; padding-top: 1rem; border-top: 1px solid var(--border-default); }
	.total-time { font-size: 0.85rem; color: var(--text-secondary); }
	.stop-btn { padding: 0.4rem 0.8rem; background: var(--warning-subtle); color: var(--warning-text); border: 1px solid var(--warning); border-radius: var(--radius-button); cursor: pointer; font-size: 0.8rem; }

	.summary { text-align: center; padding: 2rem 0; }
	.summary-time { font-size: 1.2rem; color: var(--accent-primary); margin-bottom: 1.5rem; }
	.summary-steps { list-style: none; text-align: left; max-width: 400px; margin: 0 auto 1.5rem; }
	.summary-step { padding: 0.3rem 0; font-size: 0.9rem; color: var(--text-secondary); }
	.summary-check { margin-right: 0.5rem; color: var(--success); }
	.summary-back { display: inline-block; margin-top: 1rem; }

	/* Edit form */
	.edit-form { display: flex; flex-direction: column; gap: 0.6rem; }
	.edit-form label { display: flex; flex-direction: column; gap: 0.2rem; font-size: 0.75rem; color: var(--text-secondary); }
	.edit-form input, .edit-form select { padding: 0.4rem 0.6rem; background: var(--bg-surface); border: 1px solid var(--border-default); border-radius: var(--radius-badge); color: var(--text-primary); font-size: 0.85rem; }
	.edit-form input:focus, .edit-form select:focus { outline: none; border-color: var(--accent-primary); }
	.edit-row { display: flex; gap: 0.5rem; }
	.edit-row label { flex: 1; }
	.edit-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }
	.save-btn { padding: 0.4rem 0.8rem; background: var(--accent-primary); color: var(--text-on-accent); border: none; border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem; }
	.cancel-btn { padding: 0.4rem 0.8rem; background: var(--bg-elevated); color: var(--text-secondary); border: 1px solid var(--border-default); border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem; }
	.delete-routine-btn { padding: 0.4rem 0.8rem; background: var(--warning-subtle); color: var(--warning-text); border: 1px solid var(--warning); border-radius: var(--radius-badge); cursor: pointer; font-size: 0.8rem; margin-left: auto; }

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
</style>
