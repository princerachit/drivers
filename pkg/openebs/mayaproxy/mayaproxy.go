package mayaproxy

import (
	"net/url"
	"os"
	"github.com/golang/glog"
	"kubernetes/client-go/kubernetes"
	"net/http"
	"bytes"
	"io/ioutil"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"
	"gopkg.in/yaml.v2"
)

type MayaApiService interface {
	GetMayaClusterIP(client kubernetes.Interface) (string, error)
	ProxyCreateVolume() (error)
	ProxyDeleteVolume() (error)
	MayaApiServerUrl
}

// MayaConfig is an aggregation all configuration related to mApi server
type MayaConfig struct {
	// Maya-API Server URL running in the cluster
	mapiURI url.URL
	MayaApiService
}

func (mayaConfig MayaConfig) GetMayaClusterIP(client kubernetes.Interface) (string, error) {
	namespace := os.Getenv("OPENEBS_NAMESPACE")
	if namespace == "" {
		namespace = "default"
	}

	glog.Info("OpenEBS volume provisioner namespace ", namespace)

	//Fetch the Maya ClusterIP using the Maya API Server Service
	sc, err := client.CoreV1().Services(namespace).Get(MayaApiServerService)
	if err != nil {
		glog.Errorf("Error getting %s IP Address: %v", MayaApiServerService, err)
	}

	clusterIP := sc.Spec.ClusterIP
	glog.V(2).Infof("Maya Cluster IP: %v", clusterIP)

	return clusterIP, err
}

// ProxyCreateVolume requests mapi server to create an openebs volume. It returns an error if volume creation fails
func (mayaConfig *MayaConfig) ProxyCreateVolume(vs mayav1.VolumeSpec) (error) {
	// add this where we initialize mayaconfig
	// addr := os.Getenv("MAPI_ADDR")

	//url := addr + "/latest/volumes/"

	vs.Kind = "PersistentVolumeClaim"
	vs.APIVersion = "v1"

	// Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(vs)

	glog.V(2).Infof("[DEBUG] volume Spec Created:\n%v\n", string(yamlValue))

	url, err := mayaConfig.GetVolumeURL(VersionLatest)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", url.String(), bytes.NewBuffer(yamlValue))
	req.Header.Add("Content-Type", "application/yaml")
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)

	if err != nil {
		glog.Errorf("Error when connecting maya-apiserver %v", err)
		return err
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		glog.Errorf("Unable to read response from maya-apiserver %v", err)
		return err
	}

	code := resp.StatusCode
	if code != http.StatusOK {
		glog.Errorf("Status error: %v\n", http.StatusText(code))
		return nil
	}

	glog.Infof("volume Successfully Created:\n%v\n", string(data))
	return nil
}

// ProxyDeleteVolume requests mapi server to delete an openebs volume. It returns an error if volume deletion fails
func (mayaConfig *MayaConfig) ProxyDeleteVolume() (error) {

	return nil
}

/*
func (mayaConfig *MayaConfig) ProxyListVolume() (bool, error) {

	return true, nil
}
*/
