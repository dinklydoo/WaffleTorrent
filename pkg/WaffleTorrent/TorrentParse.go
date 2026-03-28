package WaffleTorrent

import (
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"os"
	"time"

	"github.com/zeebo/bencode"
)

/*
	Parser/Deserializer for a .torrent file with Bencoded data,
	should return a deserialized Torrent Object for request to Torrent Tracker

	Example Bencode for Debian (formatted nicely *note this would all be in a single line*):

	d
		8:announce
		41:http://bttracker.debian.org:6969/announce
		7:comment
		35:"Debian CD from cdimage.debian.org"
		13:creation date
		i1573903810e
		4:info
		d
			6:length
			i351272960e
			4:name
			31:debian-10.2.0-amd64-netinst.iso
			12:piece length
			i262144e
			6:pieces
			26800:... random bytes ... (binary blob of the hashes of each piece)
		e
	e
*/

/*
	Torrent Parser Credit To
	https://github.com/j-muller/go-torrent-parser/blob/master/utils.go
	Bencode Parser Credit To
	"https://github.com/zeebo/bencode"
*/

func ParseTorrent(data *[]byte) (*Torrent, error) {

	metadata := &Metadata{}
	err := bencode.DecodeBytes(*data, metadata)
	if err != nil {
		return nil, err
	}

	info := &InfoMetadata{}
	err = bencode.DecodeBytes(metadata.Info, info)
	if err != nil {
		return nil, err
	}

	if len(info.NameUtf8) != 0 {
		info.Name = info.NameUtf8
	}

	files := make([]*File, 0)
	total_length := int64(0)

	// single file context
	if info.Length > 0 {
		files = append(files, &File{
			Path:   []string{info.Name},
			Length: info.Length,
		})
		total_length = info.Length
	} else {
		metadataFiles := make([]*FileMetadata, 0)
		err = bencode.DecodeBytes(info.Files, &metadataFiles)
		if err != nil {
			return nil, err
		}

		for _, f := range metadataFiles {
			if len(f.PathUtf8) != 0 {
				f.Path = f.PathUtf8
			}
			files = append(files, &File{
				Path:   append([]string{info.Name}, f.Path...),
				Length: f.Length,
			})
			total_length += f.Length
		}
	}

	announces := make([][]string, 0)

	if len(metadata.AnnounceList) > 0 {
		for _, announceItem := range metadata.AnnounceList {
			announces = append(announces, announceItem)
		}
	} else {
		announces = append(announces, make([]string, 0))
		announces[0] = append(announces[0], metadata.Announce)
	}

	numPieces := len(info.Pieces) / 20
	pieces := make([][20]byte, numPieces)
	for i := 0; i < numPieces; i++ {
		copy(pieces[i][:], info.Pieces[20*i:20*(i+1)])
	}

	return &Torrent{
		Announce:    announces,
		Comment:     metadata.Comment,
		CreatedBy:   metadata.CreatedBy,
		CreatedAt:   time.Unix(metadata.CreatedAt, 0),
		Length:      total_length,
		InfoHash:    toSHA1(metadata.Info),
		Private:     info.Private == 1,
		Pieces:      pieces,
		PieceLength: info.PieceLength,
		Files:       files,
	}, nil
}

func ParseTorrentFromFile(path string) (*Torrent, error) {
	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return ParseTorrent(&file)
}

func toSHA1(data []byte) []byte {
	h := sha1.New()
	h.Write(data)
	return h.Sum(nil)
}

/*

Example bittorrent response from a tracker (compressed) :
	d
		8:complete
			i3651e
		10:incomplete
			i385e
		8:interval
			i1800e
		5:peers
			300:£¬%ËÌyOk‚Ý—. ƒê@_<K+Ô\Ý Ámb^TnÈÕ^ŒAË OŒ*ÈÕ>¥³ÈÕBä)ðþ¸ÐÞ¦Ô/ãÈÕÈuÉæÈÕ
	e
*/

func ParseResponse(data []byte) (*Response, error) {
	meta := &ResponseMetadata{}
	err := bencode.DecodeBytes(data, meta)
	if err != nil {
		return nil, err
	}

	var peers []Peer
	// compact mode: 6 bytes per peer
	for i := 0; i < len(meta.Peers); i += 6 {
		compact := meta.Peers[6*i : 6*(i+1)]

		ip := fmt.Sprintf("%d.%d.%d.%d", compact[0], compact[1], compact[2], compact[3])
		port := binary.BigEndian.Uint16(compact[4:6])

		peers = append(peers, Peer{
			IP:   ip,
			Port: int(port),
		})
	}

	return &Response{
		Peers:      peers,
		Interval:   int(meta.Interval),
		TrackerId:  meta.TrackerId,
		Complete:   meta.Complete,
		Incomplete: meta.Incomplete,
	}, nil
}
