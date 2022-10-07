package sqlite_test

import (
	"testing"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

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

	if entities, err := s.GetConsoles(); err != nil || len(entities) == 0 {
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
	if entities, err := s.GetConsolePlugins(); err != nil || len(entities) == 0 {
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

	if entities, err := s.GetConsolePluginsFiles(); err != nil || len(entities) == 0 {
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

	if entities, err := s.GetConsoleFileTypes(); err != nil || len(entities) == 0 {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.ConsoleID)
			assert.Equal(t, "FileType", entity.FileType)
			assert.Equal(t, "Action", entity.Action)
		}
	}

	if entities, err := s.GetConsoleConfigs(); err != nil || len(entities) == 0 {
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

	if entities, err := s.GetConsoleLanguages(); err != nil || len(entities) == 0 {
		t.Log(err)
		t.Fail()
	} else {
		for _, entity := range entities {
			assert.Equal(t, "Slug", entity.ConsoleID)
			assert.Equal(t, uint(1), entity.Tag)
			assert.Equal(t, "Name", entity.Name)
		}
	}

	s.Close()
	clearTestEnvironment()
}
