package conf

import (
	"os"
	"reflect"
	"strconv"

	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const VERSION_MAJOR = 2
const VERSION_MINOR = 0
const VERSION_FIX = 0

type Configuration struct {
	ServerAddress                string `required:"true" default:":8080"`
	SqlDBConnection              string `required:"true"`
	RecommendationServiceAddress string `required:"true" default:"0.0.0.0:50051"`
	RedisDBConnection            string `required:"true"`

	// recommedations settings
	MaxRecommendations               int64 `default:"100"`
	DefaultRecommendations           int64 `default:"12"`
	ListIsOldAfterSeconds            int64 `default:"3600"`
	RecommendationCacheExpireSeconds int64 `default:"86400"`
	CacheClearAfterSeconds           int64 `default:"604800"`
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

		if fieldType.Type == reflect.TypeOf("") {
			v.Elem().Field(i).SetString(value)
		} else if fieldType.Type == reflect.TypeOf(int64(0)) {
			val, err := strconv.Atoi(value)
			if err != nil {
				panic("Unable to cast env variable")
			}

			v.Elem().Field(i).SetInt(int64(val))
		}

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

/******************************************************************************/

type GlobalConfiguration struct {
	Cfg       *Configuration
	SQLConn   *gorm.DB
	RedisConn *redis.Client
}

var GCong GlobalConfiguration

func (c *GlobalConfiguration) Redis() *redis.Client {
	return c.RedisConn
}

func (c *GlobalConfiguration) MaxRecommendations() int {
	return int(c.Cfg.MaxRecommendations)
}

func (c *GlobalConfiguration) DefaultRecommendations() int {
	return int(c.Cfg.DefaultRecommendations)
}

func (c *GlobalConfiguration) ListIsOldAfterSeconds() int {
	return int(c.Cfg.ListIsOldAfterSeconds)
}

func (c *GlobalConfiguration) RecommendationExpireSeconds() int {
	return int(c.Cfg.RecommendationCacheExpireSeconds)
}

func (c *GlobalConfiguration) CacheExpireSeconds() int {
	return int(c.Cfg.CacheClearAfterSeconds)
}

func LoadGlobalConfigFormEnv() *GlobalConfiguration {
	GCong.Cfg = GetConfigFromEnv()
	GCong.SQLConn = ConnectSQLDB(GCong.Cfg.SqlDBConnection)
	GCong.RedisConn = ConnectRedisDB(GCong.Cfg.RedisDBConnection)

	return &GCong
}
