package system

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
	"sync"

	"arkhive.dev/launcher/internal/buildbot"
	"arkhive.dev/launcher/internal/database"
	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/internal/network"
	"arkhive.dev/launcher/internal/network/resources"
	"arkhive.dev/launcher/internal/osconstants"
	"arkhive.dev/launcher/internal/undertow"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
)

type ConsoleEntryDownload struct {
	ConsoleEntry *console.Console
	URL          url.URL
}

type SystemEngine struct {
	databaseEngine       database.DatabaseEngine
	networkEngine        network.NetworkEngine
	settings             map[string]interface{}
	preparingConsoleList []ConsoleEntryDownload
	preparingToolsList   []Tool
	preparingPluginsList []ConsolePlugin
	extractingExtensions []string
}

func NewSystemEngine() (instance *SystemEngine, err error) {
	instance = &SystemEngine{
		extractingExtensions: []string{"zip", "rar", "7z"},
	}
	return
}

func (systemEngine *SystemEngine) Initialize(waitGroup *sync.WaitGroup) {
	if _, err := os.Stat(folder.SYSTEM); os.IsNotExist(err) {
		if err = os.Mkdir(folder.SYSTEM, 0755); err != nil {
			panic(err)
		}
	}
	if _, err := os.Stat(folder.CORES); os.IsNotExist(err) {
		if err = os.Mkdir(folder.CORES, 0755); err != nil {
			panic(err)
		}
	}
	if _, err := os.Stat(folder.TOOLS); os.IsNotExist(err) {
		if err = os.Mkdir(folder.CORES, 0755); err != nil {
			panic(err)
		}
	}
	if _, err := os.Stat(folder.TEMP); os.IsNotExist(err) {
		if err = os.Mkdir(folder.TEMP, 0755); err != nil {
			panic(err)
		}
	}
}

func (systemEngine *SystemEngine) startEngine(_ bool) (err error) {
	systemEngine.settings = make(map[string]interface{})
	systemEngine.syncSettings()

	if _, err := os.Stat(GetDefaultConfigPath()); os.IsNotExist(err) {
		systemEngine.setDefaultConfiguration()
	}
	systemEngine.setFixedConfiguration()

	//systemEngine.ToolsPreparedEventEmitter.Subscribe(systemEngine.prepareLaunchers)
	err = systemEngine.prepareTools()

	return err
}

func GetDefaultConfigPath() string {
	return filepath.Join(folder.SYSTEM, "system.cfg")
}

func (systemEngine *SystemEngine) prepareLaunchers(_ bool) {
	requestURL := url.URL{
		Scheme: buildbot.BUILDBOT_URL_SCHEME,
		Host:   buildbot.BUILDBOT_URL_HOST,
		Path:   buildbot.BUILDBOT_UPDATE_URL_PATH,
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
			logrus.Error("Buildbot request failed")
			logrus.Errorf("%+v", err)
			return
		}
		systemEngine.collectRetroArchCoresInfoFinished(response.Body)
	}()
}

func (systemEngine *SystemEngine) prepareTools() (err error) {
	var tools []tool.Tool
	if tools, err = systemEngine.databaseEngine.GetTools(); err != nil {
		logrus.Error("Cannot get tools from database")
		logrus.Errorf("%+v", err)
		return
	}
	for _, toolEntry := range tools {
		if !systemEngine.toolIsDownloaded(&toolEntry) || !systemEngine.toolIsUpdated(&toolEntry) {
			systemEngine.preparingToolsList = append(
				systemEngine.preparingToolsList, toolEntry)
		}
	}
	//systemEngine.ToolElaborationCompletedEventEmitter.Subscribe(systemEngine.prepareNextTool)
	systemEngine.prepareNextTool(true)
	return
}

