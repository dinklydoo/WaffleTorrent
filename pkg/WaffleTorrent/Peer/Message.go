package Peer

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"bufio"
	"encoding/binary"
	"errors"
	"io"
	"log"
)

func (p *Peer) UpdatePeer(msg PeerMessage) {
	switch msg.Type() {
	case Choke:
		p.Conn.PeerChoking = true
	case Unchoke:
		p.Conn.PeerChoking = false
	case Interested:
		p.Conn.PeerInterested = true
	case Bitfield:
		t := msg.(*PeerBitfield)
		copy(p.Conn.Bitfield[:], t.Bitfield[:])
	default:
	}
}

func ParseMessage(reader *bufio.Reader, torrent *WaffleTorrent.Torrent) (PeerMessage, error) {
	lengthBuf := make([]byte, 4)
	_, err := io.ReadFull(reader, lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf) - 1 // length of message : subtract id byte which we read automatically
	id, err := reader.ReadByte()                     // the id byte ^^^ (explains -1)
	if err != nil {
		return nil, err
	}
	switch MessageType(id) {
	case Choke:
		return &PeerChoke{
			PeerBase{messageType: Choke},
		}, nil
	case Unchoke:
		return &PeerUnchoke{
			PeerBase{messageType: Unchoke},
		}, nil
	case Interested:
		return &PeerInterested{
			PeerBase{Interested},
		}, nil
	case Uninterested:
		return &PeerUninterested{
			PeerBase{messageType: Uninterested},
		}, nil
	case Bitfield:
		return parseBitfield(reader, length, len(torrent.Pieces))
	case Piece:
		return parsePiece(reader, length)
	default:
		// don't throw just consume the message and continue
		log.Printf("Unrecognized message type %d", MessageType(id))
		return parseUnknown(reader, length)
	}
}

func parseBitfield(reader *bufio.Reader, length uint32, pieceCount int) (*PeerBitfield, error) {
	bytes := uint32((pieceCount + 7) / 8)
	if length != bytes {
		return nil, errors.New("invalid bitfield length")
	}
	msg := new(PeerBitfield)
	msg.messageType = Bitfield

	rawField := make([]byte, length)
	msg.Bitfield = make([]bool, pieceCount)
	_, err := io.ReadFull(reader, rawField)
	if err != nil {
		return nil, err
	}

	hiBit := byte(1 << 7) // byte high bit
	for i := 0; i < pieceCount; i++ {
		mask := hiBit >> (i % 8)
		bi := i / 8
		msg.Bitfield[i] = (rawField[bi] & mask) != 0
	}

	// check unset bits for invalidation
	set := pieceCount % 8
	if set != 0 {
		fb := rawField[length-1]
		for i := set; i < 8; i++ {
			check := fb & (hiBit >> i)
			if check != 0 {
				return nil, errors.New("bitfield piece set beyond piececount")
			}
		}
	}
	return msg, nil
}

func parsePiece(reader *bufio.Reader, length uint32) (*PeerPiece, error) {
	msg := new(PeerPiece)
	msg.messageType = Piece
	buf := make([]byte, 4)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	msg.Index = binary.BigEndian.Uint32(buf)
	_, err = io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	msg.Start = binary.BigEndian.Uint32(buf)
	msg.Block = make([]byte, length-8)
	_, err = io.ReadFull(reader, msg.Block)
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func parseUnknown(reader *bufio.Reader, length uint32) (*PeerUnkown, error) {
	msg := new(PeerUnkown)
	msg.messageType = Unknown
	buf := make([]byte, length)
	_, err := io.ReadFull(reader, buf)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
