package sqlite_test

import (
	"testing"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

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