func (systemEngine *SystemEngine) collectRetroArchCoresInfoFinished(reader io.Reader) {
	buffer := &bytes.Buffer{}
	if _, err := buffer.ReadFrom(reader); err != nil {
		logrus.Error("Buildbot request failed")
		logrus.Errorf("%+v", err)
		return
	}

	decoder := json.NewDecoder(bytes.NewReader(buffer.Bytes()))
	decoder.UseNumber()
	remoteInfo := make(map[string]interface{})
	if err := decoder.Decode(&remoteInfo); err != nil {
		logrus.Error("Buildbot JSON parsing error")
		logrus.Errorf("%+v", err)
		return
	}

	var (
		consoles []console.Console
		err      error
	)
	if consoles, err = systemEngine.databaseEngine.GetConsoles(); err != nil {
		logrus.Error("Cannot get consoles from database")
		logrus.Errorf("%+v", err)
		return
	}
	for consoleEntryIndex, consoleEntry := range consoles {
		if !systemEngine.coreIsDownloaded(&consoleEntry) || !systemEngine.coreIsUpdated(&consoleEntry) {
			for _, item := range remoteInfo["items"].([]interface{}) {
				href := item.(map[string]interface{})["href"].(string)
				suffix := consoleEntry.CoreLocation + "." + osconstants.CORES_EXTENSION + ".zip"
				if strings.HasSuffix(href, suffix) {
					systemEngine.preparingConsoleList = append(
						systemEngine.preparingConsoleList,
						ConsoleEntryDownload{
							ConsoleEntry: &consoles[consoleEntryIndex],
							URL: url.URL{
								Scheme: buildbot.BUILDBOT_URL_SCHEME,
								Host:   buildbot.BUILDBOT_URL_HOST,
								Path:   href,
							},
						})
					break
				}
			}
		}
	}

	//systemEngine.CoreElaborationCompletedEventEmitter.Subscribe(func(_ bool) {
	//	systemEngine.getPlugins(systemEngine.preparingConsoleList[0].ConsoleEntry)
	//})
	//systemEngine.PluginElaborationCompletedEventEmitter.Subscribe(systemEngine.prepareNextPlugin)
	//systemEngine.PluginsElaborationCompletedEventEmitter.Subscribe(systemEngine.prepareNextCore)
	systemEngine.prepareNextCore(true)
}

func (systemEngine *SystemEngine) prepareNextCore(first bool) {
	if len(systemEngine.preparingConsoleList) > 0 {
		if !first {
			systemEngine.preparingConsoleList = systemEngine.preparingConsoleList[1:]
		}
		if len(systemEngine.preparingConsoleList) > 0 {
			systemEngine.getCore(&systemEngine.preparingConsoleList[0])
			return
		}
	} else {
		//systemEngine.BootedEventEmitter.Emit(true)
	}
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
	//systemEngine.PluginsElaborationCompletedEventEmitter.Emit(false)
}

func (systemEngine *SystemEngine) getCore(consoleEntryDownload *ConsoleEntryDownload) {
	if !systemEngine.coreIsDownloaded(consoleEntryDownload.ConsoleEntry) ||
		!systemEngine.coreIsUpdated(consoleEntryDownload.ConsoleEntry) {
		//var (
		//	err error
		//)
		//var resource *resources.Resource
		//if resource, err = systemEngine.networkEngine.AddResource(&consoleEntryDownload.URL, folder.TEMP); err != nil {
		//	logrus.Error("Cannot add the download resource to the network engine")
		//	logrus.Errorf("%+v", err)
		//	return
		//}
		//resource.AvailableEventEmitter.Subscribe(func(_ *resources.Resource) {
		//	systemEngine.saveCoreFile(consoleEntryDownload.ConsoleEntry)
		//})
		//resource.ProgressUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
		//	url := resource.Handler.GetURL()
		//	logrus.Debugf("%s: Core download progress %d/%d (%d%%)", url.String(), resource.Available, resource.Total, resource.Available*100/resource.Total)
		//})
	} else {
		//systemEngine.CoreElaborationCompletedEventEmitter.Emit(true)
	}
}

func (systemEngine *SystemEngine) getPlugins(consoleEntry *console.Console) {
	var err error
	if systemEngine.preparingPluginsList, err = systemEngine.databaseEngine.GetConsolePluginsByConsole(consoleEntry); err != nil {
		logrus.Error("Cannot get console plugins from database")
		logrus.Errorf("%+v", err)
		return
	}
	systemEngine.prepareNextPlugin(true)
}

