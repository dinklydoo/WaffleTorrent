package Sched

// EndGame Scheduling Strategy
func (sched TorrentScheduler) ScheduleEnd() {
	//close(sched.UpdateChan) // not sure if I close this here

	for !sched.Finished() {

		// needs to be asynchronous -> wait for all pieces with timeout

		for pid, _ := range sched.PeerChan {
			sched.requestAll(pid) // request all missing pieces
		}

	}
}

func (sched TorrentScheduler) requestAll(pid PeerId) {
	com := PeerCommand{}
	com.Command = CommandGet
	com.Bitfield = sched.Bitfield

	sched.PeerChan[pid] <- &com
}

func (sched TorrentScheduler) Finished() bool {
	for _, have := range sched.Bitfield {
		if !have {
			return false
		}
	}
	return true
}
