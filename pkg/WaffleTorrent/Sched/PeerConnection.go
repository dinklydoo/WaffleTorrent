package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"bufio"
	"fmt"
	"net"
	"time"
)

func (sched TorrentScheduler) HandlePeer(p *Peer.Peer, port int, peerId string, slot PeerSlot) error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", p.IP, p.Port))
	if err != nil {
		return err
	}
	defer conn.Close()

	pc := Peer.PeerConnection{
		AmChoking:      true,
		AmInterested:   false,
		PeerChoking:    true,
		PeerInterested: false,
		Bitfield:       make([]bool, sched.PieceCount),
	}
	p.Conn = &pc

	err = Peer.TorrentHandshake(&conn, peerId, sched.Torrent.InfoHash)
	if err != nil {
		return err
	}
	err = conn.SetReadDeadline(time.Now().Add(2 * time.Minute))
	if err != nil {
		return err
	}
	reader := bufio.NewReader(conn)

	readch := make(chan Peer.PeerMessage) // read channel : buffer reads for io mux
	comch := sched.PeerChan[slot]

	sched.attachPeer(slot) // --- ATTACH PEER

	// parse first message
	msg, err := Peer.ParseMessage(reader, sched.Torrent)
	if err != nil {
		return err
	}
	sched.peerFirstMsg(msg, slot, readch)

	var cons PieceConstructor
	cons.Init(sched.Torrent.PieceLength) // initialize constructor size

	// socket reader loop
	go func(reader *bufio.Reader, torrent *WaffleTorrent.Torrent) {
		for {
			err := conn.SetReadDeadline(time.Now().Add(2 * time.Minute)) // refresh read deadline
			if err != nil {
				break
			}

			msg, err := Peer.ParseMessage(reader, torrent)
			if err != nil {
				break
			}
			readch <- msg
		}
	}(reader, sched.Torrent)

loop:
	for {
		select { // io mux with read and command channel

		case cmd, ok := <-comch:
			if !ok {
				break loop
			}

		case msg, ok := <-readch:
			if !ok {
				break loop
			}
			p.UpdatePeer(msg) // updates peer metadata

			if cons.CanRequest() && !p.Conn.PeerChoking { // can send a request
				sched.SendRequest(p.Conn.Bitfield, slot)
				cons.Waiting = true
			}

			if msg.Type() == Peer.Piece {
				cons.AddBlock(msg)
				if cons.Full() { // piece has been retrieved
					piece, err := cons.Verify(sched.Torrent.Pieces[cons.PieceIndex])
					if err != nil {
						// TODO : send piece drop message
						break
					}
					sched.AddPiece(cons.PieceIndex, piece)
					cons.Clear()
				}
			}
		}
	}
	sched.detachPeer(p, slot) // --- DETACH PEER

	return nil
}

/*
We can directly send an update to the scheduler as no commands will come to a new connection
with no request made

TODO : maybe don't make this a method of the scheduler
*/
func (sched TorrentScheduler) peerFirstMsg(msg Peer.PeerMessage, slot PeerSlot, readch chan Peer.PeerMessage) {
	switch msg.Type() {
	case Peer.Bitfield:
		t := msg.(*Peer.PeerBitfield)
		sched.UpdateChan <- &PeerUpdate{
			UpdateType: PeerBitfield,
			PeerSlot:   slot,
			Bitfield:   t.Bitfield,
		}
	default: // peer is a seeder
		seed := make([]bool, sched.PieceCount)
		for i := range seed {
			seed[i] = true
		}
		sched.UpdateChan <- &PeerUpdate{
			UpdateType: PeerBitfield,
			PeerSlot:   slot,
			Bitfield:   seed,
		}
		readch <- msg // enqueue the message again
	}
}
