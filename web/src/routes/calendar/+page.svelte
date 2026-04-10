<script lang="ts">
	import { onMount } from 'svelte';
	import { tasks, type TaskJSON } from '$lib/api';

	let allTasks: TaskJSON[] = $state([]);
	let loading = $state(true);
	let error = $state('');
	let currentMonth = $state(new Date());

	const dayNames = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];

	let monthLabel = $derived(currentMonth.toLocaleDateString(undefined, { month: 'long', year: 'numeric' }));

	interface CalendarCell {
		date: Date;
		dayNum: number;
		isCurrentMonth: boolean;
		isToday: boolean;
		tasks: TaskJSON[];
	}

	let cells = $derived.by((): CalendarCell[] => {
		const year = currentMonth.getFullYear();
		const month = currentMonth.getMonth();
		const firstDay = new Date(year, month, 1);
		const lastDay = new Date(year, month + 1, 0);
		const today = new Date();
		const todayStr = `${today.getFullYear()}-${String(today.getMonth() + 1).padStart(2, '0')}-${String(today.getDate()).padStart(2, '0')}`;

		// Monday-based week: getDay() returns 0=Sun, we want 0=Mon
		let startOffset = firstDay.getDay() - 1;
		if (startOffset < 0) startOffset = 6;

		const result: CalendarCell[] = [];

		// Previous month padding
		for (let i = startOffset - 1; i >= 0; i--) {
			const d = new Date(year, month, -i);
			result.push({ date: d, dayNum: d.getDate(), isCurrentMonth: false, isToday: false, tasks: [] });
		}

		// Current month days
		for (let day = 1; day <= lastDay.getDate(); day++) {
			const d = new Date(year, month, day);
			const dateStr = `${year}-${String(month + 1).padStart(2, '0')}-${String(day).padStart(2, '0')}`;
			const dayTasks = allTasks.filter((t) => t.due_date === dateStr);
			result.push({
				date: d,
				dayNum: day,
				isCurrentMonth: true,
				isToday: dateStr === todayStr,
				tasks: dayTasks
			});
		}

		// Next month padding to fill grid (6 rows × 7)
		while (result.length < 42) {
			const d = new Date(year, month + 1, result.length - startOffset - lastDay.getDate() + 1);
			result.push({ date: d, dayNum: d.getDate(), isCurrentMonth: false, isToday: false, tasks: [] });
		}

		return result;
	});

	function prevMonth() {
		currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() - 1, 1);
	}

	function nextMonth() {
		currentMonth = new Date(currentMonth.getFullYear(), currentMonth.getMonth() + 1, 1);
	}

	function goToday() {
		currentMonth = new Date();
	}

	function priorityClass(p?: string): string {
		if (!p || p === 'none') return '';
		return `pill-${p}`;
	}

	async function loadTasks() {
		error = '';
		try {
			const res = await tasks.list();
			allTasks = res.items.filter((t) => t.due_date);
		} catch (e) {
			error = e instanceof Error ? e.message : 'Failed to load tasks';
		} finally {
			loading = false;
		}
	}

	onMount(() => { loadTasks(); });
</script>

<div class="view">
	<div class="cal-header">
		<div class="cal-nav">
			<button onclick={prevMonth} title="Previous month">&larr;</button>
			<h1>{monthLabel}</h1>
			<button onclick={nextMonth} title="Next month">&rarr;</button>
		</div>
		<button class="today-btn" onclick={goToday}>Today</button>
	</div>

	{#if error}
		<div class="error">{error}</div>
	{/if}

	{#if loading}
		<p class="empty">Loading...</p>
	{:else}
		<div class="cal-grid">
			{#each dayNames as day}
				<div class="cal-day-header">{day}</div>
			{/each}
			{#each cells as cell}
				<div class="cal-cell" class:other-month={!cell.isCurrentMonth} class:today={cell.isToday}>
					<span class="cal-day-num">{cell.dayNum}</span>
					{#each cell.tasks.slice(0, 3) as task}
						<div class="task-pill {priorityClass(task.priority)}" title="{task.title} ({task.priority ?? 'no priority'})">
							{task.title}
						</div>
					{/each}
					{#if cell.tasks.length > 3}
						<span class="more-tasks">+{cell.tasks.length - 3} more</span>
					{/if}
				</div>
			{/each}
		</div>
	{/if}
</div>

<style>
	.view { max-width: 900px; }

	.cal-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 1.5rem; }
	.cal-nav { display: flex; align-items: center; gap: 0.75rem; }
	.cal-nav button { background: var(--bg-elevated); border: 1px solid var(--border-default); border-radius: var(--radius-badge); color: var(--text-secondary); cursor: pointer; padding: 0.35rem 0.6rem; font-size: 0.9rem; }
	.cal-nav button:hover { border-color: var(--accent-primary); color: var(--text-primary); }
	h1 { font-size: 1.3rem; color: var(--text-primary); margin: 0; min-width: 200px; text-align: center; }
	.today-btn { background: var(--accent-primary); color: var(--text-on-accent); border: none; border-radius: var(--radius-button); padding: 0.4rem 0.8rem; cursor: pointer; font-size: 0.8rem; }

	.error { padding: 0.6rem; background: var(--warning-subtle); border: 1px solid var(--warning); border-radius: var(--radius-badge); color: var(--warning-text); margin-bottom: 1rem; font-size: 0.85rem; }
	.empty { color: var(--text-tertiary); text-align: center; padding: 3rem; }

	.cal-grid { display: grid; grid-template-columns: repeat(7, 1fr); gap: 1px; background: var(--border-default); border: 1px solid var(--border-default); border-radius: var(--radius-card); overflow: hidden; }

	.cal-day-header { background: var(--bg-elevated); padding: 0.4rem; text-align: center; font-size: 0.7rem; font-weight: 600; color: var(--text-secondary); text-transform: uppercase; letter-spacing: 0.05em; }

	.cal-cell { background: var(--bg-surface); min-height: 80px; padding: 0.3rem; display: flex; flex-direction: column; gap: 0.15rem; }
	.cal-cell.other-month { opacity: 0.35; }
	.cal-cell.today { border: 2px solid var(--accent-primary); border-radius: 2px; }

	.cal-day-num { font-size: 0.7rem; font-weight: 600; color: var(--text-secondary); margin-bottom: 0.1rem; }
	.cal-cell.today .cal-day-num { color: var(--accent-primary); }

	.task-pill { font-size: 0.6rem; padding: 0.1rem 0.3rem; border-radius: 3px; background: var(--bg-elevated); color: var(--text-secondary); overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

	.pill-urgent { background: var(--priority-urgent-bg); color: var(--priority-urgent-text); }
	.pill-high { background: var(--priority-high-bg); color: var(--priority-high-text); }
	.pill-medium { background: var(--priority-medium-bg); color: var(--priority-medium-text); }
	.pill-low { background: var(--priority-low-bg); color: var(--priority-low-text); }

	.more-tasks { font-size: 0.55rem; color: var(--text-tertiary); padding: 0 0.2rem; }
</style>
