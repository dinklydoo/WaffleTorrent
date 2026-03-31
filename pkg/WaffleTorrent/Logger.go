package WaffleTorrent

import "fmt"

func (torrent Torrent) Print() {
	fmt.Println("Announce:")
	for i, announceItem := range torrent.Announce {
		fmt.Printf("\tTier %d:\n", i)
		for _, announce := range announceItem {
			fmt.Printf("\t\t%s\n", announce)
		}
	}
	fmt.Printf("Length: %d\n", torrent.Length)
	fmt.Printf("Privacy: %b\n", torrent.Private)
	fmt.Printf("FileCount: %d\n", len(torrent.Files))
	fmt.Printf("PieceLength: %d\n", torrent.PieceLength)

	fmt.Println("PieceHashes:")
	for i, piece := range torrent.Pieces {
		if i >= 10 {
			fmt.Printf("\t... %d more\n", len(torrent.Pieces)-10)
			break
		}
		fmt.Printf("\t%x\n", piece)
	}
}

func (resp Response) Print() {
	fmt.Println("Response: ")
	fmt.Printf("TrackerId: %s\n", resp.TrackerId)
	fmt.Printf("Interval: %d\n", resp.Interval)
	fmt.Printf("Complete(Seeds): %d\n", resp.Complete)
	fmt.Printf("InComplete: %d\n", resp.Incomplete)
	fmt.Println("Peers: ")
	for _, peer := range resp.Peers {
		fmt.Print("\t")
		peer.Print()
	}
}

func (peer Peer) Print() {
	fmt.Printf("Peer: %s\n", peer.ID)
	fmt.Printf("\tPort: %d\n", peer.Port)
	fmt.Printf("\tIP: %s\n", peer.IP)
}
