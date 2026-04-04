package Sched

//TODO : ENDGAME STRATEGY TRACK BLOCKS INSTEAD OF PIECES TO INCREASE THROUGHPUT
//TODO : MULTIPLE PEERS CAN CONTRIBUTE TO THE SAME PIECES

// EndGame Scheduling Strategy
func (sched TorrentScheduler) ScheduleEnd() {
	//close(sched.UpdateChan) // not sure if I close this here

	for !sched.Finished() {

		// needs to be asynchronous -> wait for all pieces with timeout

		for pid, _ := range sched.PeerChan {
			sched.requestAll(PeerSlot(pid)) // request all missing pieces
		}

	}
}

func (sched TorrentScheduler) requestAll(pid PeerSlot) {
	com := PeerCommand{}
	com.Command = CommandGet
	com.Piece = -1

	sched.PeerChan[pid] <- &com
}
