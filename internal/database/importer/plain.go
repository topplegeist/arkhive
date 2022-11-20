package importer

import (
	"bytes"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/sirupsen/logrus"
)

const PlainDatabasePath = "db.json"

type Plain struct {
	basePath string
	consoles []Console
	games    []Game
	tools    []Tool
}

func NewPlain(basePath string) *Plain {
	return &Plain{
		basePath: basePath,
		consoles: []Console{},
		games:    []Game{},
		tools:    []Tool{},
	}
}

func (p *Plain) Import(currentDBHash []byte) (importedDBHash []byte, err error) {
	if !p.canLoad() {
		logrus.Debug("The plain database is not present")
		return nil, nil
	}

	var databaseData []byte
	if databaseData, importedDBHash, err = p.load(currentDBHash); err != nil {
		return
	}

	if err = p.decode(databaseData); err != nil {
		return
	}

	return
}

func (p *Plain) canLoad() bool {
	// Check if a plain database file and the key file exists
	logrus.Debug("Checking if a plain database could be imported")
	_, existenceFlag := os.Stat(filepath.Join(p.basePath, PlainDatabasePath))
	return !os.IsNotExist(existenceFlag)
}

func (p *Plain) load(currentDBHash []byte) (databaseData []byte, encryptedDBHash []byte, err error) {
	// Read the database file to be imported
	var plainDatabaseFileReader *os.File
	if plainDatabaseFileReader, err = os.Open(filepath.Join(p.basePath, PlainDatabasePath)); err != nil {
		logrus.Error("Cannot read the plain database file")
		return
	}
	defer plainDatabaseFileReader.Close()
	databaseBuffer := &bytes.Buffer{}
	if _, err = databaseBuffer.ReadFrom(plainDatabaseFileReader); err != nil {
		logrus.Error("Cannot read the plain database")
		return
	}
	databaseData = databaseBuffer.Bytes()

	logrus.Info("Calculating the database hash")
	// Calculate the hash of the new encrypted database
	hashEncoder := sha1.New()
	if _, err = hashEncoder.Write(databaseData); err != nil {
		return
	}
	encryptedDBHash = hashEncoder.Sum(nil)

	return
}

func (p *Plain) decode(databaseData []byte) (err error) {
	decoder := json.NewDecoder(bytes.NewReader(databaseData))
	decoder.UseNumber()
	var database map[string]interface{}
	if err = decoder.Decode(&database); err != nil {
		logrus.Errorf("%+v", err)
		return
	}

	if jsonEntities, ok := database["consoles"]; ok {
		if jsonEntitiesMap, ok := jsonEntities.(map[string]interface{}); ok {
			for slug, entity := range jsonEntitiesMap {
				var console Console
				if console, err = PlainDatabaseToConsole(slug, entity); err != nil {
					return
				}
				p.consoles = append(p.consoles, console)
			}
		} else {
			err = errors.New("Console field is not an array")
			return
		}
	} else {
		logrus.Warn("No consoles parsed during the database import")
	}

	if jsonEntities, ok := database["games"]; ok {
		if jsonEntitiesMap, ok := jsonEntities.(map[string]interface{}); ok {
			for slug, entity := range jsonEntitiesMap {
				var game Game
				if game, err = PlainDatabaseToGame(slug, entity); err != nil {
					return
				}
				p.games = append(p.games, game)
			}
		} else {
			err = errors.New("Game field is not an array")
			return
		}
	} else {
		logrus.Warn("No games parsed during the database import")
	}

	if jsonEntities, ok := database["win_tools"]; ok {
		if jsonEntitiesMap, ok := jsonEntities.(map[string]interface{}); ok {
			for slug, entity := range jsonEntitiesMap {
				var tool Tool
				if tool, err = PlainDatabaseToTool(slug, entity); err != nil {
					return
				}
				p.tools = append(p.tools, tool)
			}
		} else {
			err = errors.New("Tool field is not an array")
			return
		}
	} else {
		logrus.Warn("No tools parsed during the database import")
	}

	return
}

func (p Plain) GetConsoles() (consoles []Console) {
	return p.consoles
}

func (p Plain) GetGames() (games []Game) {
	return p.games
}

func (p Plain) GetTools() (tools []Tool) {
	return p.tools
}
