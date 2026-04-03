package Peer

import "fmt"

type Peer struct {
	ID   string // kinda not used in compress format, peers identifiable by IP:Port
	IP   string
	Port int
	Conn *PeerConnection
}

type PeerConnection struct {
	AmChoking      bool
	AmInterested   bool
	PeerChoking    bool
	PeerInterested bool
	Bitfield       []bool
}

type Response struct {
	Peers      []Peer
	Interval   int
	TrackerId  string
	Complete   int
	Incomplete int
}

// DEBUG METHODS

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

func (p *Peer) Print() {
	fmt.Printf("Peer: %s\n", p.ID)
	fmt.Printf("\tPort: %d\n", p.Port)
	fmt.Printf("\tIP: %s\n", p.IP)
}
