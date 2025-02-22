<script lang="ts">
	import debounce from 'lodash.debounce';
	import { calcDiff } from '../../utils/diff';

	type Direction = 'left' | 'right';

	interface TextState {
		left: string;
		right: string;
	}

	const DEFAULT_STATE: TextState = {
		left: '',
		right: ''
	};
	let text = $state({ ...DEFAULT_STATE });

	const debouncedSendText = debounce(sendText, 250);

	async function sendText(sendDirection: Direction) {
		const currentDirection = sendDirection === 'left' ? 'right' : 'left';

		const { pos, del, ins } = calcDiff(text[sendDirection], text[currentDirection]);

		if (del > 0) {
			const deleteResponse = await fetch('api/delete', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ agent: currentDirection, position: pos, numDeletions: del })
			});

			const deleteResult = await deleteResponse.json();

			text.left = deleteResult.left;
			text.right = deleteResult.right;
		}

		if (ins !== '') {
			const response = await fetch('api/send', {
				method: 'POST',
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({ text: ins, direction: sendDirection, position: pos })
			});
			const result = await response.json();
			text.left = result.left;
			text.right = result.right;
		}
	}

	async function reset() {
		const response = await fetch('api/reset', { method: 'POST' });

		if (response.ok) {
			text = { ...DEFAULT_STATE };
		}
	}
</script>

<div id="textcontainer">
	<textarea
		id="text-left"
		spellcheck="false"
		placeholder="User 1"
		bind:value={text.left}
		oninput={() => debouncedSendText('right')}
	></textarea>

	<textarea
		id="text-right"
		spellcheck="false"
		placeholder="User 2"
		bind:value={text.right}
		oninput={() => debouncedSendText('left')}
	></textarea>
</div>

<div class="button-container">
	<button id="reset" onclick={reset}>Reset</button>
</div>

<style>
	#textcontainer {
		display: flex;
		justify-content: space-between;
		margin: 20px;
	}

	textarea {
		width: 45%;
		height: 200px;
		padding: 10px;
		border: 1px solid #ccc;
		border-radius: 5px;
		font-size: 16px;
	}

	.button-container {
		display: flex;
		justify-content: center;
		margin-top: 20px;
	}

	button {
		margin: 10px;
		padding: 10px 20px;
		border: none;
		border-radius: 5px;
		background-color: #9d6a89;
		color: white;
		font-size: 16px;
		cursor: pointer;
		transition: background-color 0.3s ease;
	}

	button:hover {
		background-color: #ae849d;
	}

	button:active {
		background-color: #cfb5c4;
	}
</style>
