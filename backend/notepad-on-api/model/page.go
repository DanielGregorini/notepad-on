package model

import (
    "time"
)

type Page struct {
    ID        int       `gorm:"primaryKey;autoIncrement" json:"id"`
    Slug      string    `gorm:"type:varchar(255);not null;unique" json:"slug"`
    Content   string    `gorm:"type:text;not null" json:"content"`
    CreatedAt time.Time `json:"created_at"`
    UpdatedAt time.Time `json:"updated_at"`
}
