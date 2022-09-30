package importer

import (
	"time"
)

type Game struct {
	Slug            string
	Name            string
	ConsoleSlug     string
	BackgroundColor string
	BackgroundImage string
	Logo            string
	Executable      string
	InsertionDate   time.Time
}
