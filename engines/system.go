package engines

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"io/fs"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"arkhive.dev/launcher/common"
	"arkhive.dev/launcher/models"
	"arkhive.dev/launcher/models/network"
	"github.com/BurntSushi/toml"
	log "github.com/sirupsen/logrus"
)

type ConsoleEntryDownload struct {
	ConsoleEntry *models.Console
	URL          url.URL
}

type SystemEngine struct {
	settings              map[string]interface{}
	databaseEngine        *DatabaseEngine
	networkEngine         *NetworkEngine
	preparingConsoleList  []ConsoleEntryDownload
	preparingToolsList    []models.Tool
	preparingPluginsList  []models.ConsolePlugin
	downloadingCoresCount int
	downloadingToolsCount int
	remainingCoresCount   int
	remainingToolsCount   int
	extractingExtensions  []string

	ToolsPreparedEventEmitter                *common.EventEmitter
	DownloadingCoresCountChangedEventEmitter *common.EventEmitter
	DownloadingToolsCountChangedEventEmitter *common.EventEmitter
	ToolElaborationCompletedEventEmitter     *common.EventEmitter
	DownloadingToolChangedEventEmitter       *common.EventEmitter
	CoreElaborationCompletedEventEmitter     *common.EventEmitter
	PluginElaborationCompletedEventEmitter   *common.EventEmitter
	PluginsElaborationCompletedEventEmitter  *common.EventEmitter
	CoresPreparedEventEmitter                *common.EventEmitter
	DownloadingCoreChangedEventEmitter       *common.EventEmitter
}

func NewSystemEngine(databaseEngine *DatabaseEngine, networkEngine *NetworkEngine) (instance *SystemEngine, err error) {
	instance = &SystemEngine{
		databaseEngine:                           databaseEngine,
		networkEngine:                            networkEngine,
		extractingExtensions:                     []string{"zip", "rar", "7z"},
		ToolsPreparedEventEmitter:                &common.EventEmitter{},
		DownloadingCoresCountChangedEventEmitter: &common.EventEmitter{},
		DownloadingToolsCountChangedEventEmitter: &common.EventEmitter{},
		ToolElaborationCompletedEventEmitter:     &common.EventEmitter{},
		DownloadingToolChangedEventEmitter:       &common.EventEmitter{},
		CoreElaborationCompletedEventEmitter:     &common.EventEmitter{},
		PluginElaborationCompletedEventEmitter:   &common.EventEmitter{},
		PluginsElaborationCompletedEventEmitter:  &common.EventEmitter{},
		CoresPreparedEventEmitter:                &common.EventEmitter{},
		DownloadingCoreChangedEventEmitter:       &common.EventEmitter{},
	}
	databaseEngine.DecryptedEventEmitter.Subscribe(instance.startEngine)
	return
}

func (systemEngine *SystemEngine) startEngine(_ bool) error {
	systemEngine.settings = make(map[string]interface{})
	systemEngine.syncSettings()
	if _, err := os.Stat(common.SYSTEM_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.SYSTEM_FOLDER_PATH, 0644)
	}
	if _, err := os.Stat(common.SYSTEM_CORE_PATH); os.IsNotExist(err) {
		os.Mkdir(common.SYSTEM_CORE_PATH, 0644)
	}
	if _, err := os.Stat(common.TOOLS_PATH); os.IsNotExist(err) {
		os.Mkdir(common.SYSTEM_CORE_PATH, 0644)
	}
	if _, err := os.Stat(common.TEMP_DOWNLOAD_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.TEMP_DOWNLOAD_FOLDER_PATH, 0644)
	}

	if _, err := os.Stat(GetDefaultConfigPath()); os.IsNotExist(err) {
		systemEngine.setDefaultConfiguration()
	}
	systemEngine.setFixedConfiguration()

	systemEngine.ToolsPreparedEventEmitter.Subscribe(systemEngine.prepareLaunchers)
	systemEngine.prepareTools()

	return nil
}

