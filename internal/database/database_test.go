package database_test

import (
	"errors"
	"sync"
	"testing"

	"arkhive.dev/launcher/internal/database"
	"arkhive.dev/launcher/internal/database/importer"
	"arkhive.dev/launcher/internal/database/mock"
	"github.com/stretchr/testify/assert"
)

const TEST_FOLDER_PATH = "test"

func baseInitialize(instance *database.Database) {
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
}

func TestInitializeUnreacheableDatabase(t *testing.T) {
	defer func() {
		errorString := recover().(error).Error()
		assert.Equal(t, "cannot open", errorString)
	}()
	instance := database.NewDatabase(&mock.MockDelegate{
		FailOpen: true,
		Error:    errors.New("cannot open"),
	}, []importer.Importer{})
	baseInitialize(instance)
	t.Fail()
}

func TestInitializeCannotMigrate(t *testing.T) {
	defer func() {
		errorString := recover().(error).Error()
		assert.Equal(t, "cannot migrate", errorString)
	}()
	instance := database.NewDatabase(&mock.MockDelegate{
		FailMigration: true,
		Error:         errors.New("cannot migrate"),
	}, []importer.Importer{})
	baseInitialize(instance)
	t.Fail()
}

func TestInitializeCannotReadDBHash(t *testing.T) {
	defer func() {
		errorString := recover().(error).Error()
		assert.Equal(t, "cannot get stored db hash", errorString)
	}()
	instance := database.NewDatabase(&mock.MockDelegate{
		CurrentHash: nil,
		Error:       errors.New("cannot get stored db hash"),
	}, []importer.Importer{})
	baseInitialize(instance)
	t.Fail()
}

func TestInitializeNoImporters(t *testing.T) {
	delegate := mock.MockDelegate{
		CurrentHash: &[]byte{},
	}
	instance := database.NewDatabase(&delegate, []importer.Importer{})
	baseInitialize(instance)
	assert.False(t, delegate.Stored)
}

func TestInitializeCannotImport(t *testing.T) {
	delegate := mock.MockDelegate{
		CurrentHash: &[]byte{},
	}
	mockImporter := mock.MockImporter{}
	instance := database.NewDatabase(&delegate, []importer.Importer{&mockImporter})
	baseInitialize(instance)
	assert.False(t, delegate.Stored)
}

func TestInitializeInvalidDatabase(t *testing.T) {
	mockImporter := mock.MockImporter{
		Error: errors.New("invalid database"),
	}
	defer func() {
		errorString := recover().(error).Error()
		assert.Equal(t, "invalid database", errorString)
	}()
	instance := database.NewDatabase(&mock.MockDelegate{
		CurrentHash: &[]byte{},
	}, []importer.Importer{&mockImporter})
	baseInitialize(instance)
	t.Fail()
}

func TestInitializeCannotStoreImported(t *testing.T) {
	delegate := mock.MockDelegate{
		CurrentHash:       &[]byte{},
		FailStoreImported: true,
		Error:             errors.New("cannot store imported"),
	}
	mockImporter := mock.MockImporter{
		EncryptedDBHash: &[]byte{},
	}
	instance := database.NewDatabase(&delegate, []importer.Importer{&mockImporter})
	baseInitialize(instance)
	assert.False(t, delegate.Stored)
}

func TestInitializeCannotStoreDBHash(t *testing.T) {
	delegate := mock.MockDelegate{
		CurrentHash:     &[]byte{},
		FailStoreDbHash: true,
		Error:           errors.New("cannot store db hash"),
	}
	hash := []byte("Fake hash")
	mockImporter := mock.MockImporter{
		EncryptedDBHash: &hash,
	}
	defer func() {
		errorString := recover().(error).Error()
		assert.Equal(t, "cannot store db hash", errorString)
	}()
	instance := database.NewDatabase(&delegate, []importer.Importer{&mockImporter})
	baseInitialize(instance)
	t.Fail()
}

func TestInitializeImporterSuccessful(t *testing.T) {
	delegate := mock.MockDelegate{
		CurrentHash: &[]byte{},
	}
	hash := []byte("Fake hash")
	mockImporter := mock.MockImporter{
		EncryptedDBHash: &hash,
	}
	instance := database.NewDatabase(&delegate, []importer.Importer{&mockImporter})
	baseInitialize(instance)
	assert.EqualValues(t, hash, *delegate.CurrentHash)
	assert.True(t, delegate.Stored)
}

func TestInitializeJustOneImporterSuccessful(t *testing.T) {
	delegate := mock.MockDelegate{
		CurrentHash: &[]byte{},
	}
	hash := []byte("Fake hash")
	mockImporter1 := mock.MockImporter{}
	mockImporter2 := mock.MockImporter{
		EncryptedDBHash: &hash,
	}
	instance := database.NewDatabase(&delegate, []importer.Importer{&mockImporter1, &mockImporter2})
	baseInitialize(instance)
	assert.EqualValues(t, hash, *delegate.CurrentHash)
	assert.True(t, delegate.Stored)
}
