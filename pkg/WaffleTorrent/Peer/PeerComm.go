package Peer

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"bufio"
	"encoding/binary"
	"errors"
)

func ParseMessage(reader *bufio.Reader, torrent *WaffleTorrent.Torrent) (PeerMessage, error) {
	lengthBuf := make([]byte, 4)
	_, err := reader.Read(lengthBuf)
	if err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lengthBuf) - 1 // length of message - length prefix
	id, err := reader.ReadByte()
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
	msg.Bitfield = make([]byte, length)
	_, err := reader.Read(msg.Bitfield)
	if err != nil {
		return nil, err
	}
	set := pieceCount % 8
	if set != 0 {
		fb := msg.Bitfield[length-1]
		for i := set; i < 8; i++ {
			if fb&(128>>i) != 0 {
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
	_, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	msg.Index = binary.BigEndian.Uint32(buf)
	_, err = reader.Read(buf)
	if err != nil {
		return nil, err
	}
	msg.Start = binary.BigEndian.Uint32(buf)
	msg.Block = make([]byte, length-8)
	n, err := reader.Read(msg.Block)
	if err != nil {
		return nil, err
	}
	if n != len(msg.Block) {
		return nil, errors.New("message length mismatch")
	}
	return msg, nil
}

func parseUnknown(reader *bufio.Reader, length uint32) (*PeerUnkown, error) {
	msg := new(PeerUnkown)
	msg.messageType = Unknown
	buf := make([]byte, length)
	_, err := reader.Read(buf)
	if err != nil {
		return nil, err
	}
	return msg, nil
}
