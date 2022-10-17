package importer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlainDatabaseToConsoleInvalidJson(t *testing.T) {
	if _, err := PlainDatabaseToConsole("consoleSlug", []string{}); err == nil {
		t.Fail()
	} else {
		assert.Error(t, err, "the console JSON is not an object")
	}
}

func TestPlainDatabaseToConsoleIncompleteJson(t *testing.T) {
	if _, err := PlainDatabaseToConsole("consoleSlug", map[string]interface{}{}); err == nil {
		t.Fail()
	}
}

func TestPlainDatabaseToConsoleStrictValuesNoFileTypes(t *testing.T) {
	if _, err := PlainDatabaseToConsole("consoleSlug", map[string]interface{}{
		"name":          "name",
		"core_location": "core_location",
	}); err == nil {
		t.Fail()
	}
}

func TestPlainDatabaseToConsole(t *testing.T) {
	var (
		entity Console
		err    error
	)
	if entity, err = PlainDatabaseToConsole("consoleSlug", map[string]interface{}{
		"is_embedded":   true,
		"name":          "name",
		"single_file":   true,
		"core_location": "core_location",
		"core_config": map[string]interface{}{
			"core_variable": "core_value",
		},
		"win_config": map[string]interface{}{
			"win_variable": "win_value",
		},
		"config": map[string]interface{}{
			"variable": "value",
		},
		"file_types": map[string]interface{}{
			"action0": []interface{}{"file_type0"},
			"action1": []interface{}{"file_type1"},
		},
		"language": map[string]interface{}{
			"mapping": map[string]interface{}{
				"0": []interface{}{"language0"},
				"1": []interface{}{"language1"},
			},
			"variable_name": "variable_name",
		},
		"plugins": map[string]interface{}{
			"plugin": map[string]interface{}{
				"collection_path": []interface{}{"collection_path"},
				"destination":     []interface{}{"destination"},
				"files":           []interface{}{"files"},
			},
		},
	}); err != nil {
		t.Log(err)
		t.Fail()
	}
	assert.Equal(t, "name", entity.Name)
	assert.True(t, entity.IsEmbedded)
	assert.True(t, entity.SingleFile)
	assert.Equal(t, "core_location", entity.CoreLocation)
	assert.Len(t, entity.Configs, 3)
	configNames := make([]string, len(entity.Configs))
	configLevels := make([]string, len(entity.Configs))
	configValues := make([]string, len(entity.Configs))
	for index, config := range entity.Configs {
		configNames[index] = config.Name
		configLevels[index] = config.Level
		configValues[index] = config.Value
	}
	assert.ElementsMatch(t, configNames, []string{"core_variable", "win_variable", "variable"})
	assert.ElementsMatch(t, configLevels, []string{"core_config", "win_config", "config"})
	assert.ElementsMatch(t, configValues, []string{"core_value", "win_value", "value"})

	assert.Len(t, entity.FileTypes, 2)
	fileTypeAction := make([]string, len(entity.FileTypes))
	fileTypeFileType := make([]string, len(entity.FileTypes))
	for index, fileType := range entity.FileTypes {
		fileTypeAction[index] = fileType.Action
		fileTypeFileType[index] = fileType.FileType
	}
	assert.ElementsMatch(t, fileTypeAction, []string{"action0", "action1"})
	assert.ElementsMatch(t, fileTypeFileType, []string{"file_type0", "file_type1"})

	assert.Len(t, entity.Languages, 2)
	languageNames := make([]string, len(entity.Languages))
	languageTags := make([]uint, len(entity.Languages))
	for index, language := range entity.Languages {
		languageNames[index] = language.Name
		languageTags[index] = language.Tag
	}
	assert.ElementsMatch(t, languageNames, []string{"language0", "language1"})
	assert.ElementsMatch(t, languageTags, []uint{0, 1})

	assert.NotNil(t, entity.LanguageVariableName)
	assert.Equal(t, "variable_name", *entity.LanguageVariableName)
	assert.Len(t, entity.Plugins, 1)
}