func (systemEngine *SystemEngine) getPlugin(consolePlugin *console.ConsolePlugin) {
	if consolePlugin.Type == "bios" {
		var (
			err                 error
			consolePluginsFiles []console.ConsolePluginsFile
		)
		if consolePluginsFiles, err = systemEngine.databaseEngine.GetConsolePluginsFilesByConsolePlugin(consolePlugin); err != nil {
			logrus.Error("Cannot get console plugins files from database")
			logrus.Errorf("%+v", err)
			return
		}
		//for consolePluginsFileIndex, consolePluginsFile := range consolePluginsFiles {
		//var consolePluginFileUrl *url.URL
		//if consolePluginFileUrl, err = url.Parse(consolePluginsFile.Url); err != nil {
		//	logrus.Error("Cannot parse tool URL")
		//	logrus.Errorf("%+v", err)
		//	return
		//}
		//var resource *resources.Resource
		//if resource, err = systemEngine.networkEngine.AddResource(
		//	consolePluginFileUrl,
		//	path.Dir(
		//		GetDownloadCorePluginPath(consolePlugin, &consolePluginsFile))); err != nil {
		//	logrus.Error("Cannot add the download resource to the network engine")
		//	logrus.Errorf("%+v", err)
		//	return
		//}
		//currentConsolePluginFileIndex := consolePluginsFileIndex
		//resource.AvailableEventEmitter.Subscribe(func(_ *resources.Resource) {
		//	systemEngine.savePluginFile(consolePlugin, &consolePluginsFiles[currentConsolePluginFileIndex], currentConsolePluginFileIndex)
		//})
		//resource.ProgressUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
		//	url := resource.Handler.GetURL()
		//	logrus.Debugf("%s: Console plugin file download progress %d/%d (%d%%)", url.String(), resource.Available, resource.Total, resource.Available*100/resource.Total)
		//})
		//}
		if len(consolePluginsFiles) == 0 {
			var console console.Console
			if console, err = systemEngine.databaseEngine.GetConsoleByConsolePlugin(consolePlugin); err != nil {
				return
			}
			logrus.Warnf("No files for console plugin in %s console", console.Slug)
			//systemEngine.PluginElaborationCompletedEventEmitter.Emit(false)
		}
	} else {
		panic("console plugin type not handled")
	}
}

func (systemEngine *SystemEngine) getTool(toolEntry *tool.Tool) {
	//var (
	//	toolUrl *url.URL
	//	err     error
	//)
	//if toolUrl, err = url.Parse(toolEntry.Url); err != nil {
	//	logrus.Error("Cannot parse tool URL")
	//	logrus.Errorf("%+v", err)
	//	return
	//}
	//var resource *resources.Resource
	//if resource, err = systemEngine.networkEngine.AddResource(toolUrl, folder.TEMP); err != nil {
	//	logrus.Error("Cannot add the download resource to the network engine")
	//	logrus.Errorf("%+v", err)
	//	return
	//}
	//resource.AvailableEventEmitter.Subscribe(func(_ *resources.Resource) {
	//	systemEngine.saveToolFile(toolEntry)
	//})
	//resource.ProgressUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
	//	url := resource.Handler.GetURL()
	//	logrus.Debugf("%s: Tool download progress %d/%d (%d%%)", url.String(), resource.Available, resource.Total, resource.Available*100/resource.Total)
	//})
}

func (systemEngine *SystemEngine) prepareNextTool(first bool) {
	if len(systemEngine.preparingToolsList) > 0 {
		if !first {
			systemEngine.preparingToolsList = systemEngine.preparingToolsList[1:]
		}
		if len(systemEngine.preparingToolsList) > 0 {
			systemEngine.getTool(&systemEngine.preparingToolsList[0])
			return
		}
	}
	//systemEngine.ToolsPreparedEventEmitter.Emit(true)
}

func (systemEngine *SystemEngine) saveCoreFile(consoleEntry *console.Console) {
	logrus.Infof("Core %s downloaded", consoleEntry.Slug)
	if err := systemEngine.extractCoreArchive(consoleEntry); err != nil {
		return
	}
	if err := systemEngine.elaborateCoreArchive(consoleEntry); err != nil {
		return
	}
	logrus.Infof("Core %s completed", consoleEntry.Slug)
	//systemEngine.CoreElaborationCompletedEventEmitter.Emit(true)
}

func (systemEngine *SystemEngine) savePluginFile(consolePlugin *console.ConsolePlugin, consolePluginsFile *console.ConsolePluginsFile, consolePluginsFileIndex int) {
	var (
		err     error
		console console.Console
	)
	if console, err = systemEngine.databaseEngine.GetConsoleByConsolePlugin(consolePlugin); err != nil {
		return
	}
	logrus.Infof("Console plugin file for %s downloaded", console.Slug)
	if err = systemEngine.extractPluginArchive(consolePlugin, consolePluginsFile, consolePluginsFileIndex); err != nil {
		return
	}
	if err = systemEngine.elaboratePluginArchive(consolePlugin, consolePluginsFile, consolePluginsFileIndex); err != nil {
		return
	}
	logrus.Infof("Console plugin file for %s completed", console.Slug)
	//systemEngine.PluginElaborationCompletedEventEmitter.Emit(false)
}

