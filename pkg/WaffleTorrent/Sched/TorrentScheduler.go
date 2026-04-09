package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"log"
	"net"
	"os"
)

const (
	maxPeers    = 100
	maxUpdates  = 250
	maxCommands = 10
)

func RunTorrentScheduler(torrent *WaffleTorrent.Torrent, peers []Peer.Peer, peerId string, listener *net.Listener) error {
	pieceCount := len(torrent.Pieces)
	InitRQueue(pieceCount)
	sched := &TorrentScheduler{
		Torrent:    torrent,
		PieceFile:  InitPieceFile(int64(torrent.Length)),
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
	log.SetOutput(logFile)

	RunPeerConnections(peers, sched, peerId, logFile)

schedLoop:
	for {
		select {
		case req, ok := <-sched.RequestChan:
			if !ok {
				break schedLoop
			}
			log.Printf("Scheduler received a request: %d", req.PeerSlot)
			sched.handleRequest(req)
		case update, ok := <-sched.UpdateChan:
			if !ok {
				break schedLoop
			}
			log.Printf("Scheduler received an update: slot %d, id: %d", update.PeerSlot, int(update.UpdateType))
			err := sched.updateSchedule(update)
			if err != nil {
				log.Fatal(err)
			}
			if sched.Finished() {
				break schedLoop
			}
		}
	}
	return nil
}

func RunPeerConnections(peers []Peer.Peer, sched *TorrentScheduler, peerId string, logFile *os.File) {
	ch := make(chan int, maxPeers)
	for i := 0; i < maxPeers; i++ { // populate channel with all valid slots in the scheduler
		ch <- i
	}
	for _, peer := range peers {
		id := <-ch
		go func(p *Peer.Peer, s *TorrentScheduler, peerId string) {
			defer func() {
				ch <- id
			}()
			if sched.PeerChan[id] == nil {
				sched.PeerChan[id] = make(chan *PeerCommand, maxCommands)
			}
			err := sched.HandlePeer(&peer, peerId, PeerSlot(id))
			if err != nil {
				log.Printf("Error handling peer %v: %v", peer, err)
			}
			ch <- id
		}(&peer, sched, peerId)
	}
}

func OpenLogFile() *os.File {
	logFile, err := os.OpenFile("./tmp/waffle.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0666)
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
	file, err := os.Create("./tmp/pieces.bin")
	if err != nil {
		log.Fatalf("error creating pieces.bin: %v", err)
	}
	err = file.Truncate(fileSize)
	if err != nil {
		log.Fatalf("error truncating pieces.bin: %v", err)
	}
	return file
}

func (sched *TorrentScheduler) updateSchedule(update *PeerUpdate) error {
	flag := update.UpdateType
	switch flag {
	case PeerBitfield:
		for i, have := range update.Bitfield {
			if have {
				sched.Holders[i]++
				sched.Rarity(i)
			}
		}
	case PeerSuccess:
		if sched.Bitfield[update.Piece] { // someone was faster
			break
		}
		sched.Bitfield[update.Piece] = true
		sched.InFlight[update.Piece]--
		RQueue.Delete(RItem[update.Piece]) // delete from the queue
		err := sched.writePiece(update.Piece, update.BlockData)
		if err != nil {
			return err
		}
	case PeerFailed:
		sched.InFlight[update.Piece]--
		sched.Rarity(update.Piece)
	case PeerDied:
		for i, have := range update.Bitfield {
			if have {
				sched.Holders[i]--
				sched.Rarity(i)
			}
		}
		sched.ActiveChan[update.PeerSlot] = false
	case PeerAttached:
		sched.ActiveChan[update.PeerSlot] = true
	}
	return nil
}

func (sched *TorrentScheduler) writePiece(piece int, data []byte) error {
	offset := uint32(piece) * sched.Torrent.PieceLength

	//end := min(offset+sched.Torrent.PieceLength, sched.Torrent.Length)
	_, err := sched.PieceFile.WriteAt(data[:], int64(offset))
	if err != nil {
		return err
	}
	return nil
}

func (sched *TorrentScheduler) handleRequest(request *PeerRequest) {
	sched.scheduleRare(request) // TODO : endgame heuristic
}

func (sched *TorrentScheduler) Finished() bool {
	for _, b := range sched.Bitfield {
		if !b {
			return false
		}
	}
	return true
}
