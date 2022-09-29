package engine_test

import (
	"sync"
	"testing"

	"arkhive.dev/launcher/internal/console"
	"arkhive.dev/launcher/internal/database/importer"
	"arkhive.dev/launcher/internal/database/mock"
	"arkhive.dev/launcher/internal/engine"
	"arkhive.dev/launcher/internal/game"
	"arkhive.dev/launcher/internal/tool"
	"github.com/stretchr/testify/assert"
)

const TEST_FOLDER_PATH = "test"

func TestInitializeUnreacheableDatabase(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			assert.Equal(t, "cannot open database", r)
		}
	}()
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mock.MockDelegate{
		FailOpen: true,
	}, []importer.Importer{})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	t.Fail()
}

func TestInitializeNoImporters(t *testing.T) {
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mock.MockDelegate{
		HashCalculated: true,
	}, []importer.Importer{})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
}

func TestInitializeImporterCannotImport(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImportValue: false,
	}
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mock.MockDelegate{
		HashCalculated: true,
	}, []importer.Importer{&mockImporter})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	assert.False(t, mockImporter.HasImported)
}

func TestInitializeImporterReturningInvalidDatabase(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImportValue: true,
		DatabaseData:   []byte("{a}"),
	}
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.True(t, mockImporter.HasImported)
	}()
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mock.MockDelegate{
		HashCalculated: true,
	}, []importer.Importer{&mockImporter})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	t.Fail()
}

func TestInitializeImporterReturningEmptyDatabase(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImportValue:  true,
		DatabaseData:    []byte("{}"),
		EncryptedDBHash: []byte("Fake hash"),
	}
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mock.MockDelegate{
		HashCalculated: true,
	}, []importer.Importer{&mockImporter})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	assert.True(t, mockImporter.HasImported)
}

func TestInitializeImporterReturningConsoleAndGameDatabase(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImportValue: true,
		DatabaseData: []byte(`
		{
			"consoles": {
				"core": {
					"name": "Core",
					"core_location": "core_location",
					"single_file": true
				}
			},
			"games": {
				"videogame": {
					"name": "Videogame",
					"console_slug": "core",
					"background_color": "#ffffff",
					"url": "url"
				}
			}
		}
        `),
		EncryptedDBHash: []byte("Fake hash"),
	}
	mockDelegate := mock.MockDelegate{
		HashCalculated: true,
	}
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mockDelegate, []importer.Importer{&mockImporter})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	assert.True(t, mockImporter.HasImported)
	var consoles []console.Console
	mockDelegate.List(&consoles)
	assert.NotEmpty(t, consoles)
	assert.Equal(t, "core", consoles[0].Slug)
	assert.Equal(t, "Core", consoles[0].Name)
	assert.Equal(t, "core_location", consoles[0].CoreLocation)
	assert.True(t, consoles[0].SingleFile)
	var games []game.Game
	mockDelegate.List(&games)
	assert.NotEmpty(t, games)
	assert.Equal(t, "videogame", games[0].Slug)
	assert.Equal(t, "Videogame", games[0].Name)
	assert.Equal(t, "core", games[0].ConsoleID)
	assert.Equal(t, "#ffffff", games[0].BackgroundColor)
	var gamesDisks []game.GameDisk
	mockDelegate.List(&gamesDisks)
	assert.NotEmpty(t, gamesDisks)
	assert.Equal(t, "url", gamesDisks[0].Url)
}

func TestInitializeImporterReturningToolDatabase(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImportValue: true,
		DatabaseData: []byte(`
		{
			"win_tools": {
				"tool": {
					"url": "url"
				}
			}
		}
        `),
		EncryptedDBHash: []byte("Fake hash"),
	}
	mockDelegate := mock.MockDelegate{
		HashCalculated: true,
	}
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mockDelegate, []importer.Importer{&mockImporter})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	assert.True(t, mockImporter.HasImported)
	var tools []tool.Tool
	mockDelegate.List(&tools)
	assert.NotEmpty(t, tools)
	assert.Equal(t, "tool", tools[0].Slug)
	assert.Equal(t, "url", tools[0].Url)
}
