package network

import (
	"net/url"

	"arkhive.dev/launcher/common"
)

type StorjResource struct {
	URL    url.URL
	Access string
}

func GetUndertow() StorjResource {
	return StorjResource{
		URL:    common.GetDefaultUndertowURL(),
		Access: common.DEFAULT_UNDERTOW_ACCESS,
	}
}

func (storjResource StorjResource) GetURL() url.URL {
	return storjResource.URL
}
