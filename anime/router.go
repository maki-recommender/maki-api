package anime

import (
	"context"
	"errors"
	"math/rand"
	"regexp"
	"rickycorte/maki/models"
	"rickycorte/maki/protos/RecommendationService"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// TODO: set from config
var maxRecommendations = 100
var defaultRecommendations = 12

/* ----------------------------------------------------------------------------*/

func getAnimeInfoHandler(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Anime id must be an integer")
	}

	if id < 1 {
		return fiber.NewError(fiber.StatusBadRequest, "Anime id must be a positive integer")
	}

	anime := models.Anime{}
	cnt, err := anime.EagerlyGetFromID(id)
	if cnt != 1 || err != nil {
		return fiber.ErrNotFound
	}

	return c.JSON(anime)
}

/* ----------------------------------------------------------------------------*/
// recommendations

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

// check if the username is valid
func isUsernameValid(username string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	return match
}

func getOrCreateUser(site *models.SupportedTrackingSite, username string) *models.TrackingSiteUser {

	//TODO: cache in redis to speed up most frequent users' queries

	user := models.TrackingSiteUser{}
	cnt, err := user.Search(int(site.ID), username)

	if err != nil {
		log.Error(err.Error())
	}

	// check if the user is known
	if cnt == 0 {
		//TODO: create user list and syncronuslt fetch user list
		log.Infof("New user %s from %s", username, site.Name)
	} else {
		// check if list need to be updated in the background
		log.Infof("Glad to see you again %s from %s", user.Username, site.Name)
	}

	user.TrackingSite = *site

	return &user
}

func checkUserListUpates(user *models.TrackingSiteUser) {
	//TODO: implement
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

//Recommend anime to the user
//TODO: optional parameters
func recommendAnimeHandler(c *fiber.Ctx) error {
	reqUsername := strings.ToLower(c.Params("username"))

	site := models.SupportedTrackingSite{}
	cnt, err := site.GetFromName(c.Params("site"))
	if cnt == 0 || err != nil {
		log.Info("Rejected recommendation request for unknown tracking site")
		return fiber.NewError(fiber.StatusBadRequest, "Unkown tracking site")
	}

	if !isUsernameValid(reqUsername) {
		log.Info("Rejected recommendation request for not allowed username")
		return fiber.NewError(fiber.StatusBadRequest, "Username not allowed")
	}

	//TODO: take k and validate

	user := getOrCreateUser(&site, reqUsername)
	checkUserListUpates(user)

	recs, err := recommendAnimeToUser(user, defaultRecommendations)

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "Something went wrong on our side")
	}

	return c.JSON(recs)
}

/* ----------------------------------------------------------------------------*/
// configuration

//Register api handlers for /anime subpaths
func RegisterHandlers(app *fiber.App) {
	animeRouter := app.Group("/anime")
	animeRouter.Get("/data/:id<int>", getAnimeInfoHandler)
	animeRouter.Get("/:site<minLen(3)>/:username<minLen(4),maxLen(30)>", recommendAnimeHandler)
}
