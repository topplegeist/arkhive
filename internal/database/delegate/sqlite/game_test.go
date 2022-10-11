package sqlite_test

import (
	"testing"
	"time"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

type GameTestFlags struct {
	ImportDisks           bool
	ImportConfigs         bool
	ImportAdditionalFiles bool
}

func storeImportedGameTestProthotype(t *testing.T, flags GameTestFlags) {
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

	var disks []importer.GameDisk
	if flags.ImportDisks {
		disks = append(disks, importer.GameDisk{
			DiskNumber:     1,
			Url:            "Url",
			Image:          &image,
			CollectionPath: &collectionPath,
		})
	}

	var configs []importer.GameConfig
	if flags.ImportDisks {
		configs = append(configs, importer.GameConfig{
			Name:  "Name",
			Value: "Value",
		})
	}

	var additionalFiles []importer.GameAdditionalFile
	if flags.ImportAdditionalFiles {
		additionalFiles = append(additionalFiles, importer.GameAdditionalFile{
			Name: "Name",
			Data: []byte("Data"),
		})
	}

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
			Disks:           disks,
			Configs:         configs,
			AdditionalFiles: additionalFiles,
		}},
		[]importer.Tool{}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if entities, err := s.GetGames(); err != nil || len(entities) == 0 {
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

	if entities, err := s.GetGameDisks(); err != nil || (len(entities) == 0 && flags.ImportDisks) {
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

	if entities, err := s.GetGameConfigs(); err != nil || (len(entities) == 0 && flags.ImportConfigs) {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.GameID)
			assert.Equal(t, "Name", entity.Name)
			assert.Equal(t, "Value", entity.Value)
		}
	}

	if entities, err := s.GetGameAdditionalFiles(); err != nil || (len(entities) == 0 && flags.ImportAdditionalFiles) {
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

func TestStoreImportedGame(t *testing.T) {
	storeImportedGameTestProthotype(t, GameTestFlags{
		ImportDisks:           true,
		ImportConfigs:         true,
		ImportAdditionalFiles: true,
	})
}
