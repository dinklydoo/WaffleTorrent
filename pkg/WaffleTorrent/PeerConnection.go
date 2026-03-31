package WaffleTorrent

import (
	"bytes"
	"fmt"
	"log"
	"net"
	"slices"
	"strings"
	"sync"
	"time"
)

const maxPeers = 250 // max 250 goroutines running concurrently

func RunPeerConnections(torrent *Torrent, peers []Peer, port int, peerId string, listener *net.Listener) {
	pieceCount := len(torrent.Pieces)
	sched := &TorrentScheduler{
		sync.RWMutex{},
		make([]*PeerConnection, 0),
		torrent,
		make([][]byte, pieceCount),
		make([]bool, pieceCount),
		make([][]*PeerConnection, pieceCount),
		make([]int, pieceCount),
		pieceCount,
		make(chan PeerCommand),
		make(chan PeerUpdate),
	}

	ch := make(chan struct{}, maxPeers)
	for _, peer := range peers {
		ch <- struct{}{} // acquire slot
		go func(p *Peer, s *TorrentScheduler, port int, peerId string) {
			defer func() { <-ch }() // release slot
			err := handlePeer(p, s, port, peerId)
			if err != nil {
				log.Fatal(err) // idk if we need fatal or just a log
			}
		}(&peer, sched, port, peerId)
	}
}

func handlePeer(p *Peer, s *TorrentScheduler, port int, peerId string) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.IP, p.Port))
	if err != nil {
		return err
	}
	defer conn.Close()
	pc := &PeerConnection{p, true, false, true, false}

	err = torrentHandshake(&conn, peerId, s.Torrent.InfoHash)
	if err != nil {
		return err
	}
	// now we can receive messages after handshake passes

	return nil
}

func createHandshake(peerId string, infoHash []byte) string {
	handshake := strings.Builder{}
	handshake.WriteString("\x13BitTorrent protocol") // protocol header

	extensions := "\x00\x00\x00\x00\x00\x00\x00\x00" // extension bytes (reserved bytes)
	handshake.WriteString(extensions)

	handshake.WriteString(string(infoHash)) // infohash
	handshake.WriteString(peerId)           // peerId

	return handshake.String()
}

func torrentHandshake(conn *net.Conn, peerId string, infoHash []byte) error {
	hs := createHandshake(peerId, s.Torrent.InfoHash)

	timeout := time.Now().Add(5 * time.Second)
	err := (*conn).SetDeadline(timeout)
	if err != nil {
		return err
	}
	_, err = (*conn).Write([]byte(hs))
	if err != nil {
		return err
	}
	response := make([]byte, 16*1024)
	read, err := (*conn).Read(response)
	if err != nil {
		return err
	}
	err = verifyHandshake(response[:read], infoHash)
	if err != nil {
		return err
	}
	return nil
}

func verifyHandshake(response []byte, infoHash []byte) error {
	hsLen := 79 // CHANGE THIS TO SOME GLOBAL VAR SO IT'S NOT HARDCODED OR WHATNOT
	if len(response) != hsLen {
		return fmt.Errorf("Handshake response length doesn't match, expect %d, got %d", hsLen, len(response))
	}
	rstream := bytes.NewBuffer(response)
	pstrlen, err := rstream.ReadByte()
	if err != nil {
		return err
	}
	pstr, err := rstream.ReadString(pstrlen)
	if err != nil {
		return err
	}
	if pstr != "BitTorrent protocol" {
		return fmt.Errorf("Handshake response is not BitTorrent, response protocol of %s", pstr)
	}

	rbuf := make([]byte, 8)
	_, err = rstream.Read(rbuf) // read past reserved bytes
	if err != nil {
		return err
	}
	recHash := make([]byte, 20)
	_, err = rstream.Read(recHash)
	if err != nil {
		return err
	}
	if !slices.Equal(recHash, infoHash) {
		return fmt.Errorf("InfoHash mismatch, response hash of %s", infoHash)
	}
	peerId := make([]byte, 20)
	_, err = rstream.Read(peerId)
	if err != nil {
		return err
	}
	return nil // looks good to me, peace out
}
