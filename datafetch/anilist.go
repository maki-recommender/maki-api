package datafetch

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"rickycorte/maki/models"
	"strings"

	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm/clause"
)

const anilistEndpoint = "https://graphql.anilist.co"
const anilistQLCheckUserExist = `
query ($username: String) {
	User(name: $username) {
	  id
	  name
	}
  }
`
const anilistQLGetUserList = `
query ($userId: Int, $userName: String, $type: MediaType) {
	MediaListCollection(userId: $userId, userName: $userName, type: $type) {
	  lists {
		name
		entries {
		  ...mediaListEntry
		}
	  }
	  user {
		name
		mediaListOptions {
		  scoreFormat
		}
	  }
	}
  }
  
  fragment mediaListEntry on MediaList {
	mediaId
	status
	score
  }
`

func anilistPost(query string, variables jobj) (*http.Response, error) {
	reqBody, _ := json.Marshal(map[string]interface{}{
		"query":     query,
		"variables": variables,
	})
	return http.Post(anilistEndpoint, "application/json", bytes.NewBuffer(reqBody))
}

// get the user if from anilist if it exsist, if not exists -1 is returned alongside an error
func anilistGetUserID(username string) (int, error) {

	resp, err := anilistPost(anilistQLCheckUserExist, jobj{"username": username})
	if err != nil || resp.StatusCode != 200 {
		return -1, err
	}

	var data jobj
	err = json.NewDecoder(resp.Body).Decode(&data)
	defer resp.Body.Close()
	if err != nil {
		return -1, err
	}

	// check if the user exists
	if _, ok := data["errors"]; ok {
		return -1, errors.New("unable to find user")
	}

	return int(data["data"].(jobj)["User"].(jobj)["id"].(float64)), nil

}

func scoreFormat2Scale(format string) float32 {
	switch format {
	case "POINT_100":
		return 100
	case "POINT_10":
		return 10
	case "POINT_10_DECIMAL":
		return 10
	case "POINT_5":
		return 5
	case "POINT_3":
		return 3
	}

	return 1
}

func anilistGetUserAnimeList(user *models.TrackingSiteUser) error {

	resp, err := anilistPost(anilistQLGetUserList, jobj{"userId": user.ExternalID, "type": "ANIME"})
	if err != nil || resp.StatusCode != 200 {
		log.Errorf("Error occured while updating user list: %s", err)
		return err
	}

	var data jobj
	err = json.NewDecoder(resp.Body).Decode(&data)
	defer resp.Body.Close()
	if err != nil {
		log.Errorf("Error occurred while updating user list: %s", err)
		return err
	}

	// check if the user exists
	if _, ok := data["errors"]; ok {
		log.Errorf("Error occured while updating user list: %s", err)
		return errors.New("unable to find user list")
	}

	medialColl := data["data"].(jobj)["MediaListCollection"].(jobj)

	user.Username = strings.ToLower(medialColl["user"].(jobj)["name"].(string))
	scoreFormat := medialColl["user"].(jobj)["mediaListOptions"].(jobj)["scoreFormat"].(string)
	scoreScale := float64(scoreFormat2Scale(scoreFormat))

	conflictClause := clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}, {Name: "anime_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"watch_status_id", "score"}),
	}

	// raw query or https://stackoverflow.com/questions/75343438/gorm-inserting-a-subquery-result
	lists := medialColl["lists"].([]interface{})
	for l := 0; l < len(lists); l++ {
		entires := lists[l].(jobj)["entries"].([]interface{})
		for i := 0; i < len(entires); i++ {
			e := entires[i].(jobj)

			animeID := clause.Expr{
				SQL:  "(SELECT id FROM animes WHERE anilist_id = ?)",
				Vars: []interface{}{int(e["mediaId"].(float64))},
			}

			watchStatusID := clause.Expr{
				SQL:  "(SELECT id FROM anime_watch_statuses WHERE anilist = ?)",
				Vars: []interface{}{e["status"]},
			}

			values := map[string]interface{}{
				"user_id":         user.ID,
				"anime_id":        animeID,
				"watch_status_id": watchStatusID,
				"score":           float32(e["score"].(float64) / scoreScale),
			}

			err := models.SqlDB.Clauses(conflictClause).Table("anime_list_entries").Create(values).Error

			if err != nil {
				log.Info(err)
			}

		}
	}

	models.SqlDB.Save(&user)

	return nil

}
