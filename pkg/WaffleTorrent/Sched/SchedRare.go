package Sched

import (
	"container/heap"
	"log"
)

/*
RarityQueue global static (mimic of static) instance of our rarity pq

Not a conventional PQ, we need to support indexed deletion and element update / reinsertion
also need to allow iteration over elements
*/

/*
⠀⠀⠀⠀⠀⢀⣀⡤⠤⠶⠄⠠⡶⢤⠤⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠀⣠⠞⠉⠓⠢⠄⣀⠀⠀⠱⡀⡇⠀⠉⠳⣤⠶⢦⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⢀⡞⠓⠒⠒⠂⠀⠤⢀⡙⠢⡀⢱⣠⣀⠀⣴⠃⠀⡈⢧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⡎⠀⣠⠒⠒⠒⢤⡀⠀⢉⡶⠛⠉⠀⠀⠀⠁⠀⠀⡄⢸⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⢸⠃⢰⠁⠀⠀⠀⡞⠉⠲⠋⠀⠀⠀⠀⠀⠀⠀⠀⠸⠁⡼⢧⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⢸⡀⠸⡆⠀⠀⢰⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⡼⠁⢸⡆⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠈⢧⠀⢧⠀⠀⣸⣧⣀⠀⠀⢀⣠⡤⢤⣤⣤⡄⠀⠀⡄⠀⠘⡇⣠⣤⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀
⠀⠘⣦⠈⢣⠀⢸⠟⡋⠀⠀⠉⠁⠀⠀⠀⠉⠀⠀⢠⠃⠀⣶⣿⡏⢸⡗⣦⣀⡠⠤⠤⠤⠤⠤⣀⠀⠀⠀⠀⠀⠀⠀
⠀⠀⠈⢧⡀⠱⡌⢷⣤⡆⠀⠀⠀⠀⠀⢀⡀⠀⢀⡎⠀⣸⣿⣿⠃⣸⠀⡟⠁⠀⠀⠀⠀⠀⠀⠀⠉⠲⣄⠀⠀⠀⠀
⠀⠀⠀⠀⠳⣄⠘⢆⠈⠉⠓⠲⡖⠚⢻⠉⠀⢀⡞⠀⡰⡹⢸⡟⢠⡏⢠⣏⣉⣀⠀⠉⠑⠢⢄⡀⠀⠀⠘⢷⡀⠀⠀
⠀⠀⠀⠀⠀⠙⣆⠈⢦⠀⠀⠀⡇⠀⢸⠀⢠⠎⠀⡼⢡⠃⡏⠷⢻⢁⢿⣏⢧⠈⠙⢦⡀⠀⠀⠙⢆⠀⠀⠈⣧⠀⠀
⠀⠀⣠⣤⣀⠀⠸⡆⠀⢇⠀⢀⡇⠀⢸⣶⠃⠀⡜⢀⠇⢸⠁⢤⠿⣭⡎⡹⠌⡇⠀⠀⢵⡀⠀⠀⠈⡆⠀⠀⢸⣧⠀
⠀⡞⢠⡟⠛⠇⢠⡇⠀⠸⡀⣸⣰⠶⣤⠇⠀⡜⠀⡎⠀⡇⠀⠀⠀⢀⡗⡷⣖⡟⠀⠀⠈⡇⠀⠀⠀⡁⠀⠀⡞⣿⠀
⠀⠹⡄⠳⢄⣠⠞⠁⠀⠀⣧⡟⠃⢀⣼⠀⢰⠁⢰⠀⡸⠀⠀⠀⢀⡼⠁⠀⢸⣇⠀⠀⢐⡇⠀⠀⢀⠇⠀⡜⠀⣿⠀
⠀⠀⠘⠦⣀⡀⠀⢀⣠⢴⠏⠀⠀⢸⠇⣇⢸⠀⠀⠀⡷⢤⣴⡖⢫⡀⠀⠀⠘⢈⡗⠀⡼⠀⠀⢀⠎⢀⠎⠀⢠⡟⠀
⠀⠀⠀⠀⠀⠈⠉⠁⠀⡿⠀⠀⠀⢸⠀⠘⢾⠀⢸⠀⠹⣾⡿⣇⡼⠋⠁⠀⢀⡞⠁⡰⠁⠀⢠⢊⡴⠁⠀⢀⡞⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⣧⠀⠀⢀⣼⡆⠀⠀⠑⢤⣧⡀⣀⡴⣹⠃⠀⠀⠀⡞⠀⣰⠃⠀⣰⢣⠊⠀⠀⣠⡿⠁⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠶⠖⠋⠀⢷⡀⠀⢀⡾⢸⠋⠁⠀⡇⠀⠀⠀⠸⡇⠀⡏⠀⢠⢣⠃⠀⢀⣴⠟⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠶⠟⠁⢸⣄⣀⡤⣧⠀⠀⢀⡴⠃⢰⡇⠀⡎⠘⠀⢀⡟⠋⠀⠀⠀⠀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠓⠚⠁⠀⠀⠀⢧⠀⡇⢀⠀⢸⠁⠀⠀⣤⣄⡀⠀
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⢧⣣⠘⡄⠈⢧⡀⠀⠀⢻⠈⣦
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⠦⣑⡄⠀⠉⠒⠒⠋⣠⠟
⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠙⠓⠒⠚⠉⠀⠀
( i am not a brony!! )*/

