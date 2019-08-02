package tflite

import (
	"strconv"
)

type PPMInfo struct {
	Width int
	Height int
	Offset int
}

func isSpace(b byte) bool {
	switch b {
	case ' ':
		return true
	case '\t':
		return true
	case '\n':
		return true
	case '\r':
		return true
	}
	return false
}

func FindPPMData(data []byte) *PPMInfo {

	i := new(PPMInfo)

	// Get the next header token
	getToken := func() string {
		// Get Token
		token := make([]byte, 0)
		for i.Offset < len(data) && !isSpace(data[i.Offset]) {
			token = append(token, data[i.Offset])
			i.Offset++
		}
		// Eat Spaces
		for i.Offset < len(data) && isSpace(data[i.Offset]) {
			i.Offset++
		}
		return string(token)
	}

	// First Token
	if getToken() != "P6" {
		return nil
	}

	var err error

	i.Width, err = strconv.Atoi(getToken())
	if err != nil {
		return nil
	}

	i.Height, err = strconv.Atoi(getToken())
	if err != nil {
		return nil
	}

	if maxVal, err := strconv.Atoi(getToken()); err != nil || maxVal != 255 {
		return nil
	}

	return i
}
