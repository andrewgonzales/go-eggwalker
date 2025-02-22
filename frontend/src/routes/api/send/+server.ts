import { API_HOST } from '$env/static/private';
import type { RequestHandler } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request }) => {
	const { text, direction, position } = await request.json();

	const response = await fetch(`${API_HOST}/send-${direction}`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ text, position })
	});

	const result = await response.json();

	return new Response(JSON.stringify(result), {
		status: response.status,
		headers: {
			'Content-Type': 'application/json'
		}
	});
};
