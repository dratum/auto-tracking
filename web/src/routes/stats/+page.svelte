<script lang="ts">
	import { onMount } from 'svelte';
	import { getStats, type Stats } from '$lib/api/client';

	let stats = $state<Stats | null>(null);
	let period = $state('week');
	let loading = $state(true);
	let error = $state('');

	const periods = ['day', 'week', 'month', 'year'];

	async function load() {
		loading = true;
		error = '';
		try {
			stats = await getStats(period);
		} catch (e) {
			error = 'Failed to load stats';
		} finally {
			loading = false;
		}
	}

	onMount(load);

	function formatDuration(min: number): string {
		const h = Math.floor(min / 60);
		const m = Math.round(min % 60);
		return h > 0 ? `${h}h ${m}m` : `${m}m`;
	}
</script>

<h1 class="mb-4 text-xl font-semibold text-gray-900">Stats</h1>

<div class="mb-4 flex gap-2">
	{#each periods as p}
		<button
			onclick={() => { period = p; load(); }}
			class="rounded-lg px-3 py-1.5 text-sm"
			class:bg-blue-600={period === p}
			class:text-white={period === p}
			class:bg-white={period !== p}
			class:text-gray-700={period !== p}
			class:shadow={period !== p}
		>
			{p}
		</button>
	{/each}
</div>

{#if loading}
	<p class="text-gray-500">Loading...</p>
{:else if error}
	<p class="text-red-500">{error}</p>
{:else if stats}
	<div class="grid grid-cols-2 gap-4 sm:grid-cols-4">
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Total distance</p>
			<p class="text-2xl font-semibold">{stats.total_distance_km.toFixed(1)} km</p>
		</div>
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Total trips</p>
			<p class="text-2xl font-semibold">{stats.total_trips}</p>
		</div>
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Total time</p>
			<p class="text-2xl font-semibold">{formatDuration(stats.total_duration_min)}</p>
		</div>
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Avg trip</p>
			<p class="text-2xl font-semibold">{stats.avg_trip_distance_km.toFixed(1)} km</p>
		</div>
	</div>
{/if}
