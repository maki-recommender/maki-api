package models

import (
	"database/sql"
	"time"
)

type AnimeFormat struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"unique"`
	Mal     string `gorm:"unique"`
	Anilist string `gorm:"unique"`
}

type AnimeAirStatus struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"unique"`
	Mal     string `gorm:"unique"`
	Anilist string `gorm:"unique"`
}

type AnimeWatchStatus struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"unique"`
	Mal     string `gorm:"unique"`
	Anilist string `gorm:"unique"`
}

type Genre struct {
	ID   uint   `gorm:"primaryKey"`
	Name string `gorm:"unique"`
}

type Anime struct {
	ID                     uint   `gorm:"primaryKey"`
	AnilistID              uint   `gorm:"index"`
	MalID                  uint   `gorm:"index"`
	Title                  string `gorm:"not null"`
	AnilistCover           sql.NullString
	MalCover               sql.NullString
	AnilistDescription     sql.NullString
	MalDescription         sql.NullString
	FormatID               uint
	Format                 AnimeFormat
	ReleaseYear            sql.NullInt64
	StatusID               uint
	Status                 AnimeAirStatus
	AnilistNormalizedScore float32
	MalNormalizedScore     float32
	CreatedAt              time.Time
	UpdatedAt              time.Time
	Genres                 []Genre `gorm:"many2many:anime_genres;"`
}

func (a *Anime) GetFromID(id int) (int, error) {
	result := SqlDB.Where("id = ?", id).Find(a)
	return int(result.RowsAffected), result.Error
}
