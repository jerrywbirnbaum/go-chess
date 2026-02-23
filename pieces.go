package main

type PieceType int

const (
	EmptyPieceType PieceType = iota
	Pawn
	Bishop
	Knight
	Rook
	Queen
	King
)

type Piece int

const (
	EmptyPiece Piece = iota
	WhitePawn
	WhiteBishop
	WhiteKnight
	WhiteRook
	WhiteQueen
	WhiteKing
	BlackPawn
	BlackBishop
	BlackKnight
	BlackRook
	BlackQueen
	BlackKing
)

func isSlidingPiece(pt PieceType) bool {
	switch pt {
	case Queen:
	case Rook:
	case Bishop:
		return true
	}
	return false
}
func pieceType(p Piece) PieceType {
	switch p {
	case EmptyPiece:
		return EmptyPieceType
	case WhitePawn:
	case BlackPawn:
		return Pawn
	case WhiteBishop:
	case BlackBishop:
		return Bishop
	case WhiteKnight:
	case BlackKnight:
		return Knight
	case WhiteRook:
	case BlackRook:
		return Rook
	case WhiteQueen:
	case BlackQueen:
		return Queen
	case WhiteKing:
	case BlackKing:
		return King
	}
	return EmptyPieceType
}
func newPiece(r rune) Piece {
	switch r {
	case 'P':
		return WhitePawn
	case 'B':
		return WhiteBishop
	case 'N':
		return WhiteKnight
	case 'R':
		return WhiteRook
	case 'Q':
		return WhiteQueen
	case 'K':
		return WhiteKing
	case 'p':
		return BlackPawn
	case 'b':
		return BlackBishop
	case 'n':
		return BlackKnight
	case 'r':
		return BlackRook
	case 'q':
		return BlackQueen
	case 'k':
		return BlackKing
	default:
		return EmptyPiece
	}
}
func printPiece(p Piece) rune {
	switch p {
	case EmptyPiece:
		return '*'
	case WhitePawn:
		return 'P'
	case WhiteBishop:
		return 'B'
	case WhiteKnight:
		return 'N'
	case WhiteRook:
		return 'R'
	case WhiteQueen:
		return 'Q'
	case WhiteKing:
		return 'K'
	case BlackPawn:
		return 'p'
	case BlackBishop:
		return 'b'
	case BlackKnight:
		return 'n'
	case BlackRook:
		return 'r'
	case BlackQueen:
		return 'q'
	case BlackKing:
		return 'k'
	}
	return '*'
}
