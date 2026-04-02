package Peer

const (
	PeerChoke        uint8 = 0
	PeerUnchoke      uint8 = 1
	PeerInterested   uint8 = 2
	PeerUninterested uint8 = 3
	PeerHave         uint8 = 4
	PeerBitfield     uint8 = 5
	PeerRequest      uint8 = 6
	PeerPiece        uint8 = 7
	PeerCancel       uint8 = 8
)

type PeerMessageType uint8

type PeerComm struct {
	Type     PeerMessageType // identifies the contents of the message via ^^
	Contents []byte          // message contents
}
