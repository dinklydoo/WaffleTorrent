package SchedRare

import (
	"WaffleTorrent/pkg/WaffleTorrent/Comm"
	"math"
)

/*
not really a queue, divide pieceItems into buckets
each bucket covers a range of rarity's based on rarity heuristic

THIS DOES NOT NEED TO BE THREAD SAFE, SCHEDULER HAS SINGLE POINT OF ACCESS !!!

Implement each bucket as a doubly linked-list of Piece Items:
	- Ensures Fast Insertion/Deletion O(1)
		- 	Pure Updates do not directly affect the bucket of the piece (ie will not be reinserted to a new bucket)
			Instead they only change the piece items fields for holders/inflight requests

		-	On Piece Retrieval or operations that require the bucket of a piece it will revalidate the position of the piece,
			if the piece does not belong to this bucket it will perform the reinsertion, otherwise it continues the operation
			(Lazy update of buckets )

	-	To prevent duplicate requests, if after requesting a piece falls into the same bucket
		reinsert to the back

	- 	Iteration will still be capped at O(n) when finding a piece for a peer,
		O(n) regardless even if we were to pq it or iterate over pieces bitfield so not much optimization
		strict ordering is not required as peers are not guaranteed to have the rarest K elements so amortize O(n)
		we just want loose ordering with efficient updates and hopefully a very slightly better search than O(n)
*/

const BucketSize = 5

var rarityQueue *RarityQueue // global instance -- don't need to attach lifespan to scheduler as the program dies with the scheduler

type RarityQueue struct {
	buckets []*Bucket
	items   []*PieceItem
	peers   uint32
}

func NewRarityQueue(pieces int) *RarityQueue {
	rq := RarityQueue{
		buckets: make([]*Bucket, BucketSize),
		items:   make([]*PieceItem, pieces),
	}
	for i := range rq.buckets {
		rq.buckets[i] = &Bucket{
			head: nil,
			tail: nil,
		}
	}
	for i := range pieces {
		rq.items[i] = &PieceItem{
			Index:        i,
			Availability: 0,
			InFlight:     0,
			Bucket:       0,
			prev:         nil,
			next:         nil,
		}
		rq.buckets[0].Insert(rq.items[i]) // all start with highest rarity (0 holders)
	}
	return &rq
}

func (rq *RarityQueue) AttachPeer(bitfield *[]bool) {
	rq.peers++
	for i, b := range *bitfield {
		if b {
			rq.IncHolder(i)
		}
	}
}

// rarity : we can assert this value is between [0, 1] higher being more rare
func (rq *RarityQueue) rarity(piece int) float64 {
	holders := rq.items[piece].Availability
	inflight := rq.items[piece].InFlight
	return float64(rq.peers-holders) / float64(rq.peers*(inflight+1))
}

func (rq *RarityQueue) validateBucket(piece int) bool {
	rarity := rq.rarity(piece)
	idx := int(BucketSize - min(BucketSize, 1+math.Floor(rarity*BucketSize)))
	item := rq.items[piece]
	if item.Bucket != idx { // lazy validation of a bucket get
		curr := rq.buckets[item.Bucket] // old bucket that is invalid
		curr.Remove(item)
		rq.buckets[idx].Insert(item)
		return false
	}
	return true
}

func (rq *RarityQueue) getBucket(piece int) *Bucket {
	rarity := rq.rarity(piece)
	idx := int(BucketSize - min(BucketSize, 1+math.Floor(rarity*BucketSize)))
	return rq.buckets[idx]
}

func (rq *RarityQueue) IncHolder(piece int) {
	item := rq.items[piece]
	item.Availability++
}

func (rq *RarityQueue) DecHolder(piece int) {
	item := rq.items[piece]
	item.Availability--
}

func (rq *RarityQueue) update(piece int) {
	// get old bucket
	item := rq.items[piece]
	bucket := rq.buckets[item.Bucket]
	bucket.Remove(item)
	rq.getBucket(piece).Insert(item)
}

func (rq *RarityQueue) updateRequest(piece int, incr bool) {
	// assert: piece reflects the correct bucket -> enforced by request rare candidate check
	item := rq.items[piece]
	rq.buckets[item.Bucket].Remove(item)
	if incr {
		item.InFlight++
	} else {
		item.InFlight--
	}
	rq.getBucket(piece).Insert(item)
}

func (rq *RarityQueue) RequestRare(request *Comm.PeerRequest) int {
	for _, bucket := range rq.buckets {
		candidate := bucket.GetPiece(request.Bitfield)
		if candidate < 0 { // didn't find a valid piece
			continue
		}
		if !rq.validateBucket(candidate) { // invalid piece -> revalidate
			continue
		}
		rq.updateRequest(candidate, true) // either -> place to lower bucket or replace at the end of same bucket
		return candidate
	}
	return -1
}

func (rq *RarityQueue) RequestFailed(piece int) {
	if rq.items[piece] == nil {
		return
	}
	rq.updateRequest(piece, false) // decrease inflight and reorder
}

