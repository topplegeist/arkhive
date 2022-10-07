package sqlite_test

import (
	"testing"
	"time"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

func TestStoreImportedGame(t *testing.T) {
	clearTestEnvironment()
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := s.Migrate(); err != nil {
		t.Fail()
	}

	backgroundImage := "backgroundImage"
	logo := "logo"
	executable := "executable"
	image := "image"
	collectionPath := "collectionPath"
	startingTime := time.Now().UnixNano()
	if err := s.StoreImported(
		[]importer.Console{},
		[]importer.Game{{
			Slug:            "Slug",
			Name:            "Name",
			ConsoleSlug:     "ConsoleSlug",
			BackgroundColor: "BackgroundColor",
			BackgroundImage: &backgroundImage,
			Logo:            &logo,
			Executable:      &executable,
			Disks: []importer.GameDisk{{
				DiskNumber:     1,
				Url:            "Url",
				Image:          &image,
				CollectionPath: &collectionPath,
			}},
			Configs: []importer.GameConfig{{
				Name:  "Name",
				Value: "Value",
			}},
			AdditionalFiles: []importer.GameAdditionalFile{{
				Name: "Name",
				Data: []byte("Data"),
			}},
		}},
		[]importer.Tool{}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if entities, err := s.GetGames(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.Slug)
			assert.Equal(t, "Name", entity.Name)
			assert.Equal(t, "ConsoleSlug", entity.ConsoleID)
			assert.Equal(t, "BackgroundColor", entity.BackgroundColor)
			assert.Equal(t, "backgroundImage", entity.BackgroundImage.String)
			assert.Equal(t, "logo", entity.Logo.String)
			assert.Equal(t, "executable", entity.Executable.String)
			assert.LessOrEqual(t, startingTime, entity.InsertionDate.UnixNano())
			assert.GreaterOrEqual(t, time.Now().UnixNano(), entity.InsertionDate.UnixNano())
		}
	}

	if entities, err := s.GetGameDisks(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.GameID)
			assert.Equal(t, uint(1), entity.DiskNumber)
			assert.Equal(t, "Url", entity.Url)
			assert.Equal(t, "image", entity.Image.String)
			assert.Equal(t, "collectionPath", entity.CollectionPath.String)
		}
	}

	if entities, err := s.GetGameConfigs(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.GameID)
			assert.Equal(t, "Name", entity.Name)
			assert.Equal(t, "Value", entity.Value)
		}
	}

	if entities, err := s.GetGameAdditionalFiles(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.GameID)
			assert.Equal(t, "Name", entity.Name)
			assert.Equal(t, []byte("Data"), entity.Data)
		}
	}

	s.Close()
	clearTestEnvironment()
}
