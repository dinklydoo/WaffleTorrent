package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"encoding/binary"
	"net"
)

func RequestBlock(conn *net.Conn, piece int, start int, length int) error {
	return sendBlock(conn, Peer.Request, piece, start, length)
}

func CancelBlock(conn *net.Conn, piece int, start int, length int) error {
	return sendBlock(conn, Peer.Cancel, piece, start, length)
}

func sendBlock(conn *net.Conn, request Peer.MessageType, index int, begin int, length int) error {
	req := make([]byte, 13) // requests are fixed in size
	req[0] = byte(request)

	// request: <len=0013><id=X><index><begin><length>
	// idx, begin and length as 4 byte BigEndian uints
	binary.BigEndian.PutUint32(req[1:5], uint32(index))
	binary.BigEndian.PutUint32(req[5:9], uint32(begin))
	binary.BigEndian.PutUint32(req[9:], uint32(length))

	_, err := (*conn).Write(req)
	if err != nil {
		return err
	}
	return nil
}
