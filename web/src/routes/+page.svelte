<script lang="ts">
	import { onMount } from 'svelte';
	import { getTrips, type Trip } from '$lib/api/client';

	let trips = $state<Trip[]>([]);
	let loading = $state(true);
	let error = $state('');

	onMount(async () => {
		try {
			const res = await getTrips();
			trips = res.trips ?? [];
		} catch (e) {
			error = 'Failed to load trips';
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

<h1 class="mb-4 text-xl font-semibold text-gray-900">Trips</h1>

{#if loading}
	<p class="text-gray-500">Loading...</p>
{:else if error}
	<p class="text-red-500">{error}</p>
{:else if trips.length === 0}
	<p class="text-gray-500">No trips yet</p>
{:else}
	<div class="overflow-hidden rounded-lg bg-white shadow">
		<table class="w-full text-left text-sm">
			<thead class="border-b bg-gray-50 text-gray-600">
				<tr>
					<th class="px-4 py-3">Date</th>
					<th class="px-4 py-3">Duration</th>
					<th class="px-4 py-3">Distance</th>
					<th class="px-4 py-3">Avg speed</th>
					<th class="px-4 py-3">Status</th>
				</tr>
			</thead>
			<tbody>
				{#each trips as trip}
					<tr class="border-b last:border-0 hover:bg-gray-50">
						<td class="px-4 py-3">
							<a href="/trips/{trip.id}" class="text-blue-600 hover:underline">
								{formatDate(trip.start_time)}
							</a>
						</td>
						<td class="px-4 py-3">{formatDuration(trip.duration_min)}</td>
						<td class="px-4 py-3">{trip.distance_km.toFixed(1)} km</td>
						<td class="px-4 py-3">{trip.avg_speed.toFixed(0)} km/h</td>
						<td class="px-4 py-3">
							<span
								class="rounded-full px-2 py-0.5 text-xs font-medium"
								class:bg-green-100={trip.status === 'completed'}
								class:text-green-700={trip.status === 'completed'}
								class:bg-yellow-100={trip.status === 'active'}
								class:text-yellow-700={trip.status === 'active'}
							>
								{trip.status}
							</span>
						</td>
					</tr>
				{/each}
			</tbody>
		</table>
	</div>
{/if}
