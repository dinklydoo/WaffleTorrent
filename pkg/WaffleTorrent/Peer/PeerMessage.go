package Peer

type MessageType byte

const (
	Choke MessageType = iota
	Unchoke
	Interested
	Uninterested
	Have
	Bitfield
	Request // not used when we leech just for iota
	Piece
	Cancel
	Unknown
)

type PeerMessage interface {
	Type() MessageType
}

type PeerBase struct {
	messageType MessageType
}

func (p *PeerBase) Type() MessageType {
	return p.messageType
}

type PeerChoke struct {
	PeerBase
}

type PeerUnchoke struct {
	PeerBase
}

type PeerInterested struct {
	PeerBase
}

type PeerUninterested struct {
	PeerBase
}

// TODO : "have" introduces some issues in bittorrents also not required
//type PeerHave struct {
//	PeerBase
//}

type PeerBitfield struct {
	PeerBase
	Bitfield []bool
}

type PeerRequest struct {
	PeerBase
	Index uint32
	Start uint32
}

type PeerCancel struct {
	PeerBase
	Index uint32
	Start uint32 // future ref for block-wise request/cancel -- endgame strategy
}
type PeerKill struct {
	PeerBase
}

type PeerPiece struct {
	PeerBase
	Index uint32
	Start uint32
	Block []byte
}

type PeerUnkown struct {
	PeerBase
	bytes []byte
}
