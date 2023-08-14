package models

import "time"

type SupportedTrackingSite struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"unique"`
	Display string
}

type TrackingSiteUser struct {
	ID             uint   `gorm:"primaryKey"`
	Username       string `gorm:"not null, uniqueIndex:user_site_unique"`
	ExternalID     string // id on the tracking size
	TrackingSiteID int    `gorm:"not null, uniqueIndex:user_site_unique"`
	TrackingSite   SupportedTrackingSite
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type AnimeListEntry struct {
	UserID        uint `gorm:"primaryKey;autoIncrement:false" json:"-"`
	User          TrackingSiteUser
	AnimeID       uint `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Anime         Anime
	WatchStatusID uint `json:"-"`
	WatchStatus   AnimeWatchStatus
	Score         float32
}
