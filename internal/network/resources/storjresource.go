package resources

import (
	"context"
	"net/url"

	"github.com/sirupsen/logrus"
	"storj.io/uplink"
)

type StorjResource struct {
	URL    url.URL
	Access string
}

func (storjResource *StorjResource) GetURL() url.URL {
	return storjResource.URL
}

func (storjResource StorjResource) Download(resource *Resource) {
	userAccess, err := uplink.ParseAccess(storjResource.Access)
	if err != nil {
		resource.SetStatus(ERROR)
		logrus.Errorf("%+v", err)
		return
	}
	project, err := uplink.OpenProject(context.Background(), userAccess)
	if err != nil {
		resource.SetStatus(ERROR)
		logrus.Errorf("%+v", err)
		return
	}
	resource.SetStatus(DOWNLOADING)
	stat, err := project.StatObject(context.Background(),
		resource.Handler.GetURL().Host,
		resource.Handler.GetURL().Path)
	if err != nil {
		resource.SetStatus(ERROR)
		logrus.Errorf("%+v", err)
		return
	}
	resource.Total = stat.System.ContentLength
	download, err := project.DownloadObject(context.Background(),
		resource.Handler.GetURL().Host,
		resource.Handler.GetURL().Path, nil)
	if err != nil {
		resource.SetStatus(ERROR)
		logrus.Errorf("%+v", err)
		return
	}
	defer download.Close()
	if err := resource.Save(download); err != nil {
		resource.SetStatus(ERROR)
		logrus.Errorf("%+v", err)
		return
	}
	resource.SetStatus(DOWNLOADED)
}
