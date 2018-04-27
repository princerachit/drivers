package mayaproxy

import (
	"net/url"
	"os"
	"encoding/json"
	"github.com/golang/glog"
	"k8s.io/client-go/kubernetes"
	"net/http"
	"bytes"
	"io/ioutil"
	mayav1 "github.com/kubernetes-incubator/external-storage/openebs/types/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"gopkg.in/yaml.v2"
	"k8s.io/client-go/rest"
	"k8s.io/api/core/v1"
	"github.com/pkg/errors"
)

type MayaApiService interface {
	GetMayaClusterIP(client kubernetes.Interface) (string, error)
	ProxyCreateVolume() (error)
	ProxyDeleteVolume() (error)
	MayaApiServerUrl
}
type K8sClientService interface {
	getK8sClient() (*kubernetes.Clientset, error)
	getSvcObject(client *kubernetes.Clientset, namespace string) (*v1.Service, error)
}

type K8sClient struct{}

// MayaConfig is an aggregate of configurations related to mApi server
type MayaConfig struct {
	// Maya-API Server URL running in the cluster
	mapiURI url.URL
	// namespace where openebs operator runs
	namespace string

	MayaApiService
}

func (k8sClient K8sClient) getK8sClient() (*kubernetes.Clientset, error) {
	// Create an InClusterConfig and use it to create a client for the controller
	// to use to communicate with Kubernetes
	config, err := rest.InClusterConfig()
	if err != nil {
		glog.Errorf("Failed to create config: %v", err)
		return nil, err
	}
	client, err := kubernetes.NewForConfig(config)
	if err != nil {
		glog.Errorf("Failed to create client: %v", err)
		return nil, err
	}

	return client, nil
}

func (k8sClient K8sClient) getSvcObject(client *kubernetes.Clientset, namespace string) (*v1.Service, error) {
	return client.CoreV1().Services(namespace).Get("maya-apiserver-service", metav1.GetOptions{})
}

/*
Using a mutatable receiver. This reduces chances of object tossing from one method to another.
The callee must check for error before proceeding to use MayaConfig
*/
func (mayaConfig *MayaConfig) SetupMayaConfig(k8sClient K8sClientService) error {

	// setup namespace
	if mayaConfig.namespace = os.Getenv("OPENEBS_NAMESPACE"); mayaConfig.namespace == "" {
		mayaConfig.namespace = "default"
	}
	glog.Info("OpenEBS volume provisioner namespace ", mayaConfig.namespace)

	client, err := k8sClient.getK8sClient()
	if err != nil {
		return errors.New("Error creating kubernetes clientset")
	}

	// setup mapiURI using the Maya API Server Service
	svc, err := k8sClient.getSvcObject(client, mayaConfig.namespace)
	if err != nil {
		glog.Errorf("Error getting maya-apiserver IP Address: %v", err)
		return errors.New("Error creating kubernetes clientset")
	}

	mapiUrl, err := url.Parse(svc.Spec.ClusterIP)
	if err != nil {
		glog.Errorf("Could not parse maya-apiserver server url: %v", err)
		return err
	}
	mayaConfig.mapiURI = *mapiUrl
	glog.V(2).Infof("Maya Cluster IP: %v", mayaConfig.mapiURI)
	return nil
}

// ProxyCreateVolume requests mapi server to create an openebs volume. It returns an error if volume creation fails
func (mayaConfig MayaConfig) ProxyCreateVolume(spec mayav1.VolumeSpec) error {
	spec.Kind = "PersistentVolumeClaim"
	spec.APIVersion = "v1"

	// Marshal serializes the value provided into a YAML document
	yamlValue, _ := yaml.Marshal(spec)

	glog.V(2).Infof("[DEBUG] volume Spec Created:\n%v\n", string(yamlValue))

	url, err := mayaConfig.GetVolumeURL(versionLatest)
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
		glog.Errorf("Error response from maya-apiserver: %v", http.StatusText(code))
		return errors.New("Error response from maya-apiserver")
	}

	glog.Infof("volume Successfully Created:\n%v\n", string(data))
	return nil
}

// ProxyDeleteVolume requests mapi server to delete an openebs volume. It returns an error if volume deletion fails
func (mayaConfig MayaConfig) ProxyDeleteVolume(volumeName string) error {
	glog.V(2).Infof("[DEBUG] Delete Volume :%v", string(volumeName))

	url, err := mayaConfig.GetVolumeDeleteURL(versionLatest, volumeName)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)

	if err != nil {
		glog.Errorf("Error when connecting to maya-apiserver  %v", err)
		return err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		glog.Errorf("HTTP Status error from maya-apiserver: %v\n", http.StatusText(code))
		return err
	}
	glog.Info("volume Deleted Successfully initiated")
	return nil
}

func (mayaConfig MayaConfig) ProxyListVolume(volumeName string) (*mayav1.Volume, error) {
	var volume mayav1.Volume

	glog.V(2).Infof("[DEBUG] Get details for Volume :%v", string(volumeName))

	url, err := mayaConfig.GetVolumeURL(versionLatest)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("GET", url.String(), nil)
	c := &http.Client{
		Timeout: timeout,
	}
	resp, err := c.Do(req)
	if err != nil {
		glog.Errorf("Error when connecting to maya-apiserver %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	code := resp.StatusCode
	if code != http.StatusOK {
		glog.Errorf("HTTP Status error from maya-apiserver: %v\n", http.StatusText(code))
		return nil, err
	}
	glog.V(2).Info("volume Details Successfully Retrieved")

	// Fill the obtained json into volume
	json.NewDecoder(resp.Body).Decode(volume)

	return &volume, nil
}