func GetUndertow() network.StorjResource {
	return network.StorjResource{
		URL: url.URL{
			Scheme: common.DEFAULT_UNDERTOW_SCHEME,
			Host:   common.DEFAULT_UNDERTOW_HOST,
			Path:   common.DEFAULT_UNDERTOW_PATH,
		},
		Access: common.DEFAULT_UNDERTOW_ACCESS,
	}
}

func GetDefaultConfigPath() string {
	return filepath.Join(common.SYSTEM_FOLDER_PATH, "system.cfg")
}

func (systemEngine *SystemEngine) prepareLaunchers(_ bool) {
	requestURL := url.URL{
		Scheme: common.RETROARCH_BUILDBOT_URL_SCHEME,
		Host:   common.RETROARCH_BUILDBOT_URL_HOST,
		Path:   common.RETROARCH_UPDATE_URL_PATH,
	}

	requestObject := make(map[string]interface{})
	requestObject["action"] = "get"
	requestObject["items"] = make(map[string]interface{})
	requestObject["items"].(map[string]interface{})["href"] = requestURL.Path
	requestObject["items"].(map[string]interface{})["what"] = 1

	var (
		requestObjectBytes []byte
		err                error
	)
	if requestObjectBytes, err = json.Marshal(requestObject); err != nil {
		return
	}
	go func() {
		var (
			response *http.Response
			err      error
		)
		if response, err = http.Post(requestURL.String(), "application/json", bytes.NewReader(requestObjectBytes)); err != nil {
			log.Error("Buildbot request failed")
			log.Error(err)
			return
		}
		systemEngine.collectRetroArchCoresInfoFinished(response.Body)
	}()
}

func (systemEngine *SystemEngine) prepareTools() (err error) {
	var tools []models.Tool
	if tools, err = systemEngine.databaseEngine.GetTools(); err != nil {
		log.Error("Cannot get tools from database")
		log.Error(err)
		return
	}
	for _, toolEntry := range tools {
		if !systemEngine.toolIsDownloaded(&toolEntry) || !systemEngine.toolIsUpdated(&toolEntry) {
			systemEngine.preparingToolsList = append(
				systemEngine.preparingToolsList, toolEntry)
		}
	}
	systemEngine.downloadingToolsCount = len(systemEngine.preparingToolsList)
	systemEngine.DownloadingToolsCountChangedEventEmitter.Emit(true)
	systemEngine.ToolElaborationCompletedEventEmitter.Subscribe(systemEngine.prepareNextTool)
	systemEngine.prepareNextTool(true)
	return
}

func (systemEngine *SystemEngine) collectRetroArchCoresInfoFinished(reader io.Reader) {
	buffer := new(bytes.Buffer)
	if _, err := buffer.ReadFrom(reader); err != nil {
		log.Error("Buildbot request failed")
		log.Error(err)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(buffer.Bytes()))
	decoder.UseNumber()
	remoteInfo := make(map[string]interface{})
	if err := decoder.Decode(&remoteInfo); err != nil {
		log.Error("Buildbot JSON parsing error")
		log.Fatal(err)
		return
	}

	var (
		consoles []models.Console
		err      error
	)
	if consoles, err = systemEngine.databaseEngine.GetConsoles(); err != nil {
		log.Error("Cannot get consoles from database")
		log.Error(err)
		return
	}
	for consoleEntryIndex, consoleEntry := range consoles {
		if !systemEngine.coreIsDownloaded(&consoleEntry) || !systemEngine.coreIsUpdated(&consoleEntry) {
			for _, item := range remoteInfo["items"].([]interface{}) {
				href := item.(map[string]interface{})["href"].(string)
				suffix := consoleEntry.CoreLocation + "." + common.CORES_EXTENSION + ".zip"
				if strings.HasSuffix(href, suffix) {
					systemEngine.preparingConsoleList = append(
						systemEngine.preparingConsoleList,
						ConsoleEntryDownload{
							ConsoleEntry: &consoles[consoleEntryIndex],
							URL: url.URL{
								Scheme: common.RETROARCH_BUILDBOT_URL_SCHEME,
								Host:   common.RETROARCH_BUILDBOT_URL_HOST,
								Path:   href,
							},
						})
					break
				}
			}
		}
	}

	systemEngine.downloadingCoresCount = len(systemEngine.preparingConsoleList)
	systemEngine.DownloadingCoresCountChangedEventEmitter.Emit(true)
	systemEngine.CoreElaborationCompletedEventEmitter.Subscribe(func(_ bool) {
		systemEngine.getPlugins(systemEngine.preparingConsoleList[0].ConsoleEntry)
	})
	systemEngine.PluginElaborationCompletedEventEmitter.Subscribe(systemEngine.prepareNextPlugin)
	systemEngine.PluginsElaborationCompletedEventEmitter.Subscribe(systemEngine.prepareNextCore)
	systemEngine.prepareNextCore(true)
}

