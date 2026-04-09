package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Comm"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"os"
)

type TorrentScheduler struct {
	Torrent    *WaffleTorrent.Torrent // reference to torrent *for hash verification and file formatting*
	PieceFile  *os.File
	Bitfield   []bool // which pieces have been retrieved
	PieceCount int    // total number of pieces

	WriteBuf    int
	UpdateChan  chan *Comm.PeerUpdate    // update queue -> scheduler reads peer updates from this
	RequestChan chan *Comm.PeerRequest   // request queue -> scheduler assigns peers work using this, peers request work explicitly
	PeerChan    []chan *Comm.PeerCommand // command queue -> scheduler sends work (command) to requested peers
	ActiveChan  []bool
}

func (sched *TorrentScheduler) SendSuccess(idx int, piece []byte, slot int) {
	sched.UpdateChan <- Comm.UpdateSuccess(slot, idx, piece)
}

// SendRequest : Send a request to the scheduler for work
func (sched *TorrentScheduler) SendRequest(bitField []bool, slot int) {
	sched.RequestChan <- Comm.Request(slot, bitField)
}

func (sched *TorrentScheduler) SendFailure(piece int, slot int) {
	sched.UpdateChan <- Comm.UpdateFailed(slot, piece)
}

func (sched *TorrentScheduler) attachPeer(slot int) {
	sched.UpdateChan <- Comm.UpdateAttached(slot)
}

func (sched *TorrentScheduler) detachPeer(p *Peer.Peer, slot int) {
	sched.UpdateChan <- Comm.UpdateDetached(slot, p.Conn.Bitfield)
}
