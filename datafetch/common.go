package datafetch

import (
	"errors"
	"rickycorte/maki/models"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type jobj = map[string]interface{}

// check if a user exists on a tracking site
func GetUserId(site, username string) (int, error) {
	if site == "anilist" {
		return anilistGetUserID(username)
	} else {
		log.Info("Ops mal is not available yet")
		return -1, errors.New("not implemented yet")
	}

}

// update tracking site user list and add it to user
func UpdateAnimeList(user *models.TrackingSiteUser) error {
	log.Infof("Started anime list update for %s user %s", user.TrackingSite.Name, user.Username)

	start := time.Now()
	site := user.TrackingSite.Name

	if site == "anilist" {
		return anilistGetUserAnimeList(user)
	} else {
		log.Info("Ops mal is not available yet")
	}

	delta := time.Since(start)
	log.Infof("Anime list update for %s user %s took %dms", site, user.Username, delta.Milliseconds())
	return nil
}
