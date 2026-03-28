package main

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]

	torrent_path := args[0]

	torrent, err := WaffleTorrent.ParseTorrentFromFile(torrent_path)
	if err != nil {
		log.Fatal(err)
	}
	torrent.Print()
}
