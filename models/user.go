package models

import (
	"time"

	"gorm.io/gorm/clause"
)

type SupportedTrackingSite struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"unique"`
	Display string
}

func (s *SupportedTrackingSite) GetFromName(name string) (int, error) {
	result := SqlDB.Where("name = ?", name).First(s)

	return int(result.RowsAffected), result.Error
}

/* ----------------------------------------------------------------------------*/

type TrackingSiteUser struct {
	ID               uint `gorm:"primaryKey"`
	Username         string
	ExternalID       int `gorm:"not null;uniqueIndex:idx_user_site_unique"` // id on the tracking size
	TrackingSiteID   int `gorm:"not null;uniqueIndex:idx_user_site_unique"`
	TrackingSite     SupportedTrackingSite
	CreatedAt        time.Time
	UpdatedAt        time.Time
	AnimeListEntries []AnimeListEntry `gorm:"foreignKey:UserID"`

	isNewUser bool
}

func (u *TrackingSiteUser) Search(siteID int, username string) (int, error) {
	result := SqlDB.Where("tracking_site_id = ? and username = ?", siteID, username).Find(u)

	return int(result.RowsAffected), result.Error
}

func (u *TrackingSiteUser) LoadAnimeListIDs() error {

	return SqlDB.Model(u).Select("AnimeID").Association("AnimeListEntries").Find(&(u.AnimeListEntries))
}

// mask the user as new (requires a list update)
func (u *TrackingSiteUser) MarkAsNew() {
	u.isNewUser = true
}

func (u *TrackingSiteUser) IsNew() bool {
	return u.isNewUser
}

func (u *TrackingSiteUser) IsListOlderThan(deltaSeconds int) bool {
	return int(time.Since(u.UpdatedAt).Seconds()) > deltaSeconds
}

// create the user or update username if the tracking site and external site ids
// are already in the database
func (u *TrackingSiteUser) Create() error {

	conflictClause := clause.OnConflict{
		Columns:   []clause.Column{{Name: "external_id"}, {Name: "tracking_site_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"username"}),
	}

	return SqlDB.Clauses(conflictClause).Create(u).Error
}

/* ----------------------------------------------------------------------------*/

type AnimeListEntry struct {
	UserID        uint `gorm:"primaryKey;autoIncrement:false" json:"-"`
	User          TrackingSiteUser
	AnimeID       uint `gorm:"primaryKey;autoIncrement:false" json:"-"`
	Anime         Anime
	WatchStatusID uint `json:"-"`
	WatchStatus   AnimeWatchStatus
	Score         float32
}
