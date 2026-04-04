package Sched

// Rarity Scheduling Strategy

func (sched TorrentScheduler) ScheduleRare() {
run_rare:
	for {
		select {
		case update := <-sched.UpdateChan:
			// run logic
			err := sched.updateSchedule(update)
			if err != nil {
				return
			}
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
