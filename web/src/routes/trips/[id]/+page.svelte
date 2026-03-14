<script lang="ts">
	import { onMount } from 'svelte';
	import { page } from '$app/state';
	import { getTrip, getTripPoints, type Trip, type GPSPoint } from '$lib/api/client';
	import Map from '$lib/components/Map.svelte';

	let trip = $state<Trip | null>(null);
	let points = $state<GPSPoint[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		const id = page.params.id;
		try {
			const [t, p] = await Promise.all([getTrip(id), getTripPoints(id)]);
			trip = t;
			points = p.points ?? [];
		} catch (e) {
			error = 'Failed to load trip';
		} finally {
			loading = false;
		}
	});

	function formatDate(iso: string): string {
		return new Date(iso).toLocaleString('ru-RU', {
			day: '2-digit',
			month: '2-digit',
			year: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function formatDuration(min: number): string {
		const h = Math.floor(min / 60);
		const m = Math.round(min % 60);
		return h > 0 ? `${h}h ${m}m` : `${m}m`;
	}
</script>

{#if loading}
	<p class="text-gray-500">Loading...</p>
{:else if error}
	<p class="text-red-500">{error}</p>
{:else if trip}
	<div class="mb-4 flex items-center gap-4">
		<a href="/" class="text-sm text-blue-600 hover:underline">&larr; Back</a>
		<h1 class="text-xl font-semibold text-gray-900">Trip {formatDate(trip.start_time)}</h1>
	</div>

	<div class="mb-4 grid grid-cols-2 gap-4 sm:grid-cols-4">
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Distance</p>
			<p class="text-lg font-semibold">{trip.distance_km.toFixed(1)} km</p>
		</div>
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Duration</p>
			<p class="text-lg font-semibold">{formatDuration(trip.duration_min)}</p>
		</div>
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Avg speed</p>
			<p class="text-lg font-semibold">{trip.avg_speed.toFixed(0)} km/h</p>
		</div>
		<div class="rounded-lg bg-white p-4 shadow">
			<p class="text-sm text-gray-500">Max speed</p>
			<p class="text-lg font-semibold">{trip.max_speed.toFixed(0)} km/h</p>
		</div>
	</div>

	<div class="h-[500px] overflow-hidden rounded-lg bg-white shadow">
		{#if points.length > 0}
			<Map {points} />
		{:else}
			<p class="flex h-full items-center justify-center text-gray-500">No GPS points</p>
		{/if}
	</div>
{/if}
