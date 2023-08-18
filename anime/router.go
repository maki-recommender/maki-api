package anime

import (
	"rickycorte/maki/models"
	"strconv"
	"strings"

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
