package WaffleTorrent

import (
	"sync"
	"time"
)

type File struct {
	// Relative path of the file
	Path []string

	// File length
	Length int64
}

type Torrent struct {
	// Announce URL's (tiered)
	Announce [][]string

	// Torrent comment
	Comment string

	// Author
	CreatedBy string

	// Creation time
	CreatedAt time.Time

	// Total Length
	Length int64

	// Torrent SHA1
	InfoHash []byte

	// Torrent privacy
	Private bool

	// Piece Length
	PieceLength int64

	// Piece Hashes
	Pieces [][20]byte

	Files []*File
}

type Peer struct {
	ID   string // kinda not used in compress format, peers identifiable by IP:Port
	IP   string
	Port int
}

type Response struct {
	Peers      []Peer
	Interval   int
	TrackerId  string
	Complete   int
	Incomplete int
}

type TorrentScheduler struct {
	Lock sync.RWMutex // mutex (could probably use a rw lock)

	PeerConnections []*PeerConnection
	Torrent         *Torrent            // reference to torrent *for hash verification and file formatting*
	Pieces          [][]byte            // retrieved piece data
	Bitfield        []bool              // which pieces have been retrieved
	Holders         [][]*PeerConnection // how many peers hold what pieces
	InFlight        []int               // which pieces are being requested
	PieceCount      int                 // total number of pieces

	CommandChan chan PeerCommand
	UpdateChan  chan PeerUpdate
}

type PeerCommand struct {
	RequestPieces []int
	CancelPieces  []int
}

type PeerUpdate struct {
	Peer      *Peer
	Piece     int    // piece index
	Status    string // error msg
	BlockData []byte // nullable -> no data sent
}

type PeerConnection struct {
	Peer           *Peer
	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
}

type HandShake struct {
	Pstr     string
	InfoHash [20]byte
	PeerID   [20]byte
}
