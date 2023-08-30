package anime

import (
	"rickycorte/maki/models"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

var genre2bitmap = map[string]int{
	"action":        1,
	"adventure":     2,
	"comedy":        3,
	"drama":         4,
	"ecchi":         5,
	"fantasy":       6,
	"hentai":        7,
	"horror":        8,
	"mahou_shoujo":  9,
	"mecha":         10,
	"music":         11,
	"mistery":       12,
	"psychological": 13,
	"romance":       14,
	"sci-fi":        15,
	"slice_of_life": 16,
	"sports":        17,
	"supernatural":  18,
	"thriller":      19,
}

func IsValidGenre(genre string) bool {
	_, ok := genre2bitmap[genre]

	return ok
}

/* ----------------------------------------------------------------------------*/

// cache entry that can be loaded from postgres
type AnimeCacheEntry struct {
	ID        uint
	hasMal    bool
	genreMask uint64
}

func (a *AnimeCacheEntry) HasMalID() bool {
	return a.hasMal
}

func (a *AnimeCacheEntry) IsHentai() bool {
	return a.HasGenre("hentai")
}

func (a *AnimeCacheEntry) HasGenre(genre string) bool {
	if !IsValidGenre(genre) {
		return false
	}

	return a.genreMask&(1<<uint64(genre2bitmap[genre])) != 0
}

/* ----------------------------------------------------------------------------*/

var animeCache = sync.Map{}

func refreshAnimeCache() {
	rows, err := models.GetAnimeCacheRows()
	if err != nil {
		log.Errorf("Error while retriving anime cache entries: %s", err)
	}
	defer rows.Close()

	for rows.Next() {
		t := AnimeCacheEntry{}
		var genre string
		rows.Scan(&t.ID, &t.hasMal, &genre)

		genre = genre[1 : len(genre)-1] // remove { }
		genres := strings.Split(genre, ",")

		for i := 0; i < len(genres); i++ {
			if IsValidGenre(genres[i]) {
				t.genreMask |= 1 << genre2bitmap[genres[i]]
			}
		}

		animeCache.Store(t.ID, &t)
	}
}

func perdiocallyRefreshAnimeCache() {

	for {
		log.Info("Refreshing anime cache")
		refreshAnimeCache()
		log.Info("Anime cache refreshed")
		time.Sleep(600 * time.Second) // TODO: make settings parameter
	}

}

// return a pointer to a cache entry. May return nill if not found
func GetAnimeLocalCacheEntry(id uint) *AnimeCacheEntry {
	val, ok := animeCache.Load(id)
	if ok {
		return val.(*AnimeCacheEntry)
	}

	return nil
}
