package Tracker

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"WaffleTorrent/pkg/WaffleTorrent/Peer"
	"errors"
	"io"
	"math/rand"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
)

func GetPeerList(torrent *WaffleTorrent.Torrent, tier int, port int, peerId string) ([]Peer.Peer, error) {
	if tier >= len(torrent.Announce) {
		return nil, errors.New("No Peer's found")
	}
	trackers := torrent.Announce[tier]

	// Buffer go-routines to handle variant response times
	ch := make(chan *Peer.Response, len(trackers))
	for _, t := range trackers {
		tracker := t
		go func() {
			resp, _ := announceToTracker(torrent, tracker, port, peerId)
			resp.Print()
			ch <- resp
		}()
	}
	peerMap := make(map[Peer.Peer]struct{})
	for i := 0; i < len(trackers); i++ {
		resp := <-ch // blocks until a tracker finishes request
		if resp != nil {
			for _, peer := range resp.Peers {
				peerMap[peer] = struct{}{}
			}
		}
	}
	var peerList []Peer.Peer
	for peer, _ := range peerMap {
		peerList = append(peerList, peer)
	}

	if len(peerList) == 0 {
		return GetPeerList(torrent, tier+1, port, peerId)
	}

	return peerList, nil
}

func announceToTracker(torrent *WaffleTorrent.Torrent, tracker string, port int, peerId string) (*Peer.Response, error) {
	url, err := constructURL(torrent, tracker, port, peerId)
	if err != nil {
		return nil, err
	}
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return ParseResponse(body)
}

func constructURL(torrent *WaffleTorrent.Torrent, tracker string, port int, peerId string) (string, error) {
	url, err := url2.Parse(tracker)
	if err != nil {
		return "", err
	}

	params := url.Query()
	params.Set("info_hash", string(torrent.InfoHash))
	params.Set("peer_id", peerId)
	params.Set("port", strconv.Itoa(port))
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("compact", "1") // always request compacted torrents -> helps us when parsing
	params.Set("left", strconv.FormatInt(torrent.Length, 10))

	url.RawQuery = params.Encode()
	return url.String(), nil
}

func GeneratePeerId() string {
	var result strings.Builder
	result.WriteString("-WAFFLE-") // not following format but idc :)
	for i := 0; i < 12; i++ {
		result.WriteString(strconv.Itoa(rand.Int() % 10))
	}
	return result.String()
}
