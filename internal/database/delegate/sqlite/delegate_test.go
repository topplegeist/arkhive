package sqlite_test

import (
	"os"
	"testing"
	"time"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

const TEST_FOLDER_PATH = "test"

func clearTestEnvironment() {
	os.RemoveAll(TEST_FOLDER_PATH)
}

func TestOpenAndClose(t *testing.T) {
	clearTestEnvironment()
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	s.Close()
	clearTestEnvironment()
}

func TestOpenAfterFirstCreation(t *testing.T) {
	clearTestEnvironment()
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	s.Close()
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	s.Close()
	clearTestEnvironment()
}

func TestMigrate(t *testing.T) {
	clearTestEnvironment()
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := s.Migrate(); err != nil {
		t.Log(err)
		t.Fail()
	}
	s.Close()
	clearTestEnvironment()
}

func TestFailMigration(t *testing.T) {
	clearTestEnvironment()
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Migrate(); err == nil {
		t.Fail()
	}
}

func TestFailClose(t *testing.T) {
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Close(); err == nil {
		t.Fail()
	}
}

func TestStoreImportedEmpty(t *testing.T) {
	clearTestEnvironment()
	s := sqlite.SQLiteDelegate{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := s.StoreImported([]importer.Console{}, []importer.Game{}, []importer.Tool{}); err != nil {
		t.Log(err)
		t.Fail()
	}
	s.Close()
	clearTestEnvironment()
}

func TestStoreImportedConsole(t *testing.T) {
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

	languageVariableName := "languageVariableName"
	destination := "destination"
	collectionPath := "collectionPath"
	if err := s.StoreImported(
		[]importer.Console{{
			Slug:                 "Slug",
			CoreLocation:         "CoreLocation",
			Name:                 "Name",
			SingleFile:           true,
			IsEmbedded:           true,
			LanguageVariableName: &languageVariableName,
			Plugins: []importer.ConsolePlugin{{
				Type: "Type",
				Files: []importer.ConsolePluginsFile{{
					Url:            "Url",
					Destination:    &destination,
					CollectionPath: &collectionPath,
				}}},
			},
			FileTypes: []importer.ConsoleFileType{{
				FileType: "FileType",
				Action:   "Action",
			}},
			Configs: []importer.ConsoleConfig{{
				Name:  "Name",
				Value: "Value",
				Level: "Level",
			}},
			Languages: []importer.ConsoleLanguage{{
				Tag:  1,
				Name: "Name",
			}},
		}},
		[]importer.Game{},
		[]importer.Tool{}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if entities, err := s.GetConsoles(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.Slug)
			assert.Equal(t, "CoreLocation", entity.CoreLocation)
			assert.Equal(t, "Name", entity.Name)
			assert.True(t, entity.SingleFile)
			assert.True(t, entity.IsEmbedded)
			assert.Equal(t, "languageVariableName", entity.LanguageVariableName.String)
		}
	}

	var consolePluginId uint
	if entities, err := s.GetConsolePlugins(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Positive(t, entity.Id)
			assert.Equal(t, "Slug", entity.ConsoleID)
			assert.Equal(t, "Type", entity.Type)
			consolePluginId = entity.Id
		}
	}

	if entities, err := s.GetConsolePluginsFiles(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, consolePluginId, entity.ConsolePluginID)
			assert.Equal(t, "Url", entity.Url)
			assert.Equal(t, "destination", entity.Destination.String)
			assert.Equal(t, "collectionPath", entity.CollectionPath.String)
		}
	}

	if entities, err := s.GetConsoleFileTypes(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.ConsoleID)
			assert.Equal(t, "FileType", entity.FileType)
			assert.Equal(t, "Action", entity.Action)
		}
	}

	if entities, err := s.GetConsoleConfigs(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.ConsoleID)
			assert.Equal(t, "Name", entity.Name)
			assert.Equal(t, "Value", entity.Value)
			assert.Equal(t, "Level", entity.Level)
		}
	}

	if entities, err := s.GetConsoleLanguages(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.ConsoleID)
			assert.Equal(t, 1, entity.Tag)
			assert.Equal(t, "Name", entity.Name)
		}
	}

	s.Close()
	clearTestEnvironment()
}

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

func TestStoreImportedTool(t *testing.T) {
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

	destination := "destination"
	collectionPath := "collectionPath"
	if err := s.StoreImported(
		[]importer.Console{},
		[]importer.Game{},
		[]importer.Tool{{
			Slug:           "Slug",
			Url:            "Url",
			CollectionPath: &collectionPath,
			Destination:    &destination,
			Types:          []string{"Types"},
		}}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if entities, err := s.GetTools(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.Slug)
			assert.Equal(t, "Url", entity.Url)
			assert.Equal(t, "collectionPath", entity.CollectionPath.String)
			assert.Equal(t, "destination", entity.Destination.String)
		}
	}

	if entities, err := s.GetToolFileTypes(); err != nil {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.ToolID)
			assert.Equal(t, "Types", entity.Type)
		}
	}

	s.Close()
	clearTestEnvironment()
}