// RQueue : Global rarity queue, don't really want to attach its lifespan to the scheduler object so lets make it global
var RQueue *RarityQueue
var RItem []*PieceItem // maps index to piece items

func InitRQueue(PieceCount int) {
	RQueue = new(RarityQueue)
	RItem = make([]*PieceItem, PieceCount)
	for i := 0; i < PieceCount; i++ {
		RItem[i] = new(PieceItem)
		RItem[i].Index = i
		RItem[i].Rarity = 0

		RQueue.Push(RItem[i])
	}
}

// Rarity : Calculate the rarity based on the
func (sched *TorrentScheduler) Rarity(index int) {
	temp := float64(sched.Holders[index]) * (float64(sched.InFlight[index]) + 1)
	if sched.Holders[index] != 0 {
		RItem[index].Rarity = 1 / temp
	}
	RQueue.Update(RItem[index])
}

type PieceItem struct {
	Index        int
	Availability int
	InFlight     int
	Rarity       float64

	heapIndex int
}
type RarityQueue []*PieceItem

//TODO : lowkey understand how the fuck do pq's really work in golang :-)

func (pq RarityQueue) Len() int { return len(pq) }

func (pq RarityQueue) Less(i, j int) bool {
	return pq[i].Rarity > pq[j].Rarity // max heap -> higher rarity is at the top
}

func (pq RarityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].heapIndex = i
	pq[j].heapIndex = j
}

func (pq *RarityQueue) Push(x any) {
	item := x.(*PieceItem)
	item.heapIndex = len(*pq)
	*pq = append(*pq, item)
}

func (pq *RarityQueue) Pop() any {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.heapIndex = -1
	*pq = old[:n-1]
	return item
}

// THESE TWO OPERATIONS ARE VERY EXPENSIVE USE SPARINGLY
func (pq *RarityQueue) Delete(item *PieceItem) {
	idx := item.heapIndex
	heap.Remove(pq, idx)
}

func (pq *RarityQueue) Update(item *PieceItem) {
	heap.Fix(pq, item.heapIndex)
}

func (sched *TorrentScheduler) scheduleRare(request *PeerRequest) {
	// iterates over PQ, finds rarest piece that the scheduler has
	for _, item := range *RQueue {
		if sched.Bitfield[item.Index] {
			continue
		}
		if request.Bitfield[item.Index] {
			log.Printf("Peer %d assigned piece %d", request.PeerSlot, item.Index)
			sched.PeerChan[request.PeerSlot] <- &PeerCommand{
				Command: CommandGet,
				Piece:   item.Index,
			}
			sched.InFlight[item.Index]++
			sched.Rarity(item.Index)
			break
		}
	}
}
