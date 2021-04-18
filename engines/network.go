package engines

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

	"arkhive.dev/launcher/common"
	"arkhive.dev/launcher/models/network"
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
	undertowResource  network.StorjResource
	databaseEngine    *DatabaseEngine
	account           network.Account
	resources         []*network.Resource
	certificateStatus CertificateStatus
	undertowPublicKey *rsa.PublicKey

	// Event emitters
	UserAccountAvailableEventEmitter      *common.EventEmitter
	UserStatusChangedEventEmitter         *common.EventEmitter
	BootedEventEmitter                    *common.EventEmitter
	NetworkProcessInitializedEventEmitter *common.EventEmitter
}

func NewNetworkEngine(databaseEngine *DatabaseEngine, undertowResource network.StorjResource) (instance *NetworkEngine, err error) {
	instance = &NetworkEngine{
		databaseEngine:                        databaseEngine,
		undertowResource:                      undertowResource,
		UserAccountAvailableEventEmitter:      new(common.EventEmitter),
		UserStatusChangedEventEmitter:         new(common.EventEmitter),
		BootedEventEmitter:                    new(common.EventEmitter),
		NetworkProcessInitializedEventEmitter: new(common.EventEmitter),
	}

	if _, err := os.Stat(common.SYSTEM_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.SYSTEM_FOLDER_PATH, 0644)
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
	privateKeyFilePath := path.Join(common.SYSTEM_FOLDER_PATH, "private.bee")
	certificateFilePath := path.Join(common.SYSTEM_FOLDER_PATH, "certificate.bee")
	_, err = os.Stat(privateKeyFilePath)
	privateKeyFileExists := !os.IsNotExist(err)
	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)

	if privateKeyFileExists && certificateFileExists {
		var privateKeyBytes []byte
		if privateKeyBytes, err = os.ReadFile(privateKeyFilePath); err != nil {
			log.Fatal(err)
			return
		}
		var privateKey *rsa.PrivateKey
		if privateKey, err = parsePrivateKey(privateKeyBytes); err != nil {
			log.Fatal("Cannot decode the private key file content")
			log.Fatal(err)
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
		if privateKey, err = generatePairKey(1024); err != nil {
			log.Fatal(err)
			return
		}
		if err = os.WriteFile(privateKeyFilePath, exportPrivateKey(privateKey), 0644); err != nil {
			log.Fatal(err)
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
	certificateFilePath := path.Join(common.SYSTEM_FOLDER_PATH, "certificate.bee")

	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)
	if !certificateFileExists {
		return
	}

	var jsonCertificateData []byte
	if jsonCertificateData, err = os.ReadFile(certificateFilePath); err != nil {
		log.Fatal(err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonCertificateData))
	decoder.UseNumber()
	var jsonCertificateDocument map[string]interface{}
	if err = decoder.Decode(&jsonCertificateDocument); err != nil {
		log.Fatal(err)
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
	if publicKey, err = parsePublicKey(publicKeyBytes); err != nil {
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

	certificateFilePath := path.Join(common.SYSTEM_FOLDER_PATH, "certificate.bee")
	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)
	if !certificateFileExists {
		return
	}
	var jsonCertificateData []byte
	if jsonCertificateData, err = os.ReadFile(certificateFilePath); err != nil {
		log.Fatal(err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonCertificateData))
	decoder.UseNumber()
	var jsonCertificateDocument map[string]interface{}
	if err = decoder.Decode(&jsonCertificateDocument); err != nil {
		log.Fatal(err)
		return
	}

	if signBase64, ok := jsonCertificateDocument["sign"].(string); ok {
		if networkEngine.account.Sign, err = base64.URLEncoding.DecodeString(signBase64); err != nil {
			log.Fatal(err)
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
	if jsonCertificateDocumentDecodedEncrypted, err = Encrypt(networkEngine.undertowPublicKey, jsonCertificateData); err != nil {
		if reflect.DeepEqual(networkEngine.account.Sign, jsonCertificateDocumentDecodedEncrypted) {
			networkEngine.certificateStatus = OFFICIAL
		}
	}
	return
}

func (networkEngine *NetworkEngine) addResource(url *url.URL, path string, allowedFiles ...string) (resource *network.Resource, err error) {
	var resourceHandler network.ResourceHandler
	switch url.Scheme {
	case "http":
		resourceHandler = &network.HTTPResource{
			URL: *url,
		}
	case "https":
		resourceHandler = &network.HTTPResource{
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
		resource = network.NewResource(resourceHandler, path, allowedFiles)
		// ToDo: connections
		go resource.Download()
	}
	return
}

func (networkEngine *NetworkEngine) addUndertow(storjResource *network.StorjResource, isMain bool) error {
	systemPath := common.SYSTEM_FOLDER_PATH
	resource := network.NewResource(storjResource, systemPath, []string{})
	networkEngine.resources = append(networkEngine.resources, resource)
	resource.StatusUpdatedEventEmitter.Subscribe(func(resource *network.Resource) {
		url := resource.Handler.GetURL()
		log.Debug(url.String(), ": Undertow status updated ", resource.Status)
	})
	resource.ProgressUpdatedEventEmitter.Subscribe(func(resource *network.Resource) {
		url := resource.Handler.GetURL()
		log.Debug(url.String(),
			": Undertow download progress ",
			resource.Available, "/", resource.Total,
			" (", resource.Available/resource.Total*100, "%)")
	})
	resource.AvailableEventEmitter.Subscribe(func(resource *network.Resource) {
		var (
			undertowPublicKeyBytes []byte
			err                    error
		)
		if undertowPublicKeyBytes, err = os.ReadFile(path.Join(resource.Path, filepath.Base(resource.Handler.GetURL().Path))); err != nil {
			log.Error(err)
			return
		}
		if networkEngine.undertowPublicKey, err = parsePublicKey(undertowPublicKeyBytes); err != nil {
			log.Error(err)
			return
		}
		networkEngine.UserStatusChangedEventEmitter.Emit(true)
		networkEngine.verifyAccountCertificateSign()
	})
	go resource.Download()

	return nil
}
