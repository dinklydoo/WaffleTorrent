package Comm

type UpdateType uint8

const (
	PeerSuccess UpdateType = iota
	PeerFailed
	PeerBitfield
	PeerDied
	PeerAttached
)

type PeerUpdate struct {
	UpdateType UpdateType
	PeerSlot   PeerSlot
	Piece      int    // piece index
	Bitfield   []bool // bitfield -> empty on non bitfield updates
	BlockData  []byte // non-empty on success message
}

func UpdateSuccess(slot int, piece int, bdata []byte) *PeerUpdate {
	return &PeerUpdate{
		UpdateType: PeerSuccess,
		PeerSlot:   PeerSlot(slot),
		Piece:      piece,
		Bitfield:   nil,
		BlockData:  bdata,
	}
}

func UpdateFailed(slot int, piece int) *PeerUpdate {
	return &PeerUpdate{
		UpdateType: PeerFailed,
		PeerSlot:   PeerSlot(slot),
		Piece:      piece,
		Bitfield:   nil,
		BlockData:  nil,
	}
}

func UpdateBitfield(slot int, bitfield []bool) *PeerUpdate {
	return &PeerUpdate{
		UpdateType: PeerBitfield,
		PeerSlot:   PeerSlot(slot),
		Piece:      -1,
		Bitfield:   bitfield,
		BlockData:  nil,
	}
}

func UpdateDetached(slot int, bitfield []bool) *PeerUpdate {
	return &PeerUpdate{
		UpdateType: PeerDied,
		PeerSlot:   PeerSlot(slot),
		Piece:      -1,
		Bitfield:   bitfield,
		BlockData:  nil,
	}
}

func UpdateAttached(slot int) *PeerUpdate {
	return &PeerUpdate{
		UpdateType: PeerAttached,
		PeerSlot:   PeerSlot(slot),
		Piece:      -1,
		Bitfield:   nil,
		BlockData:  nil,
	}
}
