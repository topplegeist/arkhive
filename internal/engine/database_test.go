package engine_test

import (
	"errors"
	"sync"
	"testing"

	"arkhive.dev/launcher/internal/database/importer"
	"arkhive.dev/launcher/internal/database/mock"
	"arkhive.dev/launcher/internal/engine"
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
		CanImport: false,
	}
	instance := engine.NewDatabase(TEST_FOLDER_PATH, &mock.MockDelegate{
		HashCalculated: true,
	}, []importer.Importer{&mockImporter})
	defer instance.Deinitialize()
	waitGroup := sync.WaitGroup{}
	waitGroup.Add(1)
	instance.Initialize(&waitGroup)
	waitGroup.Wait()
	assert.False(t, mockImporter.Imported)
}

func TestInitializeImporterReturningInvalidDatabase(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImport: true,
		Error:     errors.New("invalid database"),
	}
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.True(t, mockImporter.ImportStarted)
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

func TestInitializeImporterSuccessful(t *testing.T) {
	mockImporter := mock.MockImporter{
		CanImport:       true,
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
	assert.True(t, mockImporter.Imported)
}
