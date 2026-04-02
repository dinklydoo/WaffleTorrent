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
		PeerChan:    make(map[PeerId]chan *PeerCommand, maxPeers),
	}
	RunPeerConnections(peers, sched, port, peerId)

	return nil
}

func RunPeerConnections(peers []Peer.Peer, sched *TorrentScheduler, port int, peerId string) {
	ch := make(chan struct{}, maxPeers)
	for _, peer := range peers {
		ch <- struct{}{} // acquire slot
		go func(p *Peer.Peer, s *TorrentScheduler, port int, peerId string) {
			defer func() { <-ch }() // release slot
			err := p.HandlePeer(s.Torrent.InfoHash, port, peerId)
			// handle peer should attach the peer to the scheduler directly
			if err != nil {
				// add a disconnect peer method -> reduces fields in scheduler and pulls a new peer in

				<-ch // on error -> free the peer from the slot
			}
		}(&peer, sched, port, peerId)
	}
}
