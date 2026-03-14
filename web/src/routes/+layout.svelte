<script lang="ts">
	import favicon from '$lib/assets/favicon.svg';
	import { isAuthenticated, checkAuth, logout } from '$lib/stores/auth';
	import { onMount } from 'svelte';
	import { goto } from '$app/navigation';
	import { page } from '$app/state';
	import '../app.css';

	let { children } = $props();

	onMount(() => {
		const isLogin = page.url.pathname === '/login';
		if (!checkAuth() && !isLogin) {
			goto('/login');
		}
	});
</script>

<svelte:head>
	<link rel="icon" href={favicon} />
</svelte:head>

{#if $isAuthenticated}
	<div class="flex h-screen flex-col bg-gray-50">
		<nav class="border-b border-gray-200 bg-white px-6 py-3">
			<div class="flex items-center justify-between">
				<div class="flex items-center gap-6">
					<a href="/" class="text-lg font-semibold text-gray-900">Auto-Tracking</a>
					<a
						href="/"
						class="text-sm text-gray-600 hover:text-gray-900"
						class:font-medium={page.url.pathname === '/'}
					>
						Trips
					</a>
					<a
						href="/stats"
						class="text-sm text-gray-600 hover:text-gray-900"
						class:font-medium={page.url.pathname === '/stats'}
					>
						Stats
					</a>
				</div>
				<button onclick={logout} class="text-sm text-gray-500 hover:text-gray-700">
					Logout
				</button>
			</div>
		</nav>
		<main class="flex-1 overflow-auto p-6">
			{@render children()}
		</main>
	</div>
{:else}
	{@render children()}
{/if}
