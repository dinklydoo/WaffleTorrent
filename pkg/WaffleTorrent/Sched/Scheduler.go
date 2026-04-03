package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
)

type PeerSlot int
type UpdateType uint8
type CommandType uint8

const (
	PeerSuccess  UpdateType = 0
	PeerFailed   UpdateType = 1
	PeerBitfield UpdateType = 2
	PeerDied     UpdateType = 3
	PeerAttached UpdateType = 4
)

const (
	CommandGet    CommandType = 0
	CommandCancel CommandType = 1
)

type TorrentScheduler struct {
	Torrent    *WaffleTorrent.Torrent // reference to torrent *for hash verification and file formatting*
	Pieces     [][]byte               // retrieved piece data move this to a new holder struct
	Bitfield   []bool                 // which pieces have been retrieved
	Holders    []int                  // how many peers hold what pieces
	InFlight   []int                  // which pieces are being requested
	PieceCount int                    // total number of pieces

	UpdateChan  chan *PeerUpdate // work queue -> goroutines pull work from this
	RequestChan chan *PeerRequest
	PeerChan    []chan *PeerCommand // update queue -> scheduler pulls updates from this
	ActiveChan  []bool
}

func (sched TorrentScheduler) AddPiece(idx int, piece []byte) {
	sched.Pieces[idx] = piece
	sched.Bitfield[idx] = true
	sched.InFlight[idx]--
	sched.PieceCount++
}

// SendRequest : Send a request to the scheduler for work
func (sched TorrentScheduler) SendRequest(bitField []bool, slot PeerSlot) {
	sched.RequestChan <- &PeerRequest{
		Bitfield: bitField,
		PeerSlot: slot,
	}
}

func (sched TorrentScheduler) attachPeer(slot PeerSlot) {
	s := PeerUpdate{
		UpdateType: PeerAttached,
		PeerSlot:   slot,
	}
	sched.UpdateChan <- &s // thread safe attach signal
}

func (sched TorrentScheduler) detachPeer(p *Peer.Peer, slot PeerSlot) {
	s := PeerUpdate{
		UpdateType: PeerDied,
		PeerSlot:   slot,
		Bitfield:   p.Conn.Bitfield, // supply bitfield for decrement
	}
	sched.UpdateChan <- &s // thread safe kill signal
}

// peer requests work from the scheduler -> scheduler requests for the rarest piece from this peer

type PeerRequest struct { // signal for work
	Bitfield []bool   // what pieces it has
	PeerSlot PeerSlot // what channel to use
}

type PeerCommand struct {
	Command  CommandType
	Bitfield []bool // signal which pieces we wish to retrieve from peer
}

type PeerUpdate struct {
	UpdateType UpdateType
	PeerSlot   PeerSlot
	Piece      int    // piece index
	Bitfield   []bool // bitfield -> empty on non bitfield updates
	BlockData  []byte // non-empty on success message
}
