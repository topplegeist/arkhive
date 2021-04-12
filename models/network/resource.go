package network

import (
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"arkhive.dev/launcher/common"
	log "github.com/sirupsen/logrus"
)

type ResourceStatus int

const (
	PENDING ResourceStatus = iota
	SEARCHING_PEERS
	DOWNLOADING
	DOWNLOADED
	DOWNLOADING_TORRENT
	TORRENT_DOWNLOADED
	ABORTING
	ERROR
)

type ResourceHandler interface {
	GetURL() url.URL
}

type Resource struct {
	Handler                   ResourceHandler
	Path                      string
	AllowedFiles              []string
	Total                     int64
	Available                 int64
	Status                    ResourceStatus
	AvailableEventEmitter     *common.EventEmitter
	RemovingEventEmitter      *common.EventEmitter
	StatusUpdatedEventEmitter *common.EventEmitter
}

func NewResource(storjResource ResourceHandler, systemPath string, allowedFiles []string) *Resource {
	return &Resource{
		Handler:                   storjResource,
		Path:                      systemPath,
		AllowedFiles:              allowedFiles,
		Status:                    PENDING,
		AvailableEventEmitter:     new(common.EventEmitter),
		RemovingEventEmitter:      new(common.EventEmitter),
		StatusUpdatedEventEmitter: new(common.EventEmitter),
	}
}

func (resource *Resource) SetStatus(status ResourceStatus) {
	resource.Status = status
	switch resource.Status {
	case DOWNLOADED:
		resource.AvailableEventEmitter.Emit(true)
	case ABORTING:
		resource.RemovingEventEmitter.Emit(true)
	}
	resource.StatusUpdatedEventEmitter.Emit(resource.Status)
}

func (resource *Resource) Write(buffer []byte) (int, error) {
	bufferSize := len(buffer)
	resource.Available += int64(bufferSize)
	resource.PrintProgress()
	return bufferSize, nil
}

func (resource *Resource) PrintProgress() {
	log.Info("Downloading... ", resource.Available, "/", resource.Total, " (", resource.Available/resource.Total*100, "%)")
}

func (resource *Resource) Save(reader io.Reader) {
	out, err := os.Create(path.Join(resource.Path, filepath.Base(resource.Handler.GetURL().Path)))
	if err != nil {
		log.Error(err)
		return
	}
	if _, err := io.Copy(out, io.TeeReader(reader, resource)); err != nil {
		log.Error(err)
		return
	}
}
