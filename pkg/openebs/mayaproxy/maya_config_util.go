package mayaproxy

import (
	"net/url"
	"strings"
	"errors"
	"time"
)

const (
	// maya api server request path constants
	volumes = "/volumes"
	delete  = "/delete"
	info    = "/info"

	// maya api server versions
	VersionLatest = "/latest"

	MayaApiServerService = "maya-apiserver-service"

	// default timeout value
	timeout = 60 * time.Second
)

type MayaApiServerUrl interface {
	GetVolumeURL(version string) (url.URL, error)
	GetVolumeDeleteURL(version string) (url.URL, error)
	GetVolumeInfoURL(version, volumeName string) (url.URL, error)
}

// returns error  if version is empty.
func checkVersion(version string) error {
	//TODO: Add more checks
	if strings.TrimSpace(version) == "" {
		return errors.New("invalid version")
	}
	return nil
}

// GetVolumeURL returns the volume url of mApi server
func (mayaConfig *MayaConfig) GetVolumeURL(version string) (url.URL, error) {
	err := checkVersion(version)
	if err != nil {
		return url.URL{}, err
	}
	return url.URL{
		Scheme: mayaConfig.mapiURI.Scheme,
		Host:   mayaConfig.mapiURI.Host,
		Path:   version + volumes,
	}, nil
}

// GetVolumeDeleteURL returns the volume url of mApi server
func (mayaConfig *MayaConfig) GetVolumeDeleteURL(version string) (url.URL, error) {
	deleteVolumeURL, err := mayaConfig.GetVolumeURL(version)
	if err != nil {
		return url.URL{}, err
	}
	deleteVolumeURL.Path = deleteVolumeURL.Path + delete
	return deleteVolumeURL, nil
}

// GetVolumeInfoURL returns the info url for a volume
func (mayaConfig *MayaConfig) GetVolumeInfoURL(version, volumeName string) (url.URL, error) {
	volumeInfoURL, err := mayaConfig.GetVolumeURL(version)
	if err != nil {
		return url.URL{}, err
	}
	volumeInfoURL.Path = volumeInfoURL.Path + info + "/" + volumeName
	return volumeInfoURL, nil
}
