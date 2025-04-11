package parser

import (
	"log"
	"net/url"
	"strconv"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type RegistryEntry struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type Registry struct {
	Teachers map[string]RegistryEntry
	Students map[string]RegistryEntry
	Rooms    map[string]RegistryEntry
}

func (r *Parser) Registry() (*Registry, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(r.src))
	if err != nil {
		return nil, err
	}

	directory := &Registry{
		Teachers: make(map[string]RegistryEntry),
		Students: make(map[string]RegistryEntry),
		Rooms:    make(map[string]RegistryEntry),
	}

	doc.Find(".ulroot > li.submenu.liroot").Each(func(i int, section *goquery.Selection) {
		section.Find("li.link > a").Each(func(i int, link *goquery.Selection) {
			href, exists := link.Attr("href")
			if !exists {
				return
			}

			name := strings.TrimSpace(link.Text())
			if name == "" {
				return
			}

			u, err := url.Parse(href)
			if err != nil {
				return
			}

			q := u.Query()
			entryType := q.Get("type")

			// Skip type=9 (Classes/Teachings)
			if entryType == "9" {
				return
			}

			rawID := q.Get("id")
			if rawID == "" {
				return
			}

			numID, err := strconv.ParseUint(rawID, 10, 32)
			if err != nil {
				log.Printf("Failed to parse non-empty ID '%s': %v", rawID, err)
				return
			}

			entry := RegistryEntry{
				ID:   uint(numID),
				Name: name,
			}

			switch entryType {
			case "1":
				directory.Teachers[name] = entry
			case "2":
				directory.Students[name] = entry
			case "4":
				directory.Rooms[name] = entry
			}
		})
	})
	return directory, nil
}

func getTypeString(typeID uint) string {
	switch typeID {
	case 1:
		return "teacher"
	case 2:
		return "student"
	case 4:
		return "room"
	default:
		return "unknown"
	}
}
