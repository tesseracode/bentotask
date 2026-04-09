<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { routines, type TaskJSON, type StepJSON } from '$lib/api';

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

	function startPlay() {
		if (!routine?.steps?.length) return;
		playing = true;
		currentStep = 0;
		elapsed = 0;
		totalElapsed = 0;
		finished = false;
		stepCompleted = routine.steps.map(() => false);
		startTimer();
	}

	function startTimer() {
		stopTimer();
		timerInterval = setInterval(() => {
			elapsed += 1;
			totalElapsed += 1;
		}, 1000);
	}

	function stopTimer() {
		if (timerInterval !== null) {
			clearInterval(timerInterval);
			timerInterval = null;
		}
	}

	function nextStep() {
		if (!routine?.steps) return;
		stepCompleted[currentStep] = true;
		stepCompleted = [...stepCompleted]; // trigger reactivity

		if (currentStep < routine.steps.length - 1) {
			currentStep += 1;
			elapsed = 0;
		} else {
			stopTimer();
			finished = true;
		}
	}

	function skipStep() {
		if (!routine?.steps) return;
		if (currentStep < routine.steps.length - 1) {
			currentStep += 1;
			elapsed = 0;
		} else {
			stopTimer();
			finished = true;
		}
	}

	function stopPlay() {
		stopTimer();
		playing = false;
		finished = false;
		currentStep = 0;
		elapsed = 0;
		totalElapsed = 0;
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

	onMount(() => {
		loadRoutine();
		return () => stopTimer();
	});
</script>

<div class="view">
	{#if loading}
		<p class="empty">Loading...</p>
	{:else if error}
		<div class="error">{error}</div>
	{:else if routine}
		<div class="header">
			<a href="/routines" class="back-link">&larr; Routines</a>
			<h1>{routine.title}</h1>
			{#if routine.steps}
				<span class="step-count">{routine.steps.length} steps</span>
			{/if}
		</div>

		{#if !playing && !finished}
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
				<button class="play-btn" onclick={startPlay}>▶ Play Routine</button>
			{:else}
				<p class="empty">This routine has no steps.</p>
			{/if}

		{:else if playing && !finished && routine.steps}
			<!-- Play mode -->
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
									<button class="next-btn" onclick={nextStep}>
										{i === routine.steps.length - 1 ? '✓ Finish' : 'Next →'}
									</button>
									{#if step.optional}
										<button class="skip-btn" onclick={skipStep}>Skip</button>
									{/if}
								</div>
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
				<button class="stop-btn" onclick={stopPlay}>Stop</button>
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

	.back-link {
		font-size: 0.8rem; color: var(--text-secondary); text-decoration: none;
		display: inline-block; margin-bottom: 0.5rem;
	}

	.back-link:hover { color: var(--accent-primary); }

	h1 { font-size: 1.5rem; color: var(--text-primary); margin-bottom: 0.25rem; }
	h2 { font-size: 1.3rem; color: var(--text-primary); margin-bottom: 0.75rem; }

	.step-count { font-size: 0.8rem; color: var(--text-secondary); }

	.error {
		padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning);
		border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem;
	}

	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.step-list { list-style: none; display: flex; flex-direction: column; gap: 0.4rem; margin-bottom: 1.5rem; }

	.step-item {
		display: flex; gap: 0.75rem; padding: 0.8rem 1rem;
		background: var(--bg-surface); border: 1px solid var(--border-default);
		border-radius: var(--radius-card); box-shadow: var(--shadow-card);
		transition: border-color 0.2s, opacity 0.2s;
	}

	.step-item.active {
		border-color: var(--accent-primary);
		box-shadow: var(--shadow-elevated);
	}

	.step-item.done { opacity: 0.5; }

	.step-num {
		min-width: 1.5rem; text-align: center; font-weight: 600;
		color: var(--accent-primary); font-size: 0.9rem; margin-top: 0.1rem;
	}

	.step-item.done .step-num { color: var(--success); }

	.step-body { flex: 1; }

	.step-title { font-size: 0.95rem; font-weight: 500; color: var(--text-primary); display: block; }

	.step-meta { display: flex; gap: 0.5rem; margin-top: 0.25rem; }

	.step-duration { font-size: 0.75rem; color: var(--text-secondary); }

	.step-optional {
		font-size: 0.65rem; color: var(--text-tertiary); background: var(--bg-elevated);
		padding: 0.1rem 0.35rem; border-radius: 3px;
	}

	/* Timer */
	.step-timer { margin-top: 0.5rem; display: flex; align-items: center; gap: 0.5rem; flex-wrap: wrap; }

	.timer-value { font-size: 1.5rem; font-weight: 300; color: var(--accent-primary); font-variant-numeric: tabular-nums; }

	.timer-target { font-size: 0.85rem; color: var(--text-tertiary); }

	.progress-bar {
		width: 100%; height: 4px; background: var(--score-track);
		border-radius: 2px; overflow: hidden; margin-top: 0.3rem;
	}

	.progress-fill {
		height: 100%; background: var(--accent-primary); border-radius: 2px;
		transition: width 1s linear;
	}

	.step-actions { display: flex; gap: 0.5rem; margin-top: 0.5rem; }

	.next-btn {
		padding: 0.4rem 0.8rem; background: var(--accent-primary); color: var(--text-on-accent);
		border: none; border-radius: var(--radius-button); cursor: pointer; font-size: 0.85rem; font-weight: 500;
	}

	.skip-btn {
		padding: 0.4rem 0.8rem; background: var(--bg-elevated); color: var(--text-secondary);
		border: 1px solid var(--border-default); border-radius: var(--radius-button); cursor: pointer; font-size: 0.85rem;
	}

	.play-btn {
		padding: 0.6rem 1.5rem; background: var(--accent-primary); color: var(--text-on-accent);
		border: none; border-radius: var(--radius-button); cursor: pointer; font-size: 0.95rem; font-weight: 500;
	}

	.play-footer {
		display: flex; justify-content: space-between; align-items: center;
		padding-top: 1rem; border-top: 1px solid var(--border-default);
	}

	.total-time { font-size: 0.85rem; color: var(--text-secondary); }

	.stop-btn {
		padding: 0.4rem 0.8rem; background: var(--warning-subtle); color: var(--warning-text);
		border: 1px solid var(--warning); border-radius: var(--radius-button); cursor: pointer; font-size: 0.8rem;
	}

	/* Summary */
	.summary { text-align: center; padding: 2rem 0; }

	.summary-time { font-size: 1.2rem; color: var(--accent-primary); margin-bottom: 1.5rem; }

	.summary-steps { list-style: none; text-align: left; max-width: 400px; margin: 0 auto 1.5rem; }

	.summary-step { padding: 0.3rem 0; font-size: 0.9rem; color: var(--text-secondary); }

	.summary-check { margin-right: 0.5rem; color: var(--success); }

	.summary-back { display: inline-block; margin-top: 1rem; }
</style>
