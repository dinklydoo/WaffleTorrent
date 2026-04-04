package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"os"
)

type PeerSlot int
type UpdateType uint8
type CommandType uint8

const (
	PeerSuccess UpdateType = iota
	PeerFailed
	PeerBitfield
	PeerDied
	PeerAttached
)

const (
	CommandGet CommandType = iota
	CommandCancel
	CommandKill
)

type TorrentScheduler struct {
	Torrent    *WaffleTorrent.Torrent // reference to torrent *for hash verification and file formatting*
	PieceFile  *os.File
	Pieces     [][]byte // retrieved piece data move this to a new holder struct
	Bitfield   []bool   // which pieces have been retrieved
	Holders    []int    // how many peers hold what pieces
	InFlight   []int    // which pieces are being requested
	PieceCount int      // total number of pieces

	UpdateChan  chan *PeerUpdate    // update queue -> scheduler reads peer updates from this
	RequestChan chan *PeerRequest   // request queue -> scheduler assigns peers work using this, peers request work explicitly
	PeerChan    []chan *PeerCommand // command queue -> scheduler sends work (command) to requested peers
	ActiveChan  []bool
}

func (sched TorrentScheduler) SendSuccess(idx int, piece []byte, slot PeerSlot) {
	sched.UpdateChan <- &PeerUpdate{
		UpdateType: PeerSuccess,
		PeerSlot:   slot,
		Piece:      idx,
		BlockData:  piece,
	}
	// TODO : do this in scheduler
	//sched.Pieces[idx] = piece
	//sched.Bitfield[idx] = true
	//sched.InFlight[idx]--
	//sched.PieceCount++
}

// SendRequest : Send a request to the scheduler for work
func (sched TorrentScheduler) SendRequest(bitField []bool, slot PeerSlot) {
	sched.RequestChan <- &PeerRequest{
		Bitfield: bitField,
		PeerSlot: slot,
	}
}

func (sched TorrentScheduler) SendFailure(piece int, slot PeerSlot) {
	sched.UpdateChan <- &PeerUpdate{
		UpdateType: PeerFailed,
		Piece:      piece,
		PeerSlot:   slot,
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

type PeerUpdate struct {
	UpdateType UpdateType
	PeerSlot   PeerSlot
	Piece      int    // piece index
	Bitfield   []bool // bitfield -> empty on non bitfield updates
	BlockData  []byte // non-empty on success message
}
type PeerCommand struct {
	Command CommandType
	Piece   int // signal which piece we wish to retrieve from peer
}
