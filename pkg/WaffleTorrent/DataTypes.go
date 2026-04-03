package WaffleTorrent

import (
	"time"
)

// BlockSize : torrent block size 16 KB
const BlockSize = 16 * 1024

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
