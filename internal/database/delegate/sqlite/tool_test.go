package sqlite_test

import (
	"testing"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

type ToolTestFlags struct {
	ImportTypes bool
}

func storeImportedToolTestProthotype(t *testing.T, flags ToolTestFlags) {
	clearTestEnvironment()
	s := sqlite.SQLite{
		BasePath: TEST_FOLDER_PATH,
	}
	if err := s.Open(); err != nil {
		t.Log(err)
		t.Fail()
	}
	if err := s.Migrate(); err != nil {
		t.Fail()
	}
	var types []string
	if flags.ImportTypes {
		types = append(types, "Types")
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
			Types:          types,
		}}); err != nil {
		t.Log(err)
		t.Fail()
	}

	if entities, err := s.GetTools(); err != nil || len(entities) == 0 {
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

	if entities, err := s.GetToolFileTypes(); err != nil || len(entities) == 0 {
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

func TestStoreImportedTool(t *testing.T) {
	storeImportedToolTestProthotype(t, ToolTestFlags{
		ImportTypes: true,
	})
}
