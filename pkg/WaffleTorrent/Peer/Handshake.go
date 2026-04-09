package Peer

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"slices"
	"strings"
	"time"
)

/*
Initiate the BitTorrent Handshake with a peer, reads
handshake response from peer and verifies

if successful -> returns with no error
else -> returns an error
*/
func TorrentHandshake(conn *net.Conn, peerId string, infoHash []byte) error {
	hs := createHandshake(peerId, infoHash)

	timeout := time.Now().Add(5 * time.Second)
	err := (*conn).SetDeadline(timeout)
	if err != nil {
		return err
	}
	_, err = (*conn).Write([]byte(hs))
	if err != nil {
		return err
	}
	response := make([]byte, 68)
	read, err := io.ReadFull((*conn), response)
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
	rstream := bytes.NewBuffer(response)
	pstrlen, err := rstream.ReadByte()
	if err != nil {
		return err
	}
	pstr := make([]byte, pstrlen)
	_, err = io.ReadFull(rstream, pstr)
	if err != nil {
		return err
	}
	if string(pstr) != "BitTorrent protocol" {
		return fmt.Errorf("Handshake response is not BitTorrent, response protocol of %s", pstr)
	}

	rbuf := make([]byte, 8)
	_, err = io.ReadFull(rstream, rbuf) // read past reserved bytes -> don't care about them for now
	if err != nil {
		return err
	}
	recHash := make([]byte, 20)
	_, err = io.ReadFull(rstream, recHash)
	if err != nil {
		return err
	}
	if !slices.Equal(recHash, infoHash) {
		return fmt.Errorf("InfoHash mismatch, response hash of %s", infoHash)
	}
	peerId := make([]byte, 20)
	_, err = io.ReadFull(rstream, peerId)
	if err != nil {
		return err
	}
	return nil // looks good to me, peace out
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
