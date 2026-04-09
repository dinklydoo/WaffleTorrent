package File

import (
	"WaffleTorrent/pkg/WaffleTorrent"
	"fmt"
	"io"
	"os"
)

func FormatPieces(torrent *WaffleTorrent.Torrent) error {
	pieceFile, err := os.OpenFile("./tmp/pieces.bin", os.O_RDWR, 0777)
	defer pieceFile.Close()

	files := torrent.Files
	outPath := "./out/"
	err = os.RemoveAll(outPath)
	if err != nil {
		return fmt.Errorf("could not remove old files: %v", err)
	}
	err = os.MkdirAll(outPath, 0755)
	if err != nil {
		return fmt.Errorf("could not create directory: %v", err)
	}
	offset := int64(0)
	for _, file := range files {
		for i, dir := range file.Path {

			if i == len(file.Path)-1 {
				f, err := os.OpenFile(outPath+dir, os.O_WRONLY|os.O_CREATE, 0755)
				if err != nil {
					return fmt.Errorf("could not create file: %v", err)
				}
				_, err = io.CopyN(f, pieceFile, file.Length)
				if err != nil {
					return fmt.Errorf("could not write to file: %v", err)
				}
				f.Close()
				offset += file.Length
				_, err = pieceFile.Seek(offset, io.SeekStart)
				if err != nil {
					return fmt.Errorf("could not seek: %v", err)
				}
				break
			}
			// not the final file, just a directory
			err = os.MkdirAll(dir, 0755)
			if err != nil {
				return fmt.Errorf("could not create directory: %v", err)
			}
		}
	}
	return nil
}