func (systemEngine *SystemEngine) saveToolFile(toolEntry *tool.Tool) {
	logrus.Info("Tool %s downloaded", toolEntry.Slug)
	if err := systemEngine.extractToolArchive(toolEntry); err != nil {
		return
	}
	if err := systemEngine.elaborateToolArchive(toolEntry); err != nil {
		return
	}
	logrus.Info("Tool %s completed", toolEntry.Slug)
	//systemEngine.ToolElaborationCompletedEventEmitter.Emit(false)
}

func (systemEngine *SystemEngine) coreIsDownloaded(consoleEntry *console.Console) bool {
	coreLocation := consoleEntry.CoreLocation + "." + osconstants.CORES_EXTENSION
	if _, err := os.Stat(filepath.Join(folder.CORES, coreLocation)); os.IsNotExist(err) {
		return false
	}
	return true
}

func (systemEngine *SystemEngine) coreIsUpdated(_ *console.Console) bool {
	return true
}

func (systemEngine *SystemEngine) toolIsDownloaded(toolEntry *tool.Tool) bool {
	var toolLocation string
	if toolEntry.Destination.Valid && toolEntry.Destination.String != "" {
		toolLocation = filepath.Join(folder.TOOLS, toolEntry.Destination.String)
	} else if toolEntry.CollectionPath.Valid && toolEntry.CollectionPath.String != "" {
		toolLocation = filepath.Join(folder.TOOLS, filepath.Base(toolEntry.CollectionPath.String))
	} else {
		var (
			toolUrl *url.URL
			err     error
		)
		if toolUrl, err = url.Parse(toolEntry.Url); err != nil {
			return false
		}
		toolLocation = filepath.Join(folder.TOOLS, filepath.Base(toolUrl.Path))
	}
	if _, err := os.Stat(toolLocation); os.IsNotExist(err) {
		return false
	}
	return true
}

func (systemEngine *SystemEngine) toolIsUpdated(_ *tool.Tool) bool {
	return true
}

func (systemEngine *SystemEngine) extractCoreArchive(consoleEntry *console.Console) error {
	process := exec.Command(
		osconstants.SEVENZ_EXE_PATH,
		"x",
		GetDownloadCorePath(consoleEntry),
		"-o"+GetCoreTempPath(consoleEntry))
	if err := process.Run(); err != nil {
		logrus.Error("Error starting the extraction process")
		logrus.Errorf("%+v", err)
		return err
	}
	return nil
}

func (systemEngine *SystemEngine) elaborateCoreArchive(consoleEntry *console.Console) (err error) {
	coreTempPath := GetCoreTempPath(consoleEntry)
	filepath.Walk(coreTempPath, func(filePath string, info fs.FileInfo, err error) error {
		if path.Ext(info.Name()) != "" && path.Ext(info.Name())[1:] == osconstants.CORES_EXTENSION {
			os.Rename(filePath, GetCorePath(consoleEntry))
		}
		return nil
	})
	os.Remove(GetDownloadCorePath(consoleEntry))
	os.RemoveAll(coreTempPath)
	return
}

func (systemEngine *SystemEngine) extractPluginArchive(consolePlugin *console.ConsolePlugin, consolePluginsFiles *console.ConsolePluginsFile, consolePluginsFileIndex int) error {
	if consolePlugin.Type == "bios" {
		consolePluginFilePath := GetDownloadCorePluginPath(consolePlugin, consolePluginsFiles)
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
			osconstants.SEVENZ_EXE_PATH,
			"x",
			GetDownloadCorePluginPath(consolePlugin, consolePluginsFiles),
			"-o"+GetCorePluginTempPath(consolePlugin, consolePluginsFileIndex))
		if err := process.Run(); err != nil {
			logrus.Error("Error starting the extraction process")
			logrus.Errorf("%+v", err)
			return err
		}
	}
	return nil
}

func (systemEngine *SystemEngine) elaboratePluginArchive(consolePlugin *console.ConsolePlugin, consolePluginsFile *console.ConsolePluginsFile, consolePluginsFileIndex int) (err error) {
	if consolePlugin.Type == "bios" {
		consolePluginFilePath := GetDownloadCorePluginPath(consolePlugin, consolePluginsFile)
		destinationFolder := ""
		if consolePluginsFile.Destination.Valid {
			destinationFolder = consolePluginsFile.Destination.String
		}
		if _, err := os.Stat(destinationFolder); os.IsNotExist(err) {
			os.MkdirAll(destinationFolder, 0755)
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
			os.MkdirAll(destinationFolder, 0755)
			os.Rename(consolePluginFilePath, destinationPath)
		} else {
			extractionDir := GetCorePluginTempPath(consolePlugin, consolePluginsFileIndex)
			collectionPath := extractionDir
			if consolePluginsFile.CollectionPath.Valid {
				collectionPath = path.Join(collectionPath, consolePluginsFile.CollectionPath.String)
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
			os.RemoveAll(collectionPath)
			os.RemoveAll(extractionDir)
		}
	}
	return
}

func (systemEngine *SystemEngine) extractToolArchive(toolEntry *tool.Tool) error {
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
		osconstants.SEVENZ_EXE_PATH,
		"x",
		GetDownloadToolPath(toolEntry),
		"-o"+GetToolTempPath(toolEntry))
	if err := process.Run(); err != nil {
		logrus.Error("Error starting the extraction process")
		logrus.Errorf("%+v", err)
		return err
	}
	return nil
}

