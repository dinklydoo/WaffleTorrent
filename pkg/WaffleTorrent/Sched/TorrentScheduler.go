package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"net"
)

const (
	maxPeers   = 250
	maxUpdates = 400
)

func RunTorrentScheduler(torrent *WaffleTorrent.Torrent, peers []Peer.Peer, port int, peerId string, listener *net.Listener) error {
	pieceCount := len(torrent.Pieces)
	sched := &TorrentScheduler{
		Torrent:    torrent,
		Bitfield:   make([]bool, pieceCount),
		Holders:    make([]int, pieceCount),
		InFlight:   make([]int, pieceCount),
		PieceCount: pieceCount,

		UpdateChan:  make(chan *PeerUpdate, maxUpdates),
		RequestChan: make(chan *PeerRequest, maxPeers),
		PeerChan:    make([]chan *PeerCommand, maxPeers),
	}
	RunPeerConnections(peers, sched, port, peerId)

	return nil
}

func RunPeerConnections(peers []Peer.Peer, sched *TorrentScheduler, port int, peerId string) {
	ch := make(chan int, maxPeers)
	for i := 0; i < maxPeers; i++ { // populate channel with all valid slots in the scheduler
		ch <- i
	}

	for _, peer := range peers {
		id := <-ch
		go func(p *Peer.Peer, s *TorrentScheduler, port int, peerId string) {
			defer func() {
				ch <- id
			}()
			err := sched.HandlePeer(&peer, port, peerId, id)
			if err != nil {
				// add some form of logging
			}
			ch <- id
		}(&peer, sched, port, peerId)
	}
}
