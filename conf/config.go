package conf

import (
	"os"
	"reflect"

	"github.com/gofiber/fiber/v2/log"
)

const VERSION_MAJOR = 2
const VERSION_MINOR = 0
const VERSION_FIX = 0

type Configuration struct {
	ServerAddress   string `required:"true" default:":8080"`
	SqlDBConnection string `required:"true"`
}

// Read configuration from environment variables named like:
//	MAKI_<struct field name>
func (cfg *Configuration) ReadFromEnv() {

	log.Info("Loading configuration from env vars")

	t := reflect.TypeOf(*cfg)
	v := reflect.ValueOf(cfg)
	for i := 0; i < t.NumField(); i++ {
		fieldType := t.Field(i)

		envKey := "MAKI_" + fieldType.Name
		value := os.Getenv(envKey)
		// set default value is possible
		if value == "" && fieldType.Tag.Get("default") != "" {
			value = fieldType.Tag.Get("default")
			log.Info("[ENV CONF] Using default value for " + envKey)
		}
		// check if required and still null
		if value == "" && fieldType.Tag.Get("required") == "" {
			panic("Missing required env variable: " + envKey)
		}

		v.Elem().Field(i).SetString(value)
		log.Info("[ENV CONF] " + envKey + ": ok")
	}

	log.Info("Loaded config sucessfully")
}

// Get the app configuration from environment variables
func GetConfigFromEnv() *Configuration {

	cfg := Configuration{}
	cfg.ReadFromEnv()
	return &cfg
}
