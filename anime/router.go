package anime

import (
	"regexp"
	"rickycorte/maki/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
)

/* ----------------------------------------------------------------------------*/
// validation

// check if the username is valid
func isUsernameValid(username string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	return match
}

// convert and check the requested number of reccomendations
// on error return the standard number of recommendations
func validRecommendationNumber(value string, standard, max int) int {
	// take k from query string and validate
	if value == "" {
		return standard
	}

	k, err := strconv.Atoi(value)
	if err != nil {
		return standard
	}

	if k < 1 {
		k = 1
	} else if k > max {
		k = max
	}

	return k
}

// return a valid genre or "" if invalid
func validGenre(value string) string {
	if value == "" {
		return value
	}

	value = strings.ToLower(value)
	value = strings.Replace(value, " ", "_", -1)

	if IsValidGenre(value) {
		return value
	}

	return ""
}

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

//Recommend anime to the user
//TODO: optional parameters
func recommendAnimeHandler(c *fiber.Ctx) error {
	reqUsername := strings.ToLower(c.Params("username"))

	site := models.SupportedTrackingSite{}
	cnt, err := site.GetFromName(c.Params("site"))
	if cnt == 0 || err != nil {
		log.Info("Rejected recommendation request for unknown tracking site")
		return fiber.NewError(fiber.StatusBadRequest, "Unknown tracking site")
	}

	if !isUsernameValid(reqUsername) {
		log.Info("Rejected recommendation request for not allowed username")
		return fiber.NewError(fiber.StatusBadRequest, "Username not allowed")
	}

	user, err := getDBUser(&site, reqUsername)
	if err != nil {
		log.Error(err)
		return fiber.ErrInternalServerError
	}

	// unkown user -> must check if there is a valid user on the tracking site
	if user == nil {
		user, err = createValidDBUser(&site, reqUsername)
		if err != nil {
			log.Error(err)
			return fiber.NewError(fiber.StatusBadRequest, "Unable to find the required user")
		}
	}

	err = checkUserListUpates(user)
	if err != nil {
		log.Error(err)
		return fiber.NewError(fiber.StatusBadRequest, "Unable to find the required user anime list")
	}

	genre := validGenre(c.Query("genre"))
	defRec := cfg.DefaultRecommendations()
	maxRec := cfg.MaxRecommendations()

	filters := RecommendationFilter{
		K:            validRecommendationNumber(c.Query("k"), defRec, maxRec),
		KRandomBound: 2 * maxRec,
		Shuffle:      true,
		OnlyMal:      user.TrackingSite.Name == "mal",
		NoHentai:     genre != "hentai",
		Genre:        genre,
	}
	recs, err := recommendAnimeToUser(user, &filters)

	if err != nil {
		return fiber.ErrInternalServerError
	}

	return c.JSON(recs)
}

/* ----------------------------------------------------------------------------*/
// configuration

var cfg RecommendationConfig

type RecommendationConfig interface {
	Redis() *redis.Client

	MaxRecommendations() int
	DefaultRecommendations() int
	ListIsOldAfterSeconds() int
	// time before cache key is refreshed
	RecommendationExpireSeconds() int
	// time after cache entry is purged from redis to keep usage low
	CacheExpireSeconds() int
}

//Register api handlers for /anime subpaths
func RegisterHandlers(app *fiber.App) {
	animeRouter := app.Group("/anime")
	animeRouter.Get("/data/:id<int>", getAnimeInfoHandler)
	animeRouter.Get("/:site<minLen(3)>/:username<minLen(4),maxLen(30)>", recommendAnimeHandler)

	go perdiocallyRefreshAnimeCache()
}

func RegisterConfig(config RecommendationConfig) {
	cfg = config
}