func (systemEngine *SystemEngine) prepareNextCore(first bool) {
	if len(systemEngine.preparingConsoleList) > 0 {
		if !first {
			systemEngine.preparingConsoleList = systemEngine.preparingConsoleList[1:]
		}
		systemEngine.remainingCoresCount = len(systemEngine.preparingConsoleList)
		systemEngine.DownloadingCoreChangedEventEmitter.Emit(true)
		if systemEngine.remainingCoresCount > 0 {
			systemEngine.getCore(&systemEngine.preparingConsoleList[0])
			return
		}
	}
	systemEngine.CoresPreparedEventEmitter.Emit(true)
}

func (systemEngine *SystemEngine) prepareNextPlugin(first bool) {
	if len(systemEngine.preparingPluginsList) > 0 {
		if !first {
			systemEngine.preparingPluginsList = systemEngine.preparingPluginsList[1:]
		}
		if len(systemEngine.preparingPluginsList) > 0 {
			systemEngine.getPlugin(&systemEngine.preparingPluginsList[0])
			return
		}
	}
	systemEngine.PluginsElaborationCompletedEventEmitter.Emit(false)
}

func (systemEngine *SystemEngine) getCore(consoleEntryDownload *ConsoleEntryDownload) {
	if !systemEngine.coreIsDownloaded(consoleEntryDownload.ConsoleEntry) ||
		!systemEngine.coreIsUpdated(consoleEntryDownload.ConsoleEntry) {
		var (
			err error
		)
		var resource *network.Resource
		if resource, err = systemEngine.networkEngine.addResource(&consoleEntryDownload.URL, common.TEMP_DOWNLOAD_FOLDER_PATH); err != nil {
			log.Error("Cannot add the download resource to the network engine")
			log.Error(err)
			return
		}
		resource.AvailableEventEmitter.Subscribe(func(_ *network.Resource) {
			systemEngine.saveCoreFile(consoleEntryDownload.ConsoleEntry)
		})
	} else {
		systemEngine.CoreElaborationCompletedEventEmitter.Emit(true)
	}
}

func (systemEngine *SystemEngine) getPlugins(consoleEntry *models.Console) {
	var err error
	if systemEngine.preparingPluginsList, err = systemEngine.databaseEngine.GetConsolePluginsByConsole(consoleEntry); err != nil {
		log.Error("Cannot get console plugins from database")
		log.Error(err)
		return
	}
	systemEngine.prepareNextPlugin(true)
}

func (systemEngine *SystemEngine) getPlugin(consolePlugin *models.ConsolePlugin) {
	if consolePlugin.Type == "bios" {
		var (
			err                error
			consolePluginFiles []models.ConsolePluginsFile
		)
		if consolePluginFiles, err = systemEngine.databaseEngine.GetConsolePluginsFilesByConsolePlugin(consolePlugin); err != nil {
			log.Error("Cannot get console plugins files from database")
			log.Error(err)
			return
		}
		if len(consolePluginFiles) > 0 {
			var consolePluginFileUrl *url.URL
			if consolePluginFileUrl, err = url.Parse(consolePluginFiles[0].Url); err != nil {
				log.Error("Cannot parse tool URL")
				log.Error(err)
				return
			}
			var resource *network.Resource
			if resource, err = systemEngine.networkEngine.addResource(
				consolePluginFileUrl,
				path.Dir(
					GetDownloadCorePluginPath(consolePlugin, &consolePluginFiles[0]))); err != nil {
				log.Error("Cannot add the download resource to the network engine")
				log.Error(err)
				return
			}
			resource.AvailableEventEmitter.Subscribe(func(_ *network.Resource) {
				systemEngine.savePluginFile(consolePlugin)
			})
		} else {
			var console models.Console
			if console, err = systemEngine.databaseEngine.GetConsoleByConsolePlugin(consolePlugin); err != nil {
				return
			}
			log.Warn("No files for console plugin in ", console.Slug, " console")
			systemEngine.PluginElaborationCompletedEventEmitter.Emit(false)
		}
	} else {
		panic("console plugin type not handled")
	}
}

