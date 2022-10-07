package sqlite_test

import (
	"os"
	"testing"

	"arkhive.dev/launcher/internal/database/delegate/sqlite"
	"arkhive.dev/launcher/internal/database/importer"
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
