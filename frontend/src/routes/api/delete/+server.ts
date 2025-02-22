import { API_HOST } from '$env/static/private';
import type { RequestHandler } from '@sveltejs/kit';

export const POST: RequestHandler = async ({ request }) => {
	const { agent, position, numDeletions } = await request.json();

	const response = await fetch(`${API_HOST}/delete`, {
		method: 'POST',
		headers: {
			'Content-Type': 'application/json'
		},
		body: JSON.stringify({ agent, position, numDeletions })
	});

	const result = await response.json();

	return new Response(JSON.stringify(result), {
		status: response.status,
		headers: {
			'Content-Type': 'application/json'
		}
	});
};