func (systemEngine *SystemEngine) getTool(toolEntry *models.Tool) {
	var (
		toolUrl *url.URL
		err     error
	)
	if toolUrl, err = url.Parse(toolEntry.Url); err != nil {
		log.Error("Cannot parse tool URL")
		log.Error(err)
		return
	}
	var resource *network.Resource
	if resource, err = systemEngine.networkEngine.addResource(toolUrl, common.TEMP_DOWNLOAD_FOLDER_PATH); err != nil {
		log.Error("Cannot add the download resource to the network engine")
		log.Error(err)
		return
	}
	resource.AvailableEventEmitter.Subscribe(func(_ *network.Resource) {
		systemEngine.saveToolFile(toolEntry)
	})
}

func (systemEngine *SystemEngine) prepareNextTool(first bool) {
	if len(systemEngine.preparingToolsList) > 0 {
		if !first {
			systemEngine.preparingToolsList = systemEngine.preparingToolsList[1:]
		}
		systemEngine.remainingToolsCount = len(systemEngine.preparingToolsList)
		systemEngine.DownloadingToolChangedEventEmitter.Emit(true)
		if systemEngine.remainingToolsCount > 0 {
			systemEngine.getTool(&systemEngine.preparingToolsList[0])
			return
		}
	}
	systemEngine.ToolsPreparedEventEmitter.Emit(true)
}

func (systemEngine *SystemEngine) saveCoreFile(consoleEntry *models.Console) {
	log.Info("Core ", consoleEntry.Slug, " downloaded")
	if err := systemEngine.extractCoreArchive(consoleEntry); err != nil {
		return
	}
	if err := systemEngine.elaborateCoreArchive(consoleEntry); err != nil {
		return
	}
	log.Info("Core ", consoleEntry.Slug, " completed")
	systemEngine.CoreElaborationCompletedEventEmitter.Emit(true)
}

func (systemEngine *SystemEngine) savePluginFile(consolePlugin *models.ConsolePlugin) {
	var (
		err     error
		console models.Console
	)
	if console, err = systemEngine.databaseEngine.GetConsoleByConsolePlugin(consolePlugin); err != nil {
		return
	}
	log.Info("Console plugins for ", console.Slug, " downloaded")
	var consolePluginFiles []models.ConsolePluginsFile
	if consolePluginFiles, err = systemEngine.databaseEngine.GetConsolePluginsFilesByConsolePlugin(consolePlugin); err != nil {
		return
	}
	if err = systemEngine.extractPluginArchive(consolePlugin, consolePluginFiles); err != nil {
		return
	}
	if err = systemEngine.elaboratePluginArchive(consolePlugin, consolePluginFiles); err != nil {
		return
	}
	log.Info("Console plugins for ", console.Slug, " completed")
	systemEngine.PluginElaborationCompletedEventEmitter.Emit(false)
}

func (systemEngine *SystemEngine) saveToolFile(toolEntry *models.Tool) {
	log.Info("Tool ", toolEntry.Slug, " downloaded")
	if err := systemEngine.extractToolArchive(toolEntry); err != nil {
		return
	}
	if err := systemEngine.elaborateToolArchive(toolEntry); err != nil {
		return
	}
	log.Info("Tool ", toolEntry.Slug, " completed")
	systemEngine.ToolElaborationCompletedEventEmitter.Emit(false)
}