func (rq *RarityQueue) RequestSuccess(piece int) {
	if rq.items[piece] == nil {
		return
	}
	item := rq.items[piece]
	bucket := rq.buckets[item.Bucket]
	bucket.Remove(item)
	rq.items[piece] = nil // remove this piece
}

/*
⠀⠀⠀⠀⠀⢀⣀⡤⠤⠶⠄⠠⡶⢤⠤⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣠⠞⠉⠓⠢⠄⣀⠀⠀⠱⡀⡇⠀⠉⠳⣤⠶⢦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⢀⡞⠓⠒⠒⠂⠀⠤⢀⡙⠢⡀⢱⣠⣀⠀⣴⠃⠀⡈⢧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⡎⠀⣠⠒⠒⠒⢤⡀⠀⢉⡶⠛⠉⠀⠀⠀⠁⠀⠀⡄⢸⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⢸⠃⢰⠁⠀⠀⠀⡞⠉⠲⠋⠀⠀⠀⠀⠀⠀⠀⠀⠸⠁⡼⢧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⢸⡀⠸⡆⠀⠀⢰⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡼⠁⢸⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠈⢧⠀⢧⠀⠀⣸⣧⣀⠀⠀⢀⣠⡤⢤⣤⣤⡄⠀⠀⡄⠀⠘⡇⣠⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠘⣦⠈⢣⠀⢸⠟⡋⠀⠀⠉⠁⠀⠀⠀⠉⠀⠀⢠⠃⠀⣶⣿⡏⢸⡗⣦⣀⡠⠤⠤⠤⠤⠤⣀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠈⢧⡀⠱⡌⢷⣤⡆⠀⠀⠀⠀⠀⢀⡀⠀⢀⡎⠀⣸⣿⣿⠃⣸⠀⡟⠁⠀⠀⠀⠀⠀⠀⠀⠉⠲⣄⠀⠀⠀⠀
⠀⠀⠀⠀⠳⣄⠘⢆⠈⠉⠓⠲⡖⠚⢻⠉⠀⢀⡞⠀⡰⡹⢸⡟⢠⡏⢠⣏⣉⣀⠀⠉⠑⠢⢄⡀⠀⠀⠘⢷⡀⠀⠀
⠀⠀⠀⠀⠀⠙⣆⠈⢦⠀⠀⠀⡇⠀⢸⠀⢠⠎⠀⡼⢡⠃⡏⠷⢻⢁⢿⣏⢧⠈⠙⢦⡀⠀⠀⠙⢆⠀⠀⠈⣧⠀⠀
⠀⠀⣠⣤⣀⠀⠸⡆⠀⢇⠀⢀⡇⠀⢸⣶⠃⠀⡜⢀⠇⢸⠁⢤⠿⣭⡎⡹⠌⡇⠀⠀⢵⡀⠀⠀⠈⡆⠀⠀⢸⣧⠀
⠀⡞⢠⡟⠛⠇⢠⡇⠀⠸⡀⣸⣰⠶⣤⠇⠀⡜⠀⡎⠀⡇⠀⠀⠀⢀⡗⡷⣖⡟⠀⠀⠈⡇⠀⠀⠀⡁⠀⠀⡞⣿⠀
⠀⠹⡄⠳⢄⣠⠞⠁⠀⠀⣧⡟⠃⢀⣼⠀⢰⠁⢰⠀⡸⠀⠀⠀⢀⡼⠁⠀⢸⣇⠀⠀⢐⡇⠀⠀⢀⠇⠀⡜⠀⣿⠀
⠀⠀⠘⠦⣀⡀⠀⢀⣠⢴⠏⠀⠀⢸⠇⣇⢸⠀⠀⠀⡷⢤⣴⡖⢫⡀⠀⠀⠘⢈⡗⠀⡼⠀⠀⢀⠎⢀⠎⠀⢠⡟⠀
⠀⠀⠀⠀⠀⠈⠉⠁⠀⡿⠀⠀⠀⢸⠀⠘⢾⠀⢸⠀⠹⣾⡿⣇⡼⠋⠁⠀⢀⡞⠁⡰⠁⠀⢠⢊⡴⠁⠀⢀⡞⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣧⠀⠀⢀⣼⡆⠀⠀⠑⢤⣧⡀⣀⡴⣹⠃⠀⠀⠀⡞⠀⣰⠃⠀⣰⢣⠊⠀⠀⣠⡿⠁⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠶⠖⠋⠀⢷⡀⠀⢀⡾⢸⠋⠁⠀⡇⠀⠀⠀⠸⡇⠀⡏⠀⢠⢣⠃⠀⢀⣴⠟⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠶⠟⠁⢸⣄⣀⡤⣧⠀⠀⢀⡴⠃⢰⡇⠀⡎⠘⠀⢀⡟⠋⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠓⠚⠁⠀⠀⠀⢧⠀⡇⢀⠀⢸⠁⠀⠀⣤⣄⡀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢧⣣⠘⡄⠈⢧⡀⠀⠀⢻⠈⣦
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠦⣑⡄⠀⠉⠒⠒⠋⣠⠟
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠙⠓⠒⠚⠉⠀⠀
( i am not a brony!! )
*/
