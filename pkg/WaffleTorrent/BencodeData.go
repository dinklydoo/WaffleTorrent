package WaffleTorrent

import (
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
	Announce     string             `bencode:"announce"`
	AnnounceList [][]string         `bencode:"announce-list"`
	Comment      string             `bencode:"comment"`
	CreatedBy    string             `bencode:"created by"`
	CreatedAt    int64              `bencode:"creation date"`
	Info         bencode.RawMessage `bencode:"info"`
}

type ResponseMetadata struct {
	Peers      []byte `bencode:"peers"`
	Interval   int    `bencode:"interval"`
	TrackerId  string `bencode:"tracker_id"`
	Complete   int    `bencode:"complete"`
	Incomplete int    `bencode:"incomplete"`
}
