package mayaproxy

import (
	"testing"
	"net/url"
)

const (
	urlScheme = "https"
	mApiUrl   = "default.svc.cluster.local:5656"
)

var (
	mayaConfig = MayaConfig{
		mapiURI: url.URL{
			Scheme: urlScheme,
			Host:   mApiUrl,
		},
	}
)

func TestCheckVersion(t *testing.T) {
	version := ""
	_, err := mayaConfig.GetVolumeURL(version)
	if err == nil {
		t.Errorf("version value: \"%v\" should cause an error", version)
	}
}

func TestGetVolumeURL(t *testing.T) {

	obtainedUrl, err := mayaConfig.GetVolumeURL(VersionLatest)
	if err != nil {
		t.Error(err)
	}
	expectedVolumeUrl := mayaConfig.mapiURI.String() + "/" + "latest/volumes"

	if obtainedUrl.String() != expectedVolumeUrl {
		t.Errorf("Expected %s got %s", expectedVolumeUrl, obtainedUrl.String())
	}
}

func TestGetVolumeDeleteURL(t *testing.T) {

	obtainedUrl, err := mayaConfig.GetVolumeDeleteURL(VersionLatest)
	if err != nil {
		t.Error(err)
	}
	expectedVolumeUrl := mayaConfig.mapiURI.String() + "/" + "latest/volumes/delete"

	if obtainedUrl.String() != expectedVolumeUrl {
		t.Errorf("Expected %s got %s", expectedVolumeUrl, obtainedUrl.String())
	}
}

func TestGetVolumeInfoURL(t *testing.T) {

	obtainedUrl, err := mayaConfig.GetVolumeInfoURL(VersionLatest, "pvc-1212")
	if err != nil {
		t.Error(err)
	}
	expectedVolumeUrl := mayaConfig.mapiURI.String() + "/" + "latest/volumes/info/pvc-1212"

	if obtainedUrl.String() != expectedVolumeUrl {
		t.Errorf("Expected %s got %s", expectedVolumeUrl, obtainedUrl.String())
	}
}
