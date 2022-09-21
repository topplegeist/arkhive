package resources

import (
	"io"
	"net/url"
	"os"
	"path"
	"path/filepath"

	"github.com/sirupsen/logrus"
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
	Download(resource *Resource)
}

type Resource struct {
	Handler      ResourceHandler
	Path         string
	AllowedFiles []string
	Total        int64
	Available    int64
	Status       ResourceStatus
}

func NewResource(resourceHandler ResourceHandler, resourcePath string, allowedFiles []string) *Resource {
	return &Resource{
		Handler:      resourceHandler,
		Path:         resourcePath,
		AllowedFiles: allowedFiles,
		Status:       PENDING,
	}
}

func (resource *Resource) SetStatus(status ResourceStatus) {
	resource.Status = status
	switch resource.Status {
	case DOWNLOADED:
		//resource.AvailableEventEmitter.Emit(resource)
		break
	case ABORTING:
		//resource.RemovingEventEmitter.Emit(resource)
		break
	}
	//resource.StatusUpdatedEventEmitter.Emit(resource)
}

func (resource *Resource) Write(buffer []byte) (int, error) {
	bufferSize := len(buffer)
	resource.Available += int64(bufferSize)
	//resource.ProgressUpdatedEventEmitter.Emit(resource)
	return bufferSize, nil
}

func (resource *Resource) Download() {
	resource.Handler.Download(resource)
}

func (resource *Resource) Save(reader io.Reader) error {
	out, err := os.Create(path.Join(resource.Path, filepath.Base(resource.Handler.GetURL().Path)))
	if err != nil {
		logrus.Errorf("%+v", err)
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, io.TeeReader(reader, resource)); err != nil {
		logrus.Errorf("%+v", err)
		return err
	}
	return nil
}
