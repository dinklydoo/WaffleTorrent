package WaffleTorrent

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
)

/*
	Parser/Deserializer for a .torrent file with Bencoded data,
	should return a deserialized Torrent Object for request to Torrent Tracker

	Example BeenCode for Debian (formatted nicely *note this would all be in a single line*):

	d
		8:announce
		41:http://bttracker.debian.org:6969/announce
		7:comment
		35:"Debian CD from cdimage.debian.org"
		13:creation date
		i1573903810e
		4:info
		d
			6:length
			i351272960e
			4:name
			31:debian-10.2.0-amd64-netinst.iso
			12:piece length
			i262144e
			6:pieces
			26800:... random bytes ... (binary blob of the hashes of each piece)
		e
	e
*/

type BencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     BencodeInfo `bencode:"info"`
}

type BencodeInfo struct {
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
}

func ParseBencodeTorrent(data []byte) (*BencodeTorrent, error) {
	reader := bytes.NewReader(data)

	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 'd' {
		return nil, fmt.Errorf("invalid torrent: expected dict ('d'), got %q", b)
	}

	torrent, err := parseBencodeHeader(reader)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	info, err := parseBencodeInfo(reader)
	if err != nil {
		return nil, fmt.Errorf("parse info: %w", err)
	}
	torrent.Info = *info

	b, err = reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != 'e' {
		return nil, fmt.Errorf("invalid torrent: expected end dict ('e'), got %q", b)
	}

	return torrent, nil
}

func parseBencodeHeader(reader *bytes.Reader) (*BencodeTorrent, error) {
	var bt BencodeTorrent

	// read announce link
	err := parseAttribute(reader, "announce")
	if err != nil {
		return nil, err
	}
	announce_buf, err := readAttribute(reader)
	if err != nil {
		return nil, err
	}
	bt.Announce = string(*announce_buf)

	return &bt, nil
}

func parseBencodeInfo(reader *bytes.Reader) (*BencodeInfo, error) {
	var info BencodeInfo

	err := parseAttribute(reader, "length")
	if err != nil {
		return nil, err
	}
	length, err := parseInteger(reader)
	if err != nil {
		return nil, err
	}

	err = parseAttribute(reader, "name")
	if err != nil {
		return nil, err
	}
	name_buf, err := readAttribute(reader)
	if err != nil {
		return nil, err
	}

	// read past piece length
	err = parseAttribute(reader, "piece length")
	if err != nil {
		return nil, err
	}
	piece_length, err := parseInteger(reader)
	if err != nil {
		return nil, err
	}

	err = parseAttribute(reader, "pieces")
	if err != nil {
		return nil, err
	}
	pieces_buf, err := readAttribute(reader)
	if err != nil {
		return nil, err
	}

	info.Length = length
	info.Name = string(*name_buf)
	info.PieceLength = piece_length
	info.Pieces = string(*pieces_buf)
	return &info, nil
}

func parseAttribute(reader *bytes.Reader, name string) error {
	attrName := "__empty__"
	for attrName != name {
		b, err := reader.ReadByte()
		if err != nil {
			return err
		}
		err = reader.UnreadByte()
		if err != nil {
			return err
		}

		switch b {
		case 'i':
			{
				_, err = parseInteger(reader)
				if err != nil {
					return fmt.Errorf("invalid torrent: could not find attribute %s", name)
				}
			}
		case 'l':
			{
				err = parseList(reader)
				if err != nil {
					return fmt.Errorf("invalid torrent: could not find attribute %s", name)
				}
			}
		case 'd':
			{
				// todo can just skip for now
				_, err = reader.ReadByte()
				if err != nil {
					return fmt.Errorf("invalid torrent: could not find attribute %s", name)
				}
			}
		default:
			{
				buf, err := readAttribute(reader)
				if err != nil {
					return fmt.Errorf("invalid torrent: could not find attribute %s", name)
				}
				attrName = string(*buf)
			}
		}
	}
	return nil
}

func readAttribute(reader *bytes.Reader) (*[]byte, error) {
	length, err := readNumber(reader)
	if err != nil {
		return nil, err
	}
	b, err := reader.ReadByte()
	if err != nil {
		return nil, err
	}
	if b != ':' {
		return nil, fmt.Errorf("expected ':', got %q", b)
	}

	buf := make([]byte, length)
	_, err = reader.Read(buf)
	if err != nil {
		return nil, err
	}
	return &buf, nil
}

// read a number string from reader -> uint32
func readNumber(reader *bytes.Reader) (int, error) {
	res := 0
	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	for b >= '0' && b <= '9' { // b is numerical
		res = 10*res + int(b-'0')
		b, err = reader.ReadByte()
		if err != nil {
			return 0, err
		}
	}
	err = reader.UnreadByte()
	if err != nil {
		return 0, err
	}
	return res, nil
}

// parse bencode integer iXXXXXXe
func parseInteger(reader *bytes.Reader) (int, error) {
	b, err := reader.ReadByte()
	if err != nil {
		return 0, err
	}
	if b != 'i' {
		return 0, errors.New("invalid integer format, expected i ")
	}
	res, err := readNumber(reader)
	if err != nil {
		return 0, err
	}
	b, err = reader.ReadByte()
	if err != nil {
		return 0, err
	}
	if b != 'e' {
		return 0, errors.New("invalid integer format, expected closing e ")
	}
	return res, nil
}

func parseList(reader *bytes.Reader) error {
	return nil
}

// AUX FUNCTIONS

func (bt BencodeTorrent) Print() {
	fmt.Println("Announce:" + bt.Announce)
	fmt.Println("Name:" + bt.Info.Name)
	fmt.Println("Length:" + strconv.Itoa(bt.Info.Length))
	fmt.Println("PieceLength:" + strconv.Itoa(bt.Info.PieceLength))
	fmt.Println("Pieces:" + bt.Info.Pieces[:100] + "...")
}
