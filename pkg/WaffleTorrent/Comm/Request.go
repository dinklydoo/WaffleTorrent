package Comm

// peer requests work from the scheduler -> scheduler requests for the rarest piece from this peer
type PeerSlot int

type PeerRequest struct { // signal for work
	Bitfield *[]bool  // what pieces it has
	PeerSlot PeerSlot // what channel to use
}

func Request(slot int, bitfield *[]bool) *PeerRequest {
	return &PeerRequest{
		Bitfield: bitfield,
		PeerSlot: PeerSlot(slot),
	}
}
