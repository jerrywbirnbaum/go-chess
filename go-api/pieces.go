package main

type Color int

const (
	White Color = iota
	Black
)

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
	case Queen, Rook, Bishop:
		return true
	}
	return false
}

func isDiagonalSlidingPiece(pt PieceType) bool {
	switch pt {
	case Queen, Bishop:
		return true
	}
	return false
}

func isStriaghtSlidingPiece(pt PieceType) bool {
	switch pt {
	case Queen, Rook:
		return true
	}
	return false
}

func getColor(p Piece) Color {
	switch p {

	case WhitePawn, WhiteBishop, WhiteKnight, WhiteRook, WhiteQueen, WhiteKing:
		return Color(White)

	case BlackPawn, BlackBishop, BlackKnight, BlackRook, BlackQueen, BlackKing:
		return Color(Black)
	}
	return Color(Black)
}
func isWhite(p Piece) bool {
	switch p {

	case WhitePawn, WhiteBishop, WhiteKnight, WhiteRook, WhiteQueen, WhiteKing:
		return true

	case BlackPawn, BlackBishop, BlackKnight, BlackRook, BlackQueen, BlackKing:
		return false
	}
	return false
}

func isPawn(pt PieceType) bool {
	switch pt {
	case Pawn:
		return true
	}
	return false
}

func isKnight(pt PieceType) bool {
	switch pt {
	case Knight:
		return true
	}
	return false
}

func isRook(pt PieceType) bool {
	switch pt {
	case Rook:
		return true
	}
	return false
}

func isBishop(pt PieceType) bool {
	switch pt {
	case Bishop:
		return true
	}
	return false
}

func isQueen(pt PieceType) bool {
	switch pt {
	case Queen:
		return true
	}
	return false
}
func isKing(pt PieceType) bool {
	switch pt {
	case King:
		return true
	}
	return false
}

func isEmpty(p Piece) bool {
	switch p {
	case EmptyPiece:
		return true
	}
	return false
}

func sameColor(p Piece, c Color) bool {
	switch p {
	case EmptyPiece:
		return false
	case WhitePawn, WhiteBishop, WhiteKnight, WhiteRook, WhiteQueen, WhiteKing:
		switch c {
		case White:
			return true
		}
	case BlackPawn, BlackBishop, BlackKnight, BlackRook, BlackQueen, BlackKing:
		switch c {
		case Black:
			return true
		}
	}
	return false

}

func pieceType(p Piece) PieceType {
	switch p {
	case EmptyPiece:
		return EmptyPieceType
	case WhitePawn, BlackPawn:
		return Pawn
	case WhiteBishop, BlackBishop:
		return Bishop
	case WhiteKnight, BlackKnight:
		return Knight
	case WhiteRook, BlackRook:
		return Rook
	case WhiteQueen, BlackQueen:
		return Queen
	case WhiteKing, BlackKing:
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

func oppositeColor(c Color) Color {
	switch c {
	case White:
		return Black
	case Black:
		return White
	}
	return White
}

func newPieceTypeColor(pt PieceType, c Color) Piece {
	switch c {
	case White:
		switch pt {
		case Pawn:
			return WhitePawn
		case Knight:
			return WhiteKnight
		case Bishop:
			return WhiteBishop
		case Rook:
			return WhiteRook
		case Queen:
			return WhiteQueen
		case King:
			return WhiteKing
		}
	case Black:
		switch pt {
		case Pawn:
			return BlackPawn
		case Knight:
			return BlackKnight
		case Bishop:
			return BlackBishop
		case Rook:
			return BlackRook
		case Queen:
			return BlackQueen
		case King:
			return BlackKing
		}
	}
	return EmptyPiece
}
