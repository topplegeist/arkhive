package network

import (
	"bytes"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"time"

	"arkhive.dev/launcher/internal/engine/database"
	"arkhive.dev/launcher/internal/engine/network/models"
	"arkhive.dev/launcher/internal/engine/network/resources"
	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/pkg/encryption"
	"arkhive.dev/launcher/pkg/eventemitter"
	log "github.com/sirupsen/logrus"
)

type CertificateStatus int

const (
	INVALID CertificateStatus = iota
	AVAILABLE
	UNOFFICIAL
	OFFICIAL
)

type NetworkEngine struct {
	undertowResource  resources.StorjResource
	databaseEngine    *database.DatabaseEngine
	account           models.Account
	resources         []*resources.Resource
	certificateStatus CertificateStatus
	undertowPublicKey *rsa.PublicKey

	// Event emitters
	UserAccountAvailableEventEmitter      *eventemitter.EventEmitter
	UserStatusChangedEventEmitter         *eventemitter.EventEmitter
	BootedEventEmitter                    *eventemitter.EventEmitter
	NetworkProcessInitializedEventEmitter *eventemitter.EventEmitter
}

func NewNetworkEngine(databaseEngine *database.DatabaseEngine, undertowResource resources.StorjResource) (instance *NetworkEngine, err error) {
	instance = &NetworkEngine{
		databaseEngine:                        databaseEngine,
		undertowResource:                      undertowResource,
		UserAccountAvailableEventEmitter:      new(eventemitter.EventEmitter),
		UserStatusChangedEventEmitter:         new(eventemitter.EventEmitter),
		BootedEventEmitter:                    new(eventemitter.EventEmitter),
		NetworkProcessInitializedEventEmitter: new(eventemitter.EventEmitter),
	}

	if _, err := os.Stat(folder.SYSTEM); os.IsNotExist(err) {
		os.Mkdir(folder.SYSTEM, 0644)
	}

	go instance.importUserCryptoData()

	databaseEngine.DecryptedEventEmitter.Subscribe(func(_ bool) {
		go instance.initNetworkProcess()
	})

	return
}

func (networkEngine NetworkEngine) isUserCertificateAvailable() bool {
	return networkEngine.certificateStatus != INVALID
}

func (networkEngine *NetworkEngine) importUserCryptoData() (err error) {
	privateKeyFilePath := path.Join(folder.SYSTEM, "private.bee")
	certificateFilePath := path.Join(folder.SYSTEM, "certificate.bee")
	_, err = os.Stat(privateKeyFilePath)
	privateKeyFileExists := !os.IsNotExist(err)
	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)

	if privateKeyFileExists && certificateFileExists {
		var privateKeyBytes []byte
		if privateKeyBytes, err = os.ReadFile(privateKeyFilePath); err != nil {
			log.Error(err)
			return
		}
		var privateKey *rsa.PrivateKey
		if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
			log.Error("Cannot decode the private key file content")
			log.Error(err)
			return
		}
		networkEngine.account.PrivateKey = *privateKey
		if readCertificateError := networkEngine.readAccountCertificate(); readCertificateError == nil {
			networkEngine.certificateStatus = AVAILABLE
			networkEngine.UserAccountAvailableEventEmitter.Emit(true)
			networkEngine.UserStatusChangedEventEmitter.Emit(true)
		} else {
			log.Warn("Error reading the user certificate")
			log.Error(err)
		}
	}

	if !networkEngine.isUserCertificateAvailable() {
		var privateKey *rsa.PrivateKey
		if privateKey, err = encryption.GeneratePairKey(1024); err != nil {
			log.Error(err)
			return
		}
		if err = os.WriteFile(privateKeyFilePath, encryption.ExportPrivateKey(privateKey), 0644); err != nil {
			log.Error(err)
			return
		}
		networkEngine.account.PrivateKey = *privateKey
		networkEngine.account.PublicKey = privateKey.PublicKey
	}

	networkEngine.BootedEventEmitter.Emit(true)
	return
}

func (networkEngine *NetworkEngine) initNetworkProcess() {
	networkEngine.addUndertow(&networkEngine.undertowResource, true)
}

func (networkEngine NetworkEngine) readAccountCertificate() (err error) {
	certificateFilePath := path.Join(folder.SYSTEM, "certificate.bee")

	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)
	if !certificateFileExists {
		return
	}

	var jsonCertificateData []byte
	if jsonCertificateData, err = os.ReadFile(certificateFilePath); err != nil {
		log.Error(err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonCertificateData))
	decoder.UseNumber()
	var jsonCertificateDocument map[string]interface{}
	if err = decoder.Decode(&jsonCertificateDocument); err != nil {
		log.Error(err)
		return
	}
	networkEngine.account.Username = jsonCertificateDocument["username"].(string)
	networkEngine.account.Email = jsonCertificateDocument["email"].(string)
	var registrationDate int64
	if registrationDate, err = jsonCertificateDocument["date"].(json.Number).Int64(); err != nil {
		return
	}
	networkEngine.account.RegistrationDate = time.Unix(registrationDate, 0)
	var publicKeyBytes []byte
	if publicKeyBytes, err = base64.URLEncoding.DecodeString(jsonCertificateDocument["date"].(string)); err != nil {
		return
	}
	var publicKey *rsa.PublicKey
	if publicKey, err = encryption.ParsePublicKey(publicKeyBytes); err != nil {
		return
	}
	networkEngine.account.PublicKey = *publicKey
	return
}