func (systemEngine *SystemEngine) elaborateToolArchive(toolEntry *tool.Tool) (err error) {
	destinationFolder := folder.TOOLS
	if _, err := os.Stat(destinationFolder); os.IsNotExist(err) {
		os.Mkdir(destinationFolder, 0755)
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

func GetDownloadCorePath(consoleEntry *console.Console) string {
	fileName := consoleEntry.CoreLocation + "." + osconstants.CORES_EXTENSION + ".zip"
	return path.Join(folder.TEMP, fileName)
}

func GetDownloadCorePluginPath(consolePlugin *console.ConsolePlugin, consolePluginFile *console.ConsolePluginsFile) string {
	tempDownloadDir := GetPluginTempPath()
	return path.Join(tempDownloadDir, GetDownloadCorePluginFileName(consolePlugin, consolePluginFile))
}

func GetDownloadCorePluginFileName(consolePlugin *console.ConsolePlugin, consolePluginFile *console.ConsolePluginsFile) string {
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

func GetDownloadToolPath(toolEntry *tool.Tool) (toolPath string) {
	toolPath = folder.TEMP
	if _, err := os.Stat(toolPath); os.IsNotExist(err) {
		os.Mkdir(toolPath, 0755)
	}
	url, _ := url.Parse(toolEntry.Url)
	toolPath = path.Join(toolPath, path.Base(url.Path))
	return
}

func GetCoreTempPath(consoleEntry *console.Console) (tempDownloadDir string) {
	tempDownloadDir = folder.TEMP
	tempDownloadDir = path.Join(tempDownloadDir, consoleEntry.Slug)
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0755)
	}
	return
}

func GetCorePluginTempPath(consolePlugin *console.ConsolePlugin, fileIndex int) string {
	tempDownloadDir := GetPluginTempPath()
	return path.Join(tempDownloadDir, strconv.Itoa(fileIndex))
}

func GetToolTempPath(toolEntry *tool.Tool) (tempDownloadDir string) {
	tempDownloadDir = path.Join(folder.TEMP, toolEntry.Slug)
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.Mkdir(tempDownloadDir, 0755)
	}
	return
}

func GetPluginTempPath() string {
	tempDownloadDir := folder.TEMP
	tempDownloadDir = path.Join(tempDownloadDir, folder.PLUGIN)
	if _, err := os.Stat(tempDownloadDir); os.IsNotExist(err) {
		os.MkdirAll(tempDownloadDir, 0755)
	}
	return tempDownloadDir
}

func GetCorePath(consoleEntry *console.Console) string {
	return path.Join(
		folder.CORES,
		consoleEntry.CoreLocation+"."+osconstants.CORES_EXTENSION)
}

func GetUndertow() resources.StorjResource {
	return resources.StorjResource{
		URL: url.URL{
			Scheme: undertow.DEFAULT_SCHEME,
			Host:   undertow.DEFAULT_HOST,
			Path:   undertow.DEFAULT_PATH,
		},
		Access: undertow.DEFAULT_ACCESS,
	}
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
	if systemFolder, err = filepath.Abs(folder.SYSTEM); err != nil {
		logrus.Error("Cannot get absolute shader folder")
		logrus.Errorf("%+v", err)
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
	var databaseLanguage database.Locale
	databaseLanguage, _ = systemEngine.databaseEngine.GetLanguage()
	language := systemEngine.languageToRetroArchIndex(databaseLanguage)
	systemEngine.settings["user_language"] = language
	systemEngine.syncSettings()

	return
}

func (systemEngine *SystemEngine) languageToRetroArchIndex(databaseLanguage database.Locale) int {
	switch databaseLanguage {
	case database.FRENCH:
		return 2
	case database.SPANISH:
		return 3
	case database.GERMAN:
		return 4
	case database.ITALIAN:
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
