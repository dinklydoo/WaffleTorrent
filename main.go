package main

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]

	torrent_path := args[0]
	bytes, err := os.ReadFile(torrent_path)
	if err != nil {
		log.Fatal(err)
	}
	torrent, err := WaffleTorrent.ParseBencodeTorrent(bytes)
	if err != nil {
		log.Fatal(err)
	}
	torrent.Print()
}
