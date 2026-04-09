package SchedRare

type Bucket struct {
	head *PieceItem
	tail *PieceItem
}

type PieceItem struct {
	Index        int
	Availability uint32
	InFlight     uint32
	Bucket       int

	prev *PieceItem
	next *PieceItem
}

func (bucket *Bucket) Insert(piece *PieceItem) {
	if bucket.head == nil {
		bucket.head = piece
	} else {
		bucket.tail.next = piece
		piece.prev = bucket.tail
	}
	piece.next = nil
	bucket.tail = piece
}

func (bucket *Bucket) Remove(piece *PieceItem) {
	// assert the piece is in this bucket -> rarityQueue ensures this
	if bucket.head == piece {
		next := piece.next
		bucket.head = next
		if next != nil {
			next.prev = nil
		}
	}
	if bucket.tail == piece {
		prev := piece.prev
		bucket.tail = prev
		if prev != nil {
			prev.next = nil
		}
	}
	piece.next = nil
	piece.prev = nil
}

func (bucket *Bucket) GetPiece(bitfield *[]bool) int {
	curr := bucket.head
	for curr != nil {
		if (*bitfield)[curr.Index] {
			return curr.Index
		}
		curr = curr.next
	}
	return -1
}
