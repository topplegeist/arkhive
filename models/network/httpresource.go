package network

import (
	"net/http"
	"net/url"

	log "github.com/sirupsen/logrus"
)

type HTTPResource struct {
	URL url.URL
}

func (httpResource *HTTPResource) GetURL() url.URL {
	return httpResource.URL
}

func (httpResource *HTTPResource) Download(resource *Resource) {
	var (
		response *http.Response
		err      error
	)
	if response, err = http.Get(httpResource.URL.String()); err != nil {
		resource.SetStatus(ERROR)
		log.Error(err)
		return
	}
	resource.Total = response.ContentLength
	if err := resource.Save(response.Body); err != nil {
		resource.SetStatus(ERROR)
		log.Error(err)
		return
	}
	resource.SetStatus(DOWNLOADED)
}