func (networkEngine NetworkEngine) verifyAccountCertificateSign() (err error) {
	if networkEngine.undertowPublicKey == nil {
		return errors.New("undertow file not downloaded")
	}
	if !networkEngine.isUserCertificateAvailable() {
		return errors.New("unextistent certificate file")
	}

	certificateFilePath := path.Join(folder.SYSTEM, "certificate.bee")
	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)
	if !certificateFileExists {
		return
	}
	var jsonCertificateData []byte
	if jsonCertificateData, err = os.ReadFile(certificateFilePath); err != nil {
		log.Error(err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonCertificateData))
	decoder.UseNumber()
	var jsonCertificateDocument map[string]interface{}
	if err = decoder.Decode(&jsonCertificateDocument); err != nil {
		log.Error(err)
		return
	}

	if signBase64, ok := jsonCertificateDocument["sign"].(string); ok {
		if networkEngine.account.Sign, err = base64.URLEncoding.DecodeString(signBase64); err != nil {
			log.Error(err)
			return
		}
	}
	networkEngine.certificateStatus = UNOFFICIAL

	jsonCertificateDocumentDecoded := make(map[string]interface{})
	var ok bool = true
	if ok {
		jsonCertificateDocumentDecoded["username"], ok = jsonCertificateDocument["username"].(string)
	}
	if ok {
		jsonCertificateDocumentDecoded["email"], ok = jsonCertificateDocument["email"].(string)
	}
	if jsonCertificateDocumentDecoded["date"], err = jsonCertificateDocument["date"].(json.Number).Int64(); err != nil {
		ok = false
	}
	if ok {
		jsonCertificateDocumentDecoded["public_key"], ok = jsonCertificateDocument["public_key"].(string)
	}
	if !ok {
		err = errors.New("invalid certificate values")
		return
	}
	var jsonCertificateDocumentDecodedEncrypted []byte
	if jsonCertificateDocumentDecodedEncrypted, err = encryption.Encrypt(networkEngine.undertowPublicKey, jsonCertificateData); err != nil {
		if reflect.DeepEqual(networkEngine.account.Sign, jsonCertificateDocumentDecodedEncrypted) {
			networkEngine.certificateStatus = OFFICIAL
		}
	}
	return
}

func (networkEngine *NetworkEngine) AddResource(url *url.URL, path string, allowedFiles ...string) (resource *resources.Resource, err error) {
	var resourceHandler resources.ResourceHandler
	switch url.Scheme {
	case "http":
		resourceHandler = &resources.HTTPResource{
			URL: *url,
		}
	case "https":
		resourceHandler = &resources.HTTPResource{
			URL: *url,
		}
	case "file":
		err = errors.New("url schema not allowed")
	case "torrent":
		err = errors.New("url schema not allowed")
	case "magnet":
		err = errors.New("url schema not allowed")
	}
	if err == nil {
		resource = resources.NewResource(resourceHandler, path, allowedFiles)
		// ToDo: connections
		go resource.Download()
	}
	return
}

func (networkEngine *NetworkEngine) addUndertow(storjResource *resources.StorjResource, isMain bool) error {
	systemPath := folder.SYSTEM
	resource := resources.NewResource(storjResource, systemPath, []string{})
	networkEngine.resources = append(networkEngine.resources, resource)
	resource.StatusUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
		url := resource.Handler.GetURL()
		log.Debug(url.String(), ": Undertow status updated ", resource.Status)
	})
	resource.ProgressUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
		url := resource.Handler.GetURL()
		log.Debug(url.String(),
			": Undertow download progress ",
			resource.Available, "/", resource.Total,
			" (", resource.Available*100/resource.Total, "%)")
	})
	resource.AvailableEventEmitter.Subscribe(func(resource *resources.Resource) {
		var (
			undertowPublicKeyBytes []byte
			err                    error
		)
		if undertowPublicKeyBytes, err = os.ReadFile(path.Join(resource.Path, filepath.Base(resource.Handler.GetURL().Path))); err != nil {
			log.Error(err)
			return
		}
		if networkEngine.undertowPublicKey, err = encryption.ParsePublicKey(undertowPublicKeyBytes); err != nil {
			log.Error(err)
			return
		}
		networkEngine.UserStatusChangedEventEmitter.Emit(true)
		networkEngine.verifyAccountCertificateSign()
	})
	go resource.Download()

	return nil
}
