package anime

import (
	"rickycorte/maki/models"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

func getAnimeInfo(c *fiber.Ctx) error {
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

func recommendAnime(c *fiber.Ctx) error {
	return fiber.ErrInternalServerError
}

//Register api handlers for /anime subpaths
func RegisterHandlers(app *fiber.App) {
	animeRouter := app.Group("/anime")
	animeRouter.Get("/data/:id<int>", getAnimeInfo)
	animeRouter.Get("/data/:site<minLen(3)>/:username<minLen(4)>", getAnimeInfo)
}
