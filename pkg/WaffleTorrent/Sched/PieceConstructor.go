package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"bytes"
	"crypto/sha1"
	"fmt"
	"net"
)

const maxBuffered = 10 // max requests to a peer at a time

type PieceConstructor struct {
	PieceIndex int
	Blocks     [][WaffleTorrent.BlockSize]byte
	PieceSize  int64
	Count      int64
	Waiting    bool // flag if waiting for work
	Inflight   int  // pipeline requests -> requesting 0...N blocks
}

func (p *PieceConstructor) Init(pieceSize int64) {
	blockCount := (2*pieceSize - WaffleTorrent.BlockSize) / WaffleTorrent.BlockSize
	p.Blocks = make([][WaffleTorrent.BlockSize]byte, blockCount)
	p.PieceSize = pieceSize
	p.Count = 0
	p.PieceIndex = -1
	p.Waiting = false
	p.Inflight = 0
}

func (p *PieceConstructor) Full() bool {
	return p.Count == int64(len(p.Blocks))
}

func (p *PieceConstructor) Enqueue(conn *net.Conn) error {
	if p.Inflight == len(p.Blocks) { // pipeline is full
		return nil
	}
	p.Inflight++
	begin := WaffleTorrent.BlockSize * p.Inflight
	end := max(begin+WaffleTorrent.BlockSize, int(p.PieceSize))

	return sendBlock(conn, Peer.Request, p.PieceIndex, begin, end-begin)
}

func (p *PieceConstructor) Cancel(conn *net.Conn) error {
	for i := p.Count; i <= int64(p.Inflight); i++ {
		begin := WaffleTorrent.BlockSize * i
		end := max(begin+WaffleTorrent.BlockSize, p.PieceSize)
		err := sendBlock(conn, Peer.Cancel, p.PieceIndex, int(begin), int(end-begin))
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *PieceConstructor) piece() []byte {
	piece := make([]byte, p.PieceSize)
	for i, b := range p.Blocks {
		start := int64(i * WaffleTorrent.BlockSize)
		end := min(start+WaffleTorrent.BlockSize, p.PieceSize)
		copy(piece[start:end], b[:])
	}
	return piece
}

func (p *PieceConstructor) Clear() {
	p.PieceIndex = -1
	p.Count = 0
	p.Waiting = false
	p.Inflight = 0
}

func (p *PieceConstructor) CanRequest() bool {
	return !p.Waiting && p.PieceIndex == -1
}

func (p *PieceConstructor) Request(idx int) {
	p.PieceIndex = idx
	p.Waiting = false
	p.Inflight = maxBuffered
}

func (p *PieceConstructor) AddBlock(msg Peer.PeerMessage) {
	pm := msg.(*Peer.PeerPiece)
	copy(p.Blocks[pm.Start/WaffleTorrent.BlockSize][:], pm.Block)
	p.Count++
	p.Inflight--
}

func (p *PieceConstructor) Verify(hash [20]byte) ([]byte, error) {
	flat := p.piece()
	sha := sha1.Sum(flat)

	if bytes.Compare(sha[:], hash[:]) != 0 {
		return nil, fmt.Errorf("PieceConstructor Verify: hash mismatch")
	}
	return flat, nil
}
