package Sched

import (
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"encoding/binary"
	"net"
	"time"
)

func RequestBlock(conn net.Conn, piece int, start uint32, length uint32) error {
	return sendBlock(conn, Peer.Request, piece, start, length)
}

func CancelBlock(conn net.Conn, piece int, start uint32, length uint32) error {
	return sendBlock(conn, Peer.Cancel, piece, start, length)
}

func sendBlock(conn net.Conn, request Peer.MessageType, index int, begin uint32, length uint32) error {
	req := make([]byte, 17) // requests are fixed in size, message = length + length prefix (4B big endian)

	// request: <len=0013><id=X><index><begin><length>
	// idx, begin and length as 4 byte BigEndian uints
	binary.BigEndian.PutUint32(req[0:4], uint32(13))
	req[4] = byte(request)
	binary.BigEndian.PutUint32(req[5:9], uint32(index))
	binary.BigEndian.PutUint32(req[9:13], begin)
	binary.BigEndian.PutUint32(req[13:], length)

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err := conn.Write(req)
	if err != nil {
		return err
	}
	return nil
}
