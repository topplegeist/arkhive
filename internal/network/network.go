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
	"reflect"
	"sync"
	"time"

	"arkhive.dev/launcher/internal/folder"
	"arkhive.dev/launcher/internal/network/models"
	"arkhive.dev/launcher/internal/network/resources"
	"arkhive.dev/launcher/pkg/encryption"
	"github.com/sirupsen/logrus"
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
	account           models.Account
	resources         []*resources.Resource
	certificateStatus CertificateStatus
	undertowPublicKey *rsa.PublicKey
}

func NewNetworkEngine() (instance *NetworkEngine, err error) {
	instance = &NetworkEngine{}
	return
}

func (networkEngine *NetworkEngine) Initialize(waitGroup *sync.WaitGroup) {
	if _, err := os.Stat(folder.SYSTEM); os.IsNotExist(err) {
		if err = os.Mkdir(folder.SYSTEM, 0755); err != nil {
			panic(err)
		}
	}

	go networkEngine.importUserCryptoData()
}

func (networkEngine NetworkEngine) isUserCertificateAvailable() bool {
	return networkEngine.certificateStatus != INVALID
}

func (networkEngine *NetworkEngine) importUserCryptoData() {
	var err error
	privateKeyFilePath := path.Join(folder.SYSTEM, "private.bee")
	certificateFilePath := path.Join(folder.SYSTEM, "certificate.bee")
	_, err = os.Stat(privateKeyFilePath)
	privateKeyFileExists := !os.IsNotExist(err)
	_, err = os.Stat(certificateFilePath)
	certificateFileExists := !os.IsNotExist(err)

	if privateKeyFileExists && certificateFileExists {
		var privateKeyBytes []byte
		if privateKeyBytes, err = os.ReadFile(privateKeyFilePath); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
		var privateKey *rsa.PrivateKey
		if privateKey, err = encryption.ParsePrivateKey(privateKeyBytes); err != nil {
			logrus.Error("Cannot decode the private key file content")
			logrus.Errorf("%+v", err)
			return
		}
		networkEngine.account.PrivateKey = *privateKey
		if readCertificateError := networkEngine.readAccountCertificate(); readCertificateError == nil {
			networkEngine.certificateStatus = AVAILABLE
			//networkEngine.UserAccountAvailableEventEmitter.Emit(true)
			//networkEngine.UserStatusChangedEventEmitter.Emit(true)
		} else {
			logrus.Warn("Error reading the user certificate")
			logrus.Errorf("%+v", err)
		}
	}

	if !networkEngine.isUserCertificateAvailable() {
		var privateKey *rsa.PrivateKey
		if privateKey, err = encryption.GeneratePairKey(1024); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
		if err = os.WriteFile(privateKeyFilePath, encryption.ExportPrivateKey(privateKey), 0644); err != nil {
			logrus.Errorf("%+v", err)
			return
		}
		networkEngine.account.PrivateKey = *privateKey
		networkEngine.account.PublicKey = privateKey.PublicKey
	}

	//networkEngine.BootedEventEmitter.Emit(true)
}

func (networkEngine *NetworkEngine) initNetworkProcess() (err error) {
	err = networkEngine.addUndertow(&networkEngine.undertowResource, true)
	return
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
		logrus.Errorf("%+v", err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonCertificateData))
	decoder.UseNumber()
	var jsonCertificateDocument map[string]interface{}
	if err = decoder.Decode(&jsonCertificateDocument); err != nil {
		logrus.Errorf("%+v", err)
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
		logrus.Errorf("%+v", err)
		return
	}
	decoder := json.NewDecoder(bytes.NewReader(jsonCertificateData))
	decoder.UseNumber()
	var jsonCertificateDocument map[string]interface{}
	if err = decoder.Decode(&jsonCertificateDocument); err != nil {
		logrus.Errorf("%+v", err)
		return
	}

	if signBase64, ok := jsonCertificateDocument["sign"].(string); ok {
		if networkEngine.account.Sign, err = base64.URLEncoding.DecodeString(signBase64); err != nil {
			logrus.Errorf("%+v", err)
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
		go resource.Download()
	}
	return
}

func (networkEngine *NetworkEngine) addUndertow(storjResource *resources.StorjResource, isMain bool) error {
	systemPath := folder.SYSTEM
	resource := resources.NewResource(storjResource, systemPath, []string{})
	networkEngine.resources = append(networkEngine.resources, resource)
	//resource.StatusUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
	//	url := resource.Handler.GetURL()
	//	logrus.Debugf("%s: Undertow status updated %d", url.String(), resource.Status)
	//})
	//resource.ProgressUpdatedEventEmitter.Subscribe(func(resource *resources.Resource) {
	//	url := resource.Handler.GetURL()
	//	logrus.Debugf("%s: Undertow download progress %d/%d (%d%%)", url.String(), resource.Available, resource.Total, resource.Available*100/resource.Total)
	//})
	//resource.AvailableEventEmitter.Subscribe(func(resource *resources.Resource) {
	//	var (
	//		undertowPublicKeyBytes []byte
	//		err                    error
	//	)
	//	if undertowPublicKeyBytes, err = os.ReadFile(path.Join(resource.Path, filepath.Base(resource.Handler.GetURL().Path))); err != nil {
	//		logrus.Errorf("%+v", err)
	//		return
	//	}
	//	if networkEngine.undertowPublicKey, err = encryption.ParsePublicKey(undertowPublicKeyBytes); err != nil {
	//		logrus.Errorf("%+v", err)
	//		return
	//	}
	//	networkEngine.UserStatusChangedEventEmitter.Emit(true)
	//	networkEngine.verifyAccountCertificateSign()
	//})
	go resource.Download()

	return nil
}
