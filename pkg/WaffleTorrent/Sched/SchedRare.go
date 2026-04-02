package Sched

// Rarity Scheduling Strategy

func (sched TorrentScheduler) ScheduleRare() {
run_rare:
	for {
		select {
		case update := <-sched.UpdateChan:
			// run logic
			sched.updateSchedule(update)
		default:
			if sched.switchEndgame() {
				break run_rare
			}
			break
		}
	}
}

func (sched TorrentScheduler) switchEndgame() bool {
	pc := sched.PieceCount
	received := 0
	for _, have := range sched.Bitfield {
		if have {
			received++
		}
	}
	// received more than 80 percent of pieces -> endgame now
	return float64(received) >= 0.8*float64(pc)
}

func (sched TorrentScheduler) updateSchedule(update *PeerUpdate) {
	flag := update.UpdateType
	switch flag {
	case PeerBitfield:
		for i, have := range update.Bitfield {
			if have {
				sched.Holders[i]++
			}
		}
	case PeerSuccess:
		sched.Bitfield[update.Piece] = true
		sched.InFlight[update.Piece]--
		sched.Holders[update.Piece] = -1
	case PeerFailed:
		sched.InFlight[update.Piece]--
		sched.Holders[update.Piece]-- // this peer is not reliable -> boost priority so other holders can attempt
	}
}
