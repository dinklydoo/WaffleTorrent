package main

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"fmt"
	"log"
	"net"
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

	var listener net.Listener
	for port := 6881; port <= 6889; port++ {
		listener, err = net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err != nil {
			continue
		}
		break
	}
	if listener == nil {
		log.Fatal("Waffle: Could not find an open listener port:6881-6889")
	}
	defer listener.Close()

	peerId := WaffleTorrent.GeneratePeerId()
	_, err = WaffleTorrent.GetPeerList(torrent, 0, 6881, peerId) // 6881-6889
	if err != nil {
		log.Fatal(err)
	}

}
