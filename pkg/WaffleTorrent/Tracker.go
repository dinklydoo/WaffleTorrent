package WaffleTorrent

import (
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	url2 "net/url"
	"strconv"
	"strings"
)

func getPeerList(torrent *Torrent, tier int, port int) ([]Peer, error) {
	if tier >= len(torrent.Announce) {
		return nil, errors.New("No Peer's found")
	}
	trackers := torrent.Announce[tier]

	ch := make(chan *Response)
	for _, t := range trackers {
		tracker := t
		go func() {
			resp, _ := announceToTracker(torrent, tracker, port)
			ch <- resp
		}()
	}
	var peerMap map[string]Peer
	for i := 0; i < len(trackers); i++ {
		resp := <-ch
		if resp == nil {
			for _, peer := range resp.Peers {
				peerMap[peer.ID] = peer
			}
		}
	}
	var peerList []Peer
	for _, peer := range peerMap {
		peerList = append(peerList, peer)
	}

	if len(peerList) == 0 {
		return getPeerList(torrent, tier+1, port)
	}

	return peerList, nil
}

func announceToTracker(torrent *Torrent, tracker string, port int) (*Response, error) {
	url, err := constructURL(torrent, tracker, port)
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

func constructURL(torrent *Torrent, tracker string, port int) (string, error) {
	url, err := url2.Parse(tracker)
	if err != nil {
		return "", err
	}

	peer_id := generatePeerId()

	params := url.Query()
	params.Set("info_hash", urlByteEncode([]byte(torrent.InfoHash)))
	params.Set("peer_id", peer_id)
	params.Set("port", strconv.Itoa(port))
	params.Set("uploaded", "0")
	params.Set("downloaded", "0")
	params.Set("compact", "1") // always request compacted torrents -> helps us when parsing
	params.Set("left", strconv.FormatInt(torrent.Length, 10))

	url.RawQuery = params.Encode()
	return url.String(), nil
}

func urlByteEncode(bytes []byte) string {
	var result strings.Builder

	for _, b := range bytes {
		if (b >= 'a' && b <= 'z') ||
			(b >= 'A' && b <= 'Z') ||
			(b >= '0' && b <= '9') ||
			b == '-' || b == '.' || b == '_' || b == '~' {
			result.WriteByte(b)
		} else {
			result.WriteString(fmt.Sprintf("%%%02X", b))
		}
	}

	return result.String()
}

func generatePeerId() string {
	var result strings.Builder
	result.WriteString("-WAFFLE-") // not following format but idc :)
	for i := 0; i < 12; i++ {
		result.WriteString(strconv.Itoa(rand.Int() % 10))
	}
	return result.String()
}
