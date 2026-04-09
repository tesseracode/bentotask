<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children }: { children: Snippet } = $props();

	const navItems = [
		{ href: '/', label: 'Inbox', icon: '📥' },
		{ href: '/today', label: 'Today', icon: '📅' },
		{ href: '/habits', label: 'Habits', icon: '🔥' },
	];
</script>

<svelte:head>
	<title>BentoTask</title>
	<link rel="icon" href="/favicon.svg" type="image/svg+xml" />
</svelte:head>

<div class="app">
	<nav class="sidebar">
		<div class="logo">
			<span class="logo-icon">🍱</span>
			<span class="logo-text">BentoTask</span>
		</div>
		<ul class="nav-list">
			{#each navItems as item}
				<li>
					<a href={item.href} class="nav-link">
						<span class="nav-icon">{item.icon}</span>
						<span class="nav-label">{item.label}</span>
					</a>
				</li>
			{/each}
		</ul>
	</nav>
	<main class="content">
		{@render children()}
	</main>
</div>

<style>
	:global(*, *::before, *::after) {
		box-sizing: border-box;
		margin: 0;
		padding: 0;
	}

	:global(body) {
		font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
		background: #0f0f0f;
		color: #e0e0e0;
		line-height: 1.5;
	}

	.app {
		display: flex;
		min-height: 100vh;
	}

	.sidebar {
		width: 220px;
		background: #1a1a1a;
		border-right: 1px solid #2a2a2a;
		padding: 1rem 0;
		display: flex;
		flex-direction: column;
		flex-shrink: 0;
	}

	.logo {
		display: flex;
		align-items: center;
		gap: 0.5rem;
		padding: 0.5rem 1.25rem 1.5rem;
		font-size: 1.1rem;
		font-weight: 600;
	}

	.logo-icon {
		font-size: 1.4rem;
	}

	.nav-list {
		list-style: none;
	}

	.nav-link {
		display: flex;
		align-items: center;
		gap: 0.6rem;
		padding: 0.6rem 1.25rem;
		color: #999;
		text-decoration: none;
		font-size: 0.9rem;
		transition: background 0.15s, color 0.15s;
	}

	.nav-link:hover {
		background: #252525;
		color: #fff;
	}

	.nav-icon {
		font-size: 1.1rem;
		width: 1.5rem;
		text-align: center;
	}

	.content {
		flex: 1;
		padding: 2rem;
		max-width: 900px;
	}
</style>
