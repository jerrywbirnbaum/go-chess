<script lang="ts">
	import { Chessground } from "svelte5-chessground";
	import { Chess } from "chess.js";
	import "svelte5-chessground/style.css";

	let game = $state(new Chess());
	const fen = $derived(game.fen());
	const turn = $derived(game.turn() === "w" ? "white" : "black");

	// Calculate legal moves for Chessground

	const dests = $derived.by(() => {
		const map = new Map();
		game.moves({ verbose: true }).forEach((m) => {
			if (!map.has(m.from)) map.set(m.from, []);
			map.get(m.from).push(m.to);
		});
		return map;
	});
	async function handleMove(from: string, to: string) {
		try {
			const newGame = new Chess(game.fen());
			newGame.move({ from, to, promotion: "q" });
			game = newGame;
			console.log("Move", game.fen());
			console.log("Turn", game.turn());
		} catch (e) {
			game = new Chess(game.fen());
			console.log("Illegal move", game.fen());
			console.log("Turn", game.turn());
			return;
		}
		await engineMove();
	}

	async function engineMove() {
		const currentFen = game.fen();
		try {
			const response = await fetch(
				"http://localhost:8080/generate-moves",
				{
					method: "POST",
					headers: {
						"Content-Type":
							"application/json",
					},
					body: JSON.stringify({
						fen: currentFen,
					}),
				},
			);
			const data = await response.json();
			console.log(data);

			const newGame = new Chess(game.fen());
			newGame.move({
				from: data.start_square,
				to: data.end_square,
				promotion: "q",
			});
			game = newGame;
		} catch (error) {
			console.log(error);
		}
	}
</script>

<div class="container">
	<Chessground
		{fen}
		onMove={handleMove}
		config={{
			fen: fen,
			turnColor: turn,
			movable: {
				color: turn,
				dests: dests, // This prevents illegal moves entirely
			},
		}}
	/>
</div>

<style>
	.container :global(.cg-wrap) {
		width: 512px;
		height: 512px;
	}
</style>