func (systemEngine *SystemEngine) coreIsDownloaded(consoleEntry *models.Console) bool {
	coreLocation := consoleEntry.CoreLocation + "." + common.CORES_EXTENSION
	if _, err := os.Stat(filepath.Join(common.SYSTEM_CORE_PATH, coreLocation)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (systemEngine *SystemEngine) coreIsUpdated(_ *models.Console) bool {
	// ToDo: Check core updates from retroarch buildbot
	return true
}

func (systemEngine *SystemEngine) toolIsDownloaded(toolEntry *models.Tool) bool {
	var toolLocation string
	if toolEntry.Destination.Valid && toolEntry.Destination.String != "" {
		toolLocation = filepath.Join(common.TOOLS_PATH, toolEntry.Destination.String)
	} else if toolEntry.CollectionPath.Valid && toolEntry.CollectionPath.String != "" {
		toolLocation = filepath.Join(common.TOOLS_PATH, filepath.Base(toolEntry.CollectionPath.String))
	} else {
		var (
			toolUrl *url.URL
			err     error
		)
		if toolUrl, err = url.Parse(toolEntry.Url); err != nil {
			return false
		}
		toolLocation = filepath.Join(common.TOOLS_PATH, filepath.Base(toolUrl.Path))
	}
	if _, err := os.Stat(toolLocation); os.IsNotExist(err) {
		return false
	}
	return true
}

func (systemEngine *SystemEngine) toolIsUpdated(_ *models.Tool) bool {
	// ToDo: Check core updates from retroarch buildbot
	return true
}

func (systemEngine *SystemEngine) extractCoreArchive(consoleEntry *models.Console) error {
	process := exec.Command(
		common.SEVENZ_EXE_PATH,
		"x",
		GetDownloadCorePath(consoleEntry),
		"-o"+GetCoreTempPath(consoleEntry))
	if err := process.Run(); err != nil {
		log.Error("Error starting the extraction process")
		log.Error(err)
		return err
	}
	return nil
}

func (systemEngine *SystemEngine) elaborateCoreArchive(consoleEntry *models.Console) (err error) {
	coreTempPath := GetCoreTempPath(consoleEntry)
	filepath.Walk(coreTempPath, func(filePath string, info fs.FileInfo, err error) error {
		if path.Ext(info.Name()) != "" && path.Ext(info.Name())[1:] == common.CORES_EXTENSION {
			os.Rename(filePath, GetCorePath(consoleEntry))
		}
		return nil
	})
	os.Remove(GetDownloadCorePath(consoleEntry))
	os.RemoveAll(coreTempPath)
	return
}

func (systemEngine *SystemEngine) extractPluginArchive(consolePlugin *models.ConsolePlugin, consolePluginsFiles []models.ConsolePluginsFile) error {
	if consolePlugin.Type == "bios" {
		for fileIndex, consolePluginFile := range consolePluginsFiles {
			consolePluginFilePath := GetDownloadCorePluginPath(consolePlugin, &consolePluginFile)
			isExtractingExtension := false
			for _, item := range systemEngine.extractingExtensions {
				if item == path.Ext(consolePluginFilePath)[1:] {
					isExtractingExtension = true
					break
				}
			}
			if !isExtractingExtension {
				return nil
			}
			process := exec.Command(
				common.SEVENZ_EXE_PATH,
				"x",
				GetDownloadCorePluginPath(consolePlugin, &consolePluginFile),
				"-o"+GetCorePluginTempPath(consolePlugin, fileIndex))
			if err := process.Run(); err != nil {
				log.Error("Error starting the extraction process")
				log.Error(err)
				return err
			}
		}
	}
	return nil
}

func (systemEngine *SystemEngine) elaboratePluginArchive(consolePlugin *models.ConsolePlugin, consolePluginsFiles []models.ConsolePluginsFile) (err error) {
	if consolePlugin.Type == "bios" {
		for fileIndex, consolePluginFile := range consolePluginsFiles {
			consolePluginFilePath := GetDownloadCorePluginPath(consolePlugin, &consolePluginFile)
			destinationFolder := ""
			if consolePluginFile.Destination.Valid {
				destinationFolder = consolePluginFile.Destination.String
			}
			if _, err := os.Stat(destinationFolder); os.IsNotExist(err) {
				os.Mkdir(destinationFolder, 0644)
			}
			isExtractingExtension := false
			for _, item := range systemEngine.extractingExtensions {
				if item == path.Ext(consolePluginFilePath)[1:] {
					isExtractingExtension = true
					break
				}
			}
			if !isExtractingExtension {
				destinationPath := path.Join(destinationFolder, path.Base(consolePluginFilePath))
				os.Rename(consolePluginFilePath, destinationPath)
			} else {
				extractionDir := GetCorePluginTempPath(consolePlugin, fileIndex)
				collectionPath := extractionDir
				if consolePluginFile.CollectionPath.Valid {
					collectionPath = path.Join(collectionPath, consolePluginFile.CollectionPath.String)
				}
				var collectionFileInfo fs.FileInfo
				if collectionFileInfo, err = os.Stat(collectionPath); err != nil {
					return
				}
				if collectionFileInfo.IsDir() {
					os.Rename(collectionPath, destinationFolder)
				} else {
					destinationPath := path.Join(destinationFolder, path.Base(collectionPath))
					os.Rename(collectionPath, destinationPath)
				}
			}
		}
		os.RemoveAll(GetPluginTempPath())
	}
	return
}

func (systemEngine *SystemEngine) extractToolArchive(toolEntry *models.Tool) error {
	isExtractingExtension := false
	for _, item := range systemEngine.extractingExtensions {
		if item == path.Ext(toolEntry.Url)[1:] {
			isExtractingExtension = true
			break
		}
	}
	if !isExtractingExtension {
		return nil
	}
	process := exec.Command(
		common.SEVENZ_EXE_PATH,
		"x",
		GetDownloadToolPath(toolEntry),
		"-o"+GetToolTempPath(toolEntry))
	if err := process.Run(); err != nil {
		log.Error("Error starting the extraction process")
		log.Error(err)
		return err
	}
	return nil
}

func (systemEngine *SystemEngine) elaborateToolArchive(toolEntry *models.Tool) (err error) {
	destinationFolder := common.TOOLS_PATH
	if _, err := os.Stat(destinationFolder); os.IsNotExist(err) {
		os.Mkdir(destinationFolder, 0644)
	}
	if toolEntry.Destination.Valid && toolEntry.Destination.String != "" {
		destinationFolder = path.Join(destinationFolder, toolEntry.Destination.String)
	}
	isExtractingExtension := false
	for _, item := range systemEngine.extractingExtensions {
		if item == path.Ext(toolEntry.Url)[1:] {
			isExtractingExtension = true
			break
		}
	}
	if !isExtractingExtension {
		destinationPath := path.Join(destinationFolder, path.Base(toolEntry.Url))
		os.Rename(GetDownloadToolPath(toolEntry), destinationPath)
	} else {
		extractionDir := GetToolTempPath(toolEntry)
		collectionPath := extractionDir
		if toolEntry.CollectionPath.Valid && toolEntry.CollectionPath.String != "" {
			collectionPath = path.Join(collectionPath, toolEntry.CollectionPath.String)
		}
		var collectionFileInfo fs.FileInfo
		collectionFileInfo, err = os.Stat(collectionPath)
		if collectionFileInfo.IsDir() {
			os.Rename(collectionPath, destinationFolder)
		} else {
			destinationPath := path.Join(destinationFolder, collectionFileInfo.Name())
			os.Rename(collectionPath, destinationPath)
		}
		os.RemoveAll(extractionDir)
	}
	os.Remove(GetDownloadToolPath(toolEntry))
	return
}

func GetDownloadCorePath(consoleEntry *models.Console) string {
	fileName := consoleEntry.CoreLocation + "." + common.CORES_EXTENSION + ".zip"
	return path.Join(common.TEMP_DOWNLOAD_FOLDER_PATH, fileName)
}

func GetDownloadCorePluginPath(consolePlugin *models.ConsolePlugin, consolePluginFile *models.ConsolePluginsFile) string {
	tempDownloadDir := GetPluginTempPath()
	return path.Join(tempDownloadDir, GetDownloadCorePluginFileName(consolePlugin, consolePluginFile))
}

func GetDownloadCorePluginFileName(consolePlugin *models.ConsolePlugin, consolePluginFile *models.ConsolePluginsFile) string {
	if consolePlugin.Type == "bios" {
		if url, err := url.Parse(consolePluginFile.Url); err == nil {
			if url.Fragment != "" {
				return url.Fragment
			} else {
				return path.Base(url.Path)
			}
		}
	}
	panic("plugin file name unavailable")
}

func GetDownloadToolPath(toolEntry *models.Tool) (toolPath string) {
	toolPath = common.TEMP_DOWNLOAD_FOLDER_PATH
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		os.Mkdir(toolPath, 0644)
	}
	url, _ := url.Parse(toolEntry.Url)
	toolPath = path.Join(toolPath, path.Base(url.Path))
	return
}

func GetCoreTempPath(consoleEntry *models.Console) (tempDownloadDir string) {
	tempDownloadDir = common.TEMP_DOWNLOAD_FOLDER_PATH
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0644)
	}
	tempDownloadDir = path.Join(tempDownloadDir, consoleEntry.Slug)
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0644)
	}
	return
}

