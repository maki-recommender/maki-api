package datafetch

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"rickycorte/maki/models"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm/clause"
)

const malEndpoint = "https://myanimelist.net"

var malClientID string

func SetMalClientID(clientID string) {
	malClientID = clientID
}

func malGetUserID(username string) (int, error) {

	url := fmt.Sprintf("%s/search/prefix.json?type=user&keyword=%s&v=1", malEndpoint, username)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return -1, err
	}

	var data jobj
	err = json.NewDecoder(resp.Body).Decode(&data)
	defer resp.Body.Close()
	if err != nil {
		return -1, err
	}

	items := data["categories"].([]interface{})[0].(jobj)["items"].([]interface{})

	for i := 0; i < len(items); i++ {
		itm := items[i].(jobj)
		if strings.ToLower(itm["name"].(string)) == username {
			return int(itm["id"].(float64)), nil
		}
	}

	log.Errorf("Unable to find user: %s", err)
	return -1, errors.New("user does not exists")
}

func malFetchListBlock(username string, offset int) (*[]interface{}, error) {

	url := fmt.Sprintf("%s/animelist/%s/load.json?offset=%d&status=7", malEndpoint, username, offset*300)
	log.Infof("Fetching %d->%d of mal user %s's anime list", offset*300, (offset+1)*300, username)

	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		return nil, err
	}

	var data []interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func malGetUserAnimeList(user *models.TrackingSiteUser) error {

	//TODO: mal api goes faster?

	conflictClause := clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "anime_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"watch_status_id", "score"}),
	}

	offset := 0

	block, err := malFetchListBlock(user.Username, offset)

	if err != nil {
		log.Errorf("Unable to find mal user %s's anime list: %s", user.Username, err)
		return err
	}

	for block != nil {

		if err != nil {
			log.Errorf("Something went wrong parsing mal user %s's anime list: %s", user.Username, err)
			return err
		}

		// insert items into the database
		for i := 0; i < len(*block); i++ {
			e := (*block)[i].(jobj)

			animeID := clause.Expr{
				SQL:  "(SELECT id FROM animes WHERE mal_id = ?)",
				Vars: []interface{}{int(e["anime_id"].(float64))},
			}

			values := map[string]interface{}{
				"user_id":         user.ID,
				"anime_id":        animeID,
				"watch_status_id": int(e["status"].(float64)),
				"score":           float32(e["score"].(float64) / 10),
			}

			dberr := models.SqlDB.Clauses(conflictClause).Table("anime_list_entries").Create(values).Error

			if dberr != nil {
				log.Info(dberr)
			}

		}

		offset += 1

		if len(*block) < 300 {
			break
		} else {
			block, err = malFetchListBlock(user.Username, offset)
		}
	}

	models.SqlDB.Save(&user)

	return nil

}
