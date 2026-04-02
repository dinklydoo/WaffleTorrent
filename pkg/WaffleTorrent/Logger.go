package WaffleTorrent

import (
	"fmt"
)

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
