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
	fmt.Println("Length:")
	fmt.Printf("\t%d\n", torrent.Length)

	fmt.Println("Privacy:")
	fmt.Printf("\t%d\n", torrent.Private)

	fmt.Println("PieceLength:")
	fmt.Printf("\t%d\n", torrent.PieceLength)

	fmt.Println("PieceHashes:")
	for i, piece := range torrent.Pieces {
		if i >= 40 {
			fmt.Println("\t...")
			break
		}
		fmt.Printf("\t%x\n", piece)
	}
}