func GetCorePluginTempPath(consolePlugin *models.ConsolePlugin, fileIndex int) string {
	tempDownloadDir := GetPluginTempPath()
	return path.Join(tempDownloadDir, strconv.Itoa(fileIndex))
}

func GetToolTempPath(toolEntry *models.Tool) (tempDownloadDir string) {
	tempDownloadDir = path.Join(common.TEMP_DOWNLOAD_FOLDER_PATH, toolEntry.Slug)
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0644)
	}
	return
}

func GetPluginTempPath() string {
	tempDownloadDir := common.TEMP_DOWNLOAD_FOLDER_PATH
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0644)
	}
	tempDownloadDir = path.Join(tempDownloadDir, common.TEMP_PLUGIN_FOLDER_NAME)
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0644)
	}
	return tempDownloadDir
}

func GetCorePath(consoleEntry *models.Console) string {
	return path.Join(
		common.SYSTEM_CORE_PATH,
		consoleEntry.CoreLocation+"."+common.CORES_EXTENSION)
}

func (systemEngine *SystemEngine) setDefaultConfiguration() {
	systemEngine.settings["input_player1_l"] = "q"
	systemEngine.settings["input_player1_l2"] = "num1"
	systemEngine.settings["input_player1_l3"] = "nul"
	systemEngine.settings["input_player1_r"] = "e"
	systemEngine.settings["input_player1_r2"] = "num3"
	systemEngine.settings["input_player1_r3"] = "nul"

	systemEngine.settings["input_player1_select"] = "z"
	systemEngine.settings["input_player1_start"] = "x"

	systemEngine.settings["input_player1_up"] = "up"
	systemEngine.settings["input_player1_left"] = "left"
	systemEngine.settings["input_player1_down"] = "down"
	systemEngine.settings["input_player1_right"] = "right"

	systemEngine.settings["input_player1_x"] = "w"
	systemEngine.settings["input_player1_y"] = "a"
	systemEngine.settings["input_player1_b"] = "s"
	systemEngine.settings["input_player1_a"] = "d"
}

