// Copyright: Daniel Heidemann
// License: GNU AGPL, Version 3 or later; http://www.gnu.org/licenses/agpl.html

package anki

import (
	"archive/zip"
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"maps"
	"strings"
)

// CompressImages resizes all images inside a deckfile
// to a specified maximum size in KB and returns the
// compressed deck
func (a *Apkg) CompressImages(maxSizeKB int) ([]byte, error) {
	maxImageSize := maxSizeKB * 1024 // KB in bytes
	var buf bytes.Buffer
	writer := zip.NewWriter(&buf)

	deckfiles := a.media.index
	maps.Copy(deckfiles, a.meta.index)

	for filename, zipFile := range deckfiles {
		filedata, err := a.ReadMediaFile(filename)
		if err != nil {
			return nil, err
		}

		prefix := strings.Split(filename, "-")[0]
		if prefix == "paste" && len(filedata) > maxImageSize {
			filedata, err = compressImage(filedata, maxImageSize)
			if err != nil {
				return nil, errors.New("unsupported file type")
			}
		}

		newFile, err := writer.Create(zipFile.FileHeader.Name)
		if err != nil {
			return nil, err
		}
		_, err = newFile.Write(filedata)
		if err != nil {
			return nil, err
		}
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func compressImage(data []byte, maxImageSize int) ([]byte, error) {
	img, _, err := image.Decode(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}

	quality := 100
	step := 10

	var buf bytes.Buffer
	for {
		buf.Reset()
		err := jpeg.Encode(&buf, img, &jpeg.Options{Quality: quality})
		if err != nil {
			return data, err
		}

		quality -= step

		if buf.Len() <= maxImageSize || quality < 1 {
			return buf.Bytes(), nil
		}
	}
}
