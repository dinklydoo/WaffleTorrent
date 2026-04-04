package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"log"
	"net"
	"os"
)

const (
	maxPeers   = 100
	maxUpdates = 250
)

func RunTorrentScheduler(torrent *WaffleTorrent.Torrent, peers []Peer.Peer, port int, peerId string, listener *net.Listener) error {
	pieceCount := len(torrent.Pieces)
	sched := &TorrentScheduler{
		Torrent:    torrent,
		PieceFile:  InitPieceFile(torrent.Length),
		Bitfield:   make([]bool, pieceCount),
		Holders:    make([]int, pieceCount),
		InFlight:   make([]int, pieceCount),
		PieceCount: pieceCount,

		UpdateChan:  make(chan *PeerUpdate, maxUpdates),
		RequestChan: make(chan *PeerRequest, maxPeers),
		PeerChan:    make([]chan *PeerCommand, maxPeers),
		ActiveChan:  make([]bool, maxPeers),
	}
	// setup logfile
	logFile := OpenLogFile()
	defer logFile.Close()

	RunPeerConnections(peers, sched, port, peerId, logFile)

schedLoop:
	for {
		select {
		case req, ok := <-sched.RequestChan:
			if !ok {
				break schedLoop
			}
			sched.handleRequest(req)
		case update, ok := <-sched.UpdateChan:
			if !ok {
				break schedLoop
			}
			err := sched.updateSchedule(update)
			if err != nil {
				log.Fatal(err)
			}
		default:
			if sched.Finished() {
				break schedLoop
			}
		}
	}
	// TODO : validate pieces
	return nil
}

func RunPeerConnections(peers []Peer.Peer, sched *TorrentScheduler, port int, peerId string, logFile *os.File) {
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
			err := sched.HandlePeer(&peer, port, peerId, PeerSlot(id))
			if err != nil {
				log.Println(err) // append error to logfile
			}
			ch <- id
		}(&peer, sched, port, peerId)
	}
}

func OpenLogFile() *os.File {
	logFile, err := os.OpenFile("waffle.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening log file: %v", err)
	}
	return logFile
}

/*
Init Scratch Piece File, Just a file to import raw bits
Does not consider multiple file boundaries in multi-file torrents
*/
func InitPieceFile(fileSize int64) *os.File {
	file, err := os.Create("pieces.bin")
	if err != nil {
		log.Fatalf("error creating pieces.bin: %v", err)
	}
	err = file.Truncate(fileSize)
	if err != nil {
		log.Fatalf("error truncating pieces.bin: %v", err)
	}
	return file
}

func (sched TorrentScheduler) updateSchedule(update *PeerUpdate) error {
	flag := update.UpdateType
	switch flag {
	case PeerBitfield:
		for i, have := range update.Bitfield {
			if have {
				sched.Holders[i]++
			}
		}
	case PeerSuccess:
		sched.Bitfield[update.Piece] = true
		sched.InFlight[update.Piece]--
		err := sched.writePiece(update.Piece, update.BlockData)
		if err != nil {
			return err
		}
	case PeerFailed:
		sched.InFlight[update.Piece]--
	case PeerDied:
		for i, b := range update.Bitfield {
			if b {
				sched.Holders[i]--
			}
		}
		sched.ActiveChan[update.PeerSlot] = false
	case PeerAttached:
		sched.ActiveChan[update.PeerSlot] = true
	}
	return nil
}

func (sched *TorrentScheduler) writePiece(piece int, data []byte) error {
	offset := int64(piece) * sched.Torrent.PieceLength

	end := max(offset+sched.Torrent.PieceLength, sched.Torrent.Length)
	_, err := sched.PieceFile.WriteAt(data[:end], offset)
	if err != nil {
		return err
	}
	return nil
}

func (sched TorrentScheduler) handleRequest(request *PeerRequest) {
	// TODO : implement strategies
}

func (sched TorrentScheduler) Finished() bool {
	for _, b := range sched.Bitfield {
		if !b {
			return false
		}
	}
	return true
}
