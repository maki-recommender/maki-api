package main

import (
	"errors"
	"fmt"
	"rickycorte/maki/anime"
	"rickycorte/maki/conf"
	"rickycorte/maki/models"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/favicon"
)

// center the string given a number of spaces
func centerStr(value string, spaces int) string {
	left := int((spaces - len(value)) / 2)
	right := spaces - left - len(value)

	return fmt.Sprintf("%s%s%s", strings.Repeat(" ", left), value, strings.Repeat(" ", right))
}

func index(c *fiber.Ctx) error {
	log.Info(c.IP(), " visited index page")
	return c.JSON(fiber.Map{"message": "Hi sen(pi)!"})
}

func error404(c *fiber.Ctx) error {
	return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"message": "Not found"})
}

func errorHandler(c *fiber.Ctx, err error) error {
	// Status code defaults to 500
	code := fiber.StatusInternalServerError
	message := "Something went wrong on our side"

	// Retrieve the custom status code if it's a *fiber.Error
	var e *fiber.Error
	if errors.As(err, &e) {
		code = e.Code
		message = e.Message
	}

	return c.Status(code).JSON(fiber.Map{"message": message})
}

func main() {

	makiCfg := conf.GetConfigFromEnv()

	appName := fmt.Sprintf("Maki v%d.%d.%d", conf.VERSION_MAJOR, conf.VERSION_MINOR, conf.VERSION_FIX)
	fmt.Println(" ┌───────────────────────────────────────────────────┐")
	fmt.Printf(" |%s|\n", centerStr(appName, 51))
	fmt.Printf(" |%s|\n", centerStr("(c) rickycorte 2023", 51))
	fmt.Println(" └───────────────────────────────────────────────────┘")

	// connect and prepare databases
	sql := conf.ConnectSQLDB(makiCfg.SqlDBConnection)
	models.SetDatabase(sql)
	models.Migrate()

	// create fiber app
	app := fiber.New(fiber.Config{
		ErrorHandler: errorHandler,
	})
	app.Use(favicon.New())

	app.Get("/", index)
	anime.RegisterHandlers(app)

	//custom 404
	app.Use(error404)

	app.Listen(makiCfg.ServerAddress)
}
