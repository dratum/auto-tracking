const BASE = '/api/v1';

async function request<T>(path: string, options: RequestInit = {}): Promise<T> {
	const token = localStorage.getItem('token');

	const headers: Record<string, string> = {
		...Object.fromEntries(new Headers(options.headers).entries())
	};
	if (token) {
		headers['Authorization'] = `Bearer ${token}`;
	}

	const res = await fetch(`${BASE}${path}`, { ...options, headers });

	if (res.status === 401) {
		localStorage.removeItem('token');
		window.location.href = '/login';
		throw new Error('Unauthorized');
	}

	if (!res.ok) {
		throw new Error(`${res.status} ${res.statusText}`);
	}

	return res.json();
}

export interface Trip {
	id: string;
	start_time: string;
	end_time: string;
	distance_km: number;
	duration_min: number;
	max_speed: number;
	avg_speed: number;
	status: string;
}

export interface TripsResponse {
	trips: Trip[];
	total: number;
	page: number;
	limit: number;
}

export interface GPSPoint {
	lat: number;
	lon: number;
	speed: number;
	time: string;
}

export interface PointsResponse {
	points: GPSPoint[];
}

export interface Stats {
	period: string;
	total_distance_km: number;
	total_trips: number;
	total_duration_min: number;
	avg_trip_distance_km: number;
}

export interface LoginResponse {
	token: string;
	expires_at: string;
}

export async function login(username: string, password: string): Promise<LoginResponse> {
	return request<LoginResponse>('/auth/login', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		body: JSON.stringify({ username, password })
	});
}

export async function getTrips(page = 1, limit = 20): Promise<TripsResponse> {
	return request<TripsResponse>(`/trips?page=${page}&limit=${limit}`);
}

export async function getTrip(id: string): Promise<Trip> {
	return request<Trip>(`/trips/${id}`);
}

export async function getTripPoints(id: string): Promise<PointsResponse> {
	return request<PointsResponse>(`/trips/${id}/points`);
}

export async function getStats(period: string = 'week'): Promise<Stats> {
	return request<Stats>(`/stats?period=${period}`);
}
