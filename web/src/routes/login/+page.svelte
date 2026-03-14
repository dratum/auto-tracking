<script lang="ts">
	import { login } from '$lib/api/client';
	import { isAuthenticated } from '$lib/stores/auth';
	import { goto } from '$app/navigation';

	let username = $state('');
	let password = $state('');
	let error = $state('');
	let loading = $state(false);

	async function handleSubmit() {
		error = '';
		loading = true;
		try {
			const res = await login(username, password);
			localStorage.setItem('token', res.token);
			isAuthenticated.set(true);
			goto('/');
		} catch (e) {
			error = 'Wrong username or password';
		} finally {
			loading = false;
		}
	}
</script>

<div class="flex min-h-screen items-center justify-center bg-gray-50">
	<div class="w-full max-w-sm rounded-lg bg-white p-8 shadow">
		<h1 class="mb-6 text-center text-2xl font-semibold text-gray-900">Auto-Tracking</h1>

		<form onsubmit={handleSubmit}>
			<div class="mb-4">
				<label for="username" class="mb-1 block text-sm text-gray-700">Username</label>
				<input
					id="username"
					type="text"
					bind:value={username}
					class="w-full rounded border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
					required
				/>
			</div>

			<div class="mb-6">
				<label for="password" class="mb-1 block text-sm text-gray-700">Password</label>
				<input
					id="password"
					type="password"
					bind:value={password}
					class="w-full rounded border border-gray-300 px-3 py-2 focus:border-blue-500 focus:outline-none"
					required
				/>
			</div>

			{#if error}
				<p class="mb-4 text-sm text-red-500">{error}</p>
			{/if}

			<button
				type="submit"
				disabled={loading}
				class="w-full rounded bg-blue-600 px-4 py-2 text-white hover:bg-blue-700 disabled:opacity-50"
			>
				{loading ? 'Logging in...' : 'Log in'}
			</button>
		</form>
	</div>
</div>
