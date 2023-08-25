package anime

import (
	"context"
	"errors"
	"math/rand"
	"rickycorte/maki/datafetch"
	"rickycorte/maki/models"
	"rickycorte/maki/protos/RecommendationService"
	"time"

	"github.com/gofiber/fiber/v2/log"
)

type AnimeRecommendationResult struct {
	ID              int                `json:"id"`
	Username        string             `json:"username"`
	Site            string             `json:"site"`
	LastListUpdate  time.Time          `json:"last_list_update"`
	K               int                `json:"k"`
	Recommendations []RecommendedAnime `json:"recommendations"`
}

type RecommendedAnime struct {
	ID          uint     `json:"id"`
	Anilist     uint     `json:"anilist"`
	Mal         uint     `json:"mal"`
	Title       string   `json:"title"`
	CoverUrl    string   `json:"cover_url"`
	Format      string   `json:"format"`
	ReleaseYear *int     `json:"release_year"`
	Score       int      `json:"score"`
	Genres      []string `json:"genres"`
	Affinity    float32  `json:"affinity"`
}

func (ra *RecommendedAnime) FromPair(a *models.Anime, r *RecommendationService.RecommendedItem) {
	ra.ID = a.ID
	ra.Anilist = a.AnilistID
	ra.Mal = a.MalID
	ra.Title = a.Title
	ra.CoverUrl = *a.AnilistCover
	ra.Format = a.Format.Name
	ra.ReleaseYear = a.ReleaseYear
	ra.Score = int(a.AnilistNormalizedScore * 100)
	for i := 0; i < len(a.Genres); i++ {
		ra.Genres = append(ra.Genres, a.Genres[i].Name)
	}
	ra.Affinity = r.Score
}

func (ra *RecommendedAnime) FromPairPreferMal(a *models.Anime, r *RecommendationService.RecommendedItem) {
	ra.FromPair(a, r)
	if a.MalCover != nil {
		ra.CoverUrl = *a.MalCover
	}
	if a.MalNormalizedScore != 0 {
		ra.Score = int(a.MalNormalizedScore * 100)
	}
}

//get db user, may return nil, nil if user was not found and no error occurred
func getDBUser(site *models.SupportedTrackingSite, username string) (*models.TrackingSiteUser, error) {

	//TODO: cache in redis to speed up most frequent users' queries
	user := models.TrackingSiteUser{}
	cnt, err := user.Search(int(site.ID), username)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	if cnt == 0 {
		return nil, nil
	}

	user.TrackingSite = *site

	return &user, nil
}

//  creates one by checking if the user actualy exists on the tracking site
func createValidDBUser(site *models.SupportedTrackingSite, username string) (*models.TrackingSiteUser, error) {

	id, err := datafetch.GetUserId(site.Name, username)
	if err != nil {
		return nil, err
	}

	user := models.TrackingSiteUser{
		Username:     username,
		TrackingSite: *site,
		ExternalID:   id,
	}
	user.MarkAsNew()

	user.Create()

	log.Infof("Created new %s user: %s", site.Name, username)

	return &user, nil
}

// this function will return errors only for foreground requests
func checkUserListUpates(user *models.TrackingSiteUser) error {
	// sync list in foreground if user is new
	if user.IsNew() {
		return datafetch.UpdateAnimeList(user)
	} else {
		// sync in background if list is considered outdated
		if user.IsListOlderThan(listIsOldAfterSeconds) {
			go datafetch.UpdateAnimeList(user)
		}
	}

	return nil
}

func animeList2RPCWatchList(animeList []models.AnimeListEntry) *RecommendationService.WatchedAnime {
	watchList := RecommendationService.WatchedAnime{}
	for i := 0; i < len(animeList); i++ {
		watchList.Items = append(
			watchList.Items,
			&RecommendationService.Item{Id: uint32(animeList[i].AnimeID)},
		)
	}

	return &watchList
}

func generateNewRecommendations(user *models.TrackingSiteUser) (*RecommendationService.RecommendedAnime, error) {
	// prepare recommendations
	recService, err := RecommendationService.GetRecommendationServiceClient()
	if err != nil {
		log.Error("Recommendation service not available")
		return nil, errors.New("recommendation service not available")
	}

	// load user list if is not already available
	if len(user.AnimeListEntries) == 0 {
		user.LoadAnimeListIDs()
		log.Infof("Loaded %d list items from the db for user %s", len(user.AnimeListEntries), user.Username)
	}

	watchList := animeList2RPCWatchList(user.AnimeListEntries)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	return recService.GetAnimeRecommendations(ctx, watchList)
}

func recommendAnimeToUser(user *models.TrackingSiteUser, k int) (*AnimeRecommendationResult, error) {

	//TODO: check redis cache
	recs, err := generateNewRecommendations(user)

	log.Infof("Got %d fresh recommendations for %s user %s", len(recs.Items), user.TrackingSite.Name, user.Username)

	if err != nil {
		return nil, err
	}

	if k > len(recs.Items) {
		k = len(recs.Items)
	}

	// apply filters here

	// shuffle fist top 2 * maxRecommendations items to give user a bit of variability
	recItems := recs.Items[:k]
	rand.Seed(time.Now().UnixNano())
	rand.Shuffle(len(recItems), func(i, j int) { recItems[i], recItems[j] = recItems[j], recItems[i] })

	//convert to list of its
	var ids []int = make([]int, k)
	for i := 0; i < k; i++ {
		ids[i] = int(recItems[i].Id)
	}

	// fetch data from db
	animes, err := models.EagerlyGetAnimesFromIDs(ids)
	if err != nil {
		return nil, err
	}

	recommendations := AnimeRecommendationResult{
		int(user.ID),
		user.Username,
		user.TrackingSite.Name,
		user.UpdatedAt,
		k,
		make([]RecommendedAnime, k),
	}
	// populate list by pairing data from db to reccomendations
	for i := 0; i < k; i++ {
		for j := 0; j < len(animes); j++ {
			if animes[j].ID == uint(recItems[i].Id) {
				recommendations.Recommendations[i].FromPair(&animes[j], recItems[i])
			}
		}
	}

	return &recommendations, nil
}
