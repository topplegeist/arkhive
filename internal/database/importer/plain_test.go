package importer_test

import (
	"encoding/hex"
	"testing"

	"arkhive.dev/launcher/internal/database/importer"
	"github.com/stretchr/testify/assert"
)

func TestImportWrongFolder(t *testing.T) {
	i := importer.NewPlainImporter("not_existing_path")
	value, err := i.Import([]byte{})
	assert.Nil(t, value)
	assert.Nil(t, err)
}

func TestImportSameImportedHash(t *testing.T) {
	currentHash := make([]byte, 20)
	hex.Decode(currentHash, []byte("fa9ada43da1797362c5f3e3ec1f8a5cbbf9a8a34"))
	i := importer.NewPlainImporter("../../../test/invalid_database")
	value, err := i.Import(currentHash)
	assert.Nil(t, value)
	assert.Nil(t, err)
}
