<script lang="ts">
	import { onMount } from 'svelte';
	import type { GPSPoint } from '$lib/api/client';
	import L from 'leaflet';
	import 'leaflet/dist/leaflet.css';

	let { points }: { points: GPSPoint[] } = $props();

	let mapContainer: HTMLDivElement;
	let map: L.Map | undefined;

	onMount(() => {
		map = L.map(mapContainer).setView([55.75, 37.61], 12);

		L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
			attribution: '&copy; OpenStreetMap contributors'
		}).addTo(map);

		return () => {
			map?.remove();
		};
	});

	$effect(() => {
		if (!map || points.length === 0) return;

		map.eachLayer((layer) => {
			if (layer instanceof L.Polyline || layer instanceof L.CircleMarker) {
				map!.removeLayer(layer);
			}
		});

		const latlngs = points.map((p) => [p.lat, p.lon] as L.LatLngTuple);
		const polyline = L.polyline(latlngs, { color: '#3b82f6', weight: 3 }).addTo(map);
		map.fitBounds(polyline.getBounds(), { padding: [30, 30] });

		const first = points[0];
		const last = points[points.length - 1];
		L.circleMarker([first.lat, first.lon], { radius: 6, color: '#22c55e', fillOpacity: 1 })
			.bindPopup('Start')
			.addTo(map);
		L.circleMarker([last.lat, last.lon], { radius: 6, color: '#ef4444', fillOpacity: 1 })
			.bindPopup('Finish')
			.addTo(map);
	});
</script>

<div bind:this={mapContainer} class="h-full w-full rounded-lg"></div>
