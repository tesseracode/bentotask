<script lang="ts">
	import type { Snippet } from 'svelte';
	import '$lib/theme.css';
	let { children }: { children: Snippet } = $props();
</script>

<div class="mirror-shell">
	{@render children()}
</div>

<style>
	:global(body) {
		margin: 0;
		padding: 0;
	}

	.mirror-shell {
		background: #000;
		color: #fff;
		min-height: 100vh;
		padding: 2rem 3rem;
		font-family: var(--font-body);
	}
</style>
