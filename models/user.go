package models

type SupportedTrackingSite struct {
	ID      uint   `gorm:"primaryKey"`
	Name    string `gorm:"unique"`
	Display string
}
