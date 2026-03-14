import { writable } from 'svelte/store';

export const isAuthenticated = writable(false);

export function checkAuth(): boolean {
	const token = localStorage.getItem('token');
	const hasToken = !!token;
	isAuthenticated.set(hasToken);
	return hasToken;
}

export function logout() {
	localStorage.removeItem('token');
	isAuthenticated.set(false);
	window.location.href = '/login';
}
