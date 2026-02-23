<script lang="ts">
	import { Chessground } from 'svelte5-chessground';
	import 'svelte5-chessground/style.css';

	let fen = 'rnbqkbnr/pppppppp/8/8/8/8/PPPPPPPP/RNBQKBNR w KQkq - 0 1';

	function handleMove(from: string, to: string) {
		console.log('Move:', from, 'to', to);
	}
</script>

<div class="container">
    <Chessground {fen} onMove={handleMove} />
</div>

<style>
	.container :global(.cg-wrap) {
		width: 512px;
		height: 512px;
	}
</style>
