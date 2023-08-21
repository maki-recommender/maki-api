package datafetch

import (
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
		return malGetUserID(username)
	}

}

// update tracking site user list and add it to user
func UpdateAnimeList(user *models.TrackingSiteUser) error {
	log.Infof("Started anime list update for %s user %s", user.TrackingSite.Name, user.Username)

	start := time.Now()
	site := user.TrackingSite.Name

	var err error = nil

	if site == "anilist" {
		err = anilistGetUserAnimeList(user)
	} else {
		err = malGetUserAnimeList(user)
	}

	if err != nil {
		log.Errorf("Error while updating user list: %s", err)
	}

	delta := time.Since(start)
	log.Infof("Anime list update for %s user %s took %dms", site, user.Username, delta.Milliseconds())
	return err
}
