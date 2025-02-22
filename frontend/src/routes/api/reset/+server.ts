import { API_HOST } from '$env/static/private';
import type { RequestHandler } from '@sveltejs/kit';

export const POST: RequestHandler = async () => {
	const response = await fetch(`${API_HOST}/reset`, {
		method: 'POST'
	});

	const result = await response.json();

	return new Response(JSON.stringify(result), {
		status: response.status,
		headers: {
			'Content-Type': 'application/json'
		}
	});
};
