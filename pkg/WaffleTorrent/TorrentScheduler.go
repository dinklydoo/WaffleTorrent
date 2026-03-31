package WaffleTorrent

func (sched *TorrentScheduler) newConnection(conn *PeerConnection) {
	sched.PeerConnections = append(sched.PeerConnections, conn)

}
