package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var SqlDB *gorm.DB

func SetDatabase(db *gorm.DB) {
	SqlDB = db
}

// populate db with default data
func populateDefault() {

	var animeFormats = []AnimeFormat{
		{1, "tv", "TV", "TV"},
		{2, "short", "SHORT", "TV_SHORT"},
		{3, "movie", "MOVIE", "MOVIE"},
		{4, "special", "SPECIAL", "SPECIAL"},
		{5, "ova", "OVA", "OVA"},
		{6, "ona", "ONA", "ONA"},
		{7, "music", "MUSIC", "MUSIC"},
	}

	SqlDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&animeFormats)

	var animeAirStatuses = []AnimeAirStatus{
		{1, "complete", "COMPLETE", "FINISHED"},
		{2, "airing", "AIRING", "RELEASING"},
		{3, "not released", "NOT_RELEASED", "NOT_YET_RELEASED"},
		{4, "cancelled", "CANCELLED", "CANCELLED"},
		{5, "hiatus", "HIATUS", "HIATUS"},
	}
	SqlDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&animeAirStatuses)

	var animeWatchStatuses = []AnimeWatchStatus{
		{1, "current", "CURRENT", "CURRENT"},
		{2, "completed", "COMPLETED", "COMPLETED"},
		{3, "paused", "PAUSED", "PAUSED"},
		{4, "dropped", "DROPPED", "DROPPED"},
		{6, "planning", "PLANNING", "PLANNING"},
		{8, "rewatching", "REWATCHING", "REWATCHING"},
	}
	SqlDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&animeWatchStatuses)

	var supportedTrackingSites = []SupportedTrackingSite{
		{1, "anilist", "Anilist"},
		{2, "mal", "MyAnimeList"},
	}
	SqlDB.Clauses(clause.OnConflict{DoNothing: true}).Create(&supportedTrackingSites)

}

func Migrate() {
	SqlDB.AutoMigrate(&SupportedTrackingSite{})

	SqlDB.AutoMigrate(&Genre{})

	SqlDB.AutoMigrate(&AnimeFormat{})
	SqlDB.AutoMigrate(&AnimeAirStatus{})
	SqlDB.AutoMigrate(&AnimeWatchStatus{})
	SqlDB.AutoMigrate(&Anime{})

	populateDefault()
}