func (systemEngine *SystemEngine) setFixedConfiguration() (err error) {
	var systemFolder string
	if systemFolder, err = filepath.Abs(common.SYSTEM_FOLDER_PATH); err != nil {
		log.Error("Cannot get absolute shader folder")
		log.Error(err)
		return
	}

	shadersFolderPath := filepath.Join(systemFolder, "shaders")
	systemEngine.settings["system_directory"] = systemFolder
	systemEngine.settings["global_core_options"] = true
	systemEngine.settings["video_shader_dir"] = shadersFolderPath
	systemEngine.settings["video_windowed_fullscreen"] = true
	systemEngine.settings["input_audio_mute"] = "nul"
	systemEngine.settings["input_cheat_index_minus"] = "nul"
	systemEngine.settings["input_cheat_index_plus"] = "nul"
	systemEngine.settings["input_cheat_toggle"] = "nul"
	systemEngine.settings["input_desktop_menu_toggle"] = "nul"
	systemEngine.settings["input_fps_toggle"] = "nul"
	systemEngine.settings["input_frame_advance"] = "nul"
	systemEngine.settings["input_grab_mouse_toggle"] = "nul"
	systemEngine.settings["input_hold_fast_forward"] = "nul"
	systemEngine.settings["input_hold_slowmotion"] = "nul"
	systemEngine.settings["input_load_state"] = "nul"
	systemEngine.settings["input_menu_toggle"] = "nul"
	systemEngine.settings["input_movie_record_toggle"] = "nul"
	systemEngine.settings["input_netplay_game_watch"] = "nul"
	systemEngine.settings["input_osk_toggle"] = "nul"
	systemEngine.settings["input_pause_toggle"] = "nul"
	systemEngine.settings["input_reset"] = "nul"
	systemEngine.settings["input_rewind"] = "nul"
	systemEngine.settings["input_save_state"] = "nul"
	systemEngine.settings["input_screenshot"] = "nul"
	systemEngine.settings["input_send_debug_info"] = "nul"
	systemEngine.settings["input_shader_next"] = "nul"
	systemEngine.settings["input_shader_prev"] = "nul"
	systemEngine.settings["input_state_slot_decrease"] = "nul"
	systemEngine.settings["input_state_slot_increase"] = "nul"
	systemEngine.settings["input_toggle_fast_forward"] = "nul"
	systemEngine.settings["input_toggle_fullscreen"] = "nul"
	systemEngine.settings["input_volume_down"] = "nul"
	systemEngine.settings["input_volume_up"] = "nul"
	if runtime.GOOS == "windows" {
		systemEngine.settings["video_driver"] = "gl"
		systemEngine.settings["input_joypad_driver"] = "xinput"
	}
	systemEngine.settings["menu_enable_widgets"] = false
	systemEngine.settings["video_shader_enable"] = false

	// Core defaults
	systemEngine.settings["input_libretro_device_p1"] = "1"
	systemEngine.settings["input_libretro_device_p2"] = "1"
	systemEngine.settings["input_libretro_device_p3"] = "1"
	systemEngine.settings["input_libretro_device_p4"] = "1"
	systemEngine.settings["aspect_ratio_index"] = 22
	systemEngine.settings["video_rotation"] = 0
	systemEngine.settings["video_scale_integer"] = false

	// Language handling
	var databaseLanguage Locale
	databaseLanguage, _ = systemEngine.databaseEngine.GetLanguage()
	language := systemEngine.languageToRetroArchIndex(databaseLanguage)
	systemEngine.settings["user_language"] = language
	systemEngine.syncSettings()

	return
}

func (systemEngine *SystemEngine) languageToRetroArchIndex(databaseLanguage Locale) int {
	switch databaseLanguage {
	case FRENCH:
		return 2
	case SPANISH:
		return 3
	case GERMAN:
		return 4
	case ITALIAN:
		return 5
	}
	return 0
}

func (systemEngine *SystemEngine) syncSettings() (err error) {
	savedSettingsMap := make(map[string]interface{})
	configFilePath := GetDefaultConfigPath()
	if _, err = os.Stat(configFilePath); !os.IsNotExist(err) {
		var configFileData []byte
		if configFileData, err = os.ReadFile(configFilePath); err != nil {
			return
		}
		if err = toml.Unmarshal(configFileData, &savedSettingsMap); err != nil {
			return
		}
	}
	for settingKey, settingValue := range savedSettingsMap {
		if _, ok := systemEngine.settings[settingKey]; !ok {
			systemEngine.settings[settingKey] = settingValue
		}
	}
	var file *os.File
	if file, err = os.OpenFile(configFilePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644); err != nil {
		return
	}
	return toml.NewEncoder(bufio.NewWriter(file)).Encode(systemEngine.settings)
}
