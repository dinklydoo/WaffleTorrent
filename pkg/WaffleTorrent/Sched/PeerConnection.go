package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"bufio"
	"bytes"
	"fmt"
	"net"
	"slices"
	"strings"
	"time"
)

func (sched TorrentScheduler) HandlePeer(p *Peer.Peer, port int, peerId string, slot int) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.IP, p.Port))
	if err != nil {
		return err
	}
	defer conn.Close()
	pc := Peer.PeerConnection{true, false, true, false, make([]byte, sched.PieceCount)}
	p.Conn = &pc

	err = torrentHandshake(&conn, peerId, sched.Torrent.InfoHash)
	if err != nil {
		return err
	}

	updateChan := &sched.UpdateChan
	reqChan := &sched.RequestChan
	workChan := &sched.PeerChan[slot]

	// now we can receive messages after handshake passes
	reader := bufio.NewReader(conn)

	// TODO: send attach message to sched

	// parse first message
	msg, err := Peer.ParseMessage(reader, sched.Torrent)
	if err != nil {
		return err
	}
	if msg.Type() == Peer.Bitfield {

	} else { // peer is a seeder -> has all pieces

	}

	for true {
		_, err = Peer.ParseMessage(reader, sched.Torrent)
		if err != nil {
			return err // immediately drop this peer -> data is not trustworthy
		}
		break
	}

	// TODO: send kill message to scheduler
	return nil
}

// ESTABLISH PEER CONNECTION

func createHandshake(peerId string, infoHash []byte) string {
	handshake := strings.Builder{}
	handshake.WriteString("\x13BitTorrent protocol") // protocol header

	extensions := "\x00\x00\x00\x00\x00\x00\x00\x00" // extension bytes (reserved bytes)
	handshake.WriteString(extensions)

	handshake.WriteString(string(infoHash)) // infohash
	handshake.WriteString(peerId)           // peerId

	return handshake.String()
}

/*
Initiate the BitTorrent Handshake with a peer, reads
handshake response from peer and verifies

if successful -> returns with no error
else -> returns an error
*/
func torrentHandshake(conn *net.Conn, peerId string, infoHash []byte) error {
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
