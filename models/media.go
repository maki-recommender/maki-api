package models

import (
	"database/sql"
	"time"

	"gorm.io/gorm/clause"
)

type AnimeFormat struct {
	ID      uint   `gorm:"primaryKey" json:"-"`
	Name    string `gorm:"unique"`
	Mal     string `gorm:"unique" json:"-"`
	Anilist string `gorm:"unique" json:"-"`
}

/* ----------------------------------------------------------------------------*/

type AnimeAirStatus struct {
	ID      uint   `gorm:"primaryKey" json:"-"`
	Name    string `gorm:"unique"`
	Mal     string `gorm:"unique" json:"-"`
	Anilist string `gorm:"unique" json:"-"`
}

/* ----------------------------------------------------------------------------*/

type AnimeWatchStatus struct {
	ID      uint   `gorm:"primaryKey" json:"-"`
	Name    string `gorm:"unique"`
	Mal     string `gorm:"unique" json:"-"`
	Anilist string `gorm:"unique" json:"-"`
}

/* ----------------------------------------------------------------------------*/

type Genre struct {
	ID   uint   `gorm:"primaryKey" json:"-"`
	Name string `gorm:"unique"`
}

/* ----------------------------------------------------------------------------*/

type Anime struct {
	ID                     uint   `gorm:"primaryKey"`
	AnilistID              uint   `gorm:"unique"`
	MalID                  uint   `gorm:"index"`
	Title                  string `gorm:"not null"`
	AnilistCover           *string
	MalCover               *string
	AnilistDescription     *string
	MalDescription         *string
	FormatID               uint `json:"-"`
	Format                 AnimeFormat
	ReleaseYear            *int
	StatusID               uint `json:"-"`
	Status                 AnimeAirStatus
	AnilistNormalizedScore float32
	MalNormalizedScore     float32
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Genres                 []Genre `gorm:"many2many:anime_genres;"`
}

// Eagerly get anime data from database
func (a *Anime) EagerlyGetFromID(id int) (int, error) {
	result := SqlDB.Preload("Genres").Preload("Format").Preload("Status").Where("id = ?", id).Find(a)
	return int(result.RowsAffected), result.Error
}

func EagerlyGetAnimesFromIDs(ids []int) ([]Anime, error) {
	var items []Anime

	result := SqlDB.Preload(clause.Associations).Where("id IN ?", ids).Find(&items)
	return items, result.Error
}

func GetAnimeCacheRows() (*sql.Rows, error) {
	return SqlDB.Raw(`select  animes.id, mal_id is not null, array_agg(genres.name)
		from animes join anime_genres on animes.id = anime_id JOIN genres on genres.id = genre_id
		group by animes.id`).Rows()
}
