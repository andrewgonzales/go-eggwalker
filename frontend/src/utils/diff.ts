// https://github.com/josephg/crdt-from-scratch/blob/master/editor.ts

type DiffResult = { pos: number; del: number; ins: string };

// This is a very simple diff function. Notably it doesn't take into account
// the user's cursor position - so as you type "aaaaa", we can't tell which
// "a" you've just inserted each time.
export const calcDiff = (oldval: string, newval: string): DiffResult => {
	// Strings are immutable and have reference equality. I think this test is O(1), so its worth doing.
	if (oldval === newval) return { pos: 0, del: 0, ins: '' };

	const oldChars = [...oldval];
	const newChars = [...newval];

	let commonStart = 0;
	while (oldChars[commonStart] === newChars[commonStart]) {
		commonStart++;
	}

	let commonEnd = 0;
	while (
		oldChars[oldChars.length - 1 - commonEnd] === newChars[newChars.length - 1 - commonEnd] &&
		commonEnd + commonStart < oldChars.length &&
		commonEnd + commonStart < newChars.length
	) {
		commonEnd++;
	}

	const del =
		oldChars.length !== commonStart + commonEnd ? oldChars.length - commonStart - commonEnd : 0;
	const ins =
		newChars.length !== commonStart + commonEnd
			? newChars.slice(commonStart, newChars.length - commonEnd).join('')
			: '';

	return {
		pos: commonStart,
		del,
		ins
	};
};
