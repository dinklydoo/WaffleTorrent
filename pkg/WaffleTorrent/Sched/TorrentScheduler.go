package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Comm"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"WaffleTorrent/pkg/WaffleTorrent/SchedRare"
	"log"
	"net"
	"os"
)

const (
	maxPeers    = 20
	maxUpdates  = 80
	maxCommands = 10
)

var rarityQueue *SchedRare.RarityQueue

func RunTorrentScheduler(torrent *WaffleTorrent.Torrent, peers []Peer.Peer, peerId string, listener *net.Listener) error {
	pieceCount := len(torrent.Pieces)
	rarityQueue = SchedRare.NewRarityQueue(len(torrent.Pieces))
	sched := &TorrentScheduler{
		Torrent:    torrent,
		PieceFile:  InitPieceFile(int64(torrent.Length)),
		Bitfield:   make([]bool, pieceCount),
		PieceCount: pieceCount,
		WriteBuf:   0,

		UpdateChan:  make(chan *Comm.PeerUpdate, maxUpdates),
		RequestChan: make(chan *Comm.PeerRequest, maxPeers),
		PeerChan:    make([]chan *Comm.PeerCommand, maxPeers),
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
				sched.PeerChan[id] = make(chan *Comm.PeerCommand, maxCommands)
			}
			log.Printf("socket: %v assigned slot %d", p.IP, id)
			err := sched.HandlePeer(&peer, peerId, id)
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

func (sched *TorrentScheduler) updateSchedule(update *Comm.PeerUpdate) error {
	flag := update.UpdateType
	switch flag {
	case Comm.PeerBitfield:
		for i := 0; i < len(sched.Bitfield); i++ {
			if sched.Bitfield[i] {
				rarityQueue.IncHolder(i)
			}
		}
	case Comm.PeerSuccess:
		sched.Bitfield[update.Piece] = true
		rarityQueue.RequestSuccess(update.Piece)

		err := sched.writePiece(update.Piece, update.BlockData)
		if err != nil {
			return err
		}
	case Comm.PeerFailed:
		rarityQueue.RequestFailed(update.Piece)
	case Comm.PeerDied:
		for i := 0; i < len(sched.Bitfield); i++ {
			if sched.Bitfield[i] {
				rarityQueue.DecHolder(i)
			}
		}
		sched.ActiveChan[update.PeerSlot] = false
	case Comm.PeerAttached:
		sched.ActiveChan[update.PeerSlot] = true
	}
	return nil
}

func (sched *TorrentScheduler) writePiece(piece int, data []byte) error {
	offset := uint32(piece) * sched.Torrent.PieceLength

	_, err := sched.PieceFile.WriteAt(data[:], int64(offset))
	if err != nil {
		return err
	}

	sched.WriteBuf++
	if sched.WriteBuf%10 == 0 || sched.WriteBuf == sched.PieceCount {
		err = sched.PieceFile.Sync()
		if err != nil {
			return err
		}
	}
	return nil
}

func (sched *TorrentScheduler) handleRequest(request *Comm.PeerRequest) {
	pidx := rarityQueue.RequestRare(request)
	if pidx < 0 {
		sched.PeerChan[request.PeerSlot] <- Comm.KillCommand()
	} else {
		sched.PeerChan[request.PeerSlot] <- Comm.GetCommand(pidx)
	}
}

func (sched *TorrentScheduler) Finished() bool {
	for _, b := range sched.Bitfield {
		if !b {
			return false
		}
	}
	return true
}
