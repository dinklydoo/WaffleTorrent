package WaffleTorrent

import (
	"time"

	"github.com/zeebo/bencode"
)

type FileMetadata struct {
	Path     []string `bencode:"path"`
	PathUtf8 []string `bencode:"path.utf-8"`
	Length   int64    `bencode:"length"`
}

type InfoMetadata struct {
	PieceLength int64  `bencode:"piece length"`
	Pieces      []byte `bencode:"pieces"`

	// single file context
	Name     string `bencode:"name"`
	NameUtf8 string `bencode:"name.utf-8"`
	Length   int64  `bencode:"length"`
	Private  int    `bencode:"private"`

	// multi file context
	Files bencode.RawMessage `bencode:"files"`
}

type Metadata struct {
	// Foobar   []interface{} `bencode:"announce-list"`
	Announce     []string           `bencode:"announce"`
	AnnounceList [][]string         `bencode:"announce-list"`
	Comment      string             `bencode:"comment"`
	CreatedBy    string             `bencode:"created by"`
	CreatedAt    int64              `bencode:"creation date"`
	Info         bencode.RawMessage `bencode:"info"`
}

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

type ResponseMetadata struct {
	Peers      []byte `bencode:"peers"`
	Interval   int    `bencode:"interval"`
	TrackerId  string `bencode:"tracker_id"`
	Complete   int    `bencode:"complete"`
	Incomplete int    `bencode:"incomplete"`
}

type Peer struct {
	ID   string
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
