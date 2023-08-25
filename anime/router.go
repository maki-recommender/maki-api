package anime

import (
	"regexp"
	"rickycorte/maki/models"
	"strconv"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
)

// TODO: set from config
var maxRecommendations = 100
var defaultRecommendations = 12
var listIsOldAfterSeconds = 15 * 60

/* ----------------------------------------------------------------------------*/
// validation

// check if the username is valid
func isUsernameValid(username string) bool {
	match, _ := regexp.MatchString("^[a-zA-Z0-9_-]+$", username)
	return match
}

// convert and check the requested number of reccomendations
// on error return the standard number of recommendations
func validReccommendationNumber(value string, standard, max int) int {
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
		return fiber.NewError(fiber.StatusBadRequest, "Unkown tracking site")
	}

	if !isUsernameValid(reqUsername) {
		log.Info("Rejected recommendation request for not allowed username")
		return fiber.NewError(fiber.StatusBadRequest, "Username not allowed")
	}

	// read required number of recommendations from query and validate
	k := validReccommendationNumber(c.Query("k"), defaultRecommendations, maxRecommendations)

	user, err := getDBUser(&site, reqUsername)
	if err != nil {
		return fiber.ErrInternalServerError
	}

	// unkown user -> must check if there is a valid user on the tracking site
	if user == nil {
		user, err = createValidDBUser(&site, reqUsername)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "Unable to find the required user")
		}
	}

	err = checkUserListUpates(user)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Unable to find the required user anime list")
	}

	//TODO: filter hentai if not requested, filter items without mal ids for mal users
	recs, err := recommendAnimeToUser(user, k)

	if err != nil {
		return fiber.ErrInternalServerError
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
