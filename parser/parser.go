package parser

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
)

const (
	smartSeparator = "@££@"
)

type Parser struct {
	src string
}

func FromResponseBody(reader io.Reader) (*Parser, error) {
	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, err
	}

	return FromString(string(body))
}

func FromString(text string) (*Parser, error) {
	if strings.HasPrefix(text, "-:") {
		return nil, errors.New("sajax raised error:" + text[2:])
	}

	if strings.HasPrefix(text, "+:") {
		text = text[2:]
	}

	if strings.HasPrefix(text, "\"") {
		err := json.Unmarshal([]byte(text), &text)
		if err != nil {
			return nil, err
		}
	}

	if strings.Contains(text, smartSeparator) {
		parts := strings.SplitN(text, smartSeparator, 3)
		text = parts[1]
	}

	return &Parser{
		src: text,
	}, nil
}
