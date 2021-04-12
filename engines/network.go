package engines

import (
	"bytes"
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"os"
	"path"
	"time"

	"arkhive.dev/launcher/common"
	"arkhive.dev/launcher/models/network"
	log "github.com/sirupsen/logrus"
	"storj.io/uplink"
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

	// Signals
	UserAccountAvailable      *common.EventEmitter
	UserStatusChanged         *common.EventEmitter
	Booted                    *common.EventEmitter
	NetworkProcessInitialized *common.EventEmitter
}

func NewNetworkEngine(databaseEngine *DatabaseEngine, undertowResource network.StorjResource) (instance *NetworkEngine, err error) {
	instance = &NetworkEngine{
		databaseEngine:            databaseEngine,
		undertowResource:          undertowResource,
		UserAccountAvailable:      new(common.EventEmitter),
		UserStatusChanged:         new(common.EventEmitter),
		Booted:                    new(common.EventEmitter),
		NetworkProcessInitialized: new(common.EventEmitter),
	}

	if _, err := os.Stat(common.SYSTEM_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.SYSTEM_FOLDER_PATH, 0755)
	}
	if _, err := os.Stat(common.TEMP_DOWNLOAD_FOLDER_PATH); os.IsNotExist(err) {
		os.Mkdir(common.TEMP_DOWNLOAD_FOLDER_PATH, 0755)
	}

	go instance.importUserCryptoData()

	databaseEngine.DecryptedEventEmitter.Subscribe(func(_ bool) {
		go instance.initNetworkProcess()
	})

	return
}

func (networkEngine NetworkEngine) isUserCertificateAvailable() bool {
	return networkEngine.certificateStatus == OFFICIAL
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
			return
		}
		networkEngine.account.PrivateKey = *privateKey
		if readCertificateError := networkEngine.readAccountCertificate(); readCertificateError == nil {
			networkEngine.certificateStatus = AVAILABLE
			networkEngine.UserAccountAvailable.Emit(true)
			networkEngine.UserStatusChanged.Emit(true)
		}
	}

	if !networkEngine.isUserCertificateAvailable() {
		var privateKey *rsa.PrivateKey
		if privateKey, err = generatePairKey(1024); err != nil {
			log.Fatal(err)
			return
		}
		if err = os.WriteFile(privateKeyFilePath, exportPrivateKey(privateKey), 0755); err != nil {
			log.Fatal(err)
			return
		}
		networkEngine.account.PrivateKey = *privateKey
		networkEngine.account.PublicKey = privateKey.PublicKey
	}

	networkEngine.Booted.Emit(true)
	return
}

func (networkEngine *NetworkEngine) initNetworkProcess() {
	// ToDo: Setup client parameters
	// defer networkEngine.session.Close()
	// ToDo: Load session resume data
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

func (networkEngine *NetworkEngine) addUndertow(storjResource *network.StorjResource, isMain bool) error {
	systemPath := common.SYSTEM_FOLDER_PATH
	resource := network.NewResource(storjResource, systemPath, []string{})
	networkEngine.resources = append(networkEngine.resources, resource)
	resource.StatusUpdatedEventEmitter.Subscribe(func(status network.ResourceStatus) {
		log.Debug(resource.Path, ": Undertow status updated ", resource.Status)
	})
	resource.AvailableEventEmitter.Subscribe(func(_ bool) {
		networkEngine.NetworkProcessInitialized.Emit(true)
	})
	go func() {
		userAccess, err := uplink.ParseAccess(storjResource.Access)
		if err != nil {
			resource.SetStatus(network.ERROR)
			log.Error(err)
			return
		}
		project, err := uplink.OpenProject(context.Background(), userAccess)
		if err != nil {
			resource.SetStatus(network.ERROR)
			log.Error(err)
			return
		}
		resource.SetStatus(network.DOWNLOADING)
		stat, err := project.StatObject(context.Background(), resource.Handler.GetURL().Host, resource.Handler.GetURL().Path)
		if err != nil {
			resource.SetStatus(network.ERROR)
			log.Error(err)
			return
		}
		resource.Total = stat.System.ContentLength
		download, err := project.DownloadObject(context.Background(), resource.Handler.GetURL().Host, resource.Handler.GetURL().Path, nil)
		if err != nil {
			resource.SetStatus(network.ERROR)
			log.Error(err)
			return
		}
		defer download.Close()
		resource.Save(download)
		resource.SetStatus(network.DOWNLOADED)
	}()

	return nil
}
