package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
	"k8s.io/kubernetes/pkg/util/interrupt"

	"github.com/jcleira/kubernetes-test-106741/dep"
)

type params struct {
	namespace string
	pod       string
	container string
}

func main() {
	var (
		namespace = flag.String("namespace", "default", "namespace")
		pod       = flag.String("pod", "", "pod")
		container = flag.String("container", "", "container")
	)
	t := params{
		namespace: *namespace,
		pod:       *pod,
		container: *container,
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()

	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		log.Fatal(err)
	}

	groupversion := schema.GroupVersion{
		Group:   "",
		Version: "v1",
	}

	config.GroupVersion = &groupversion
	config.APIPath = "/api"
	config.ContentType = runtime.ContentTypeJSON
	config.NegotiatedSerializer = dep.BasicNegotiatedSerializer{}

	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatal(err)
	}

	fn := func() error {
		req := restclient.Post().
			Resource("pods").
			Name(t.pod).
			Namespace(t.namespace).
			SubResource("log").
			Param("pretty", "true")

		req.VersionedParams(
			&v1.PodLogOptions{
				Container: t.container,
			},
			scheme.ParameterCodec,
		)
		executor, err := remotecommand.NewSPDYExecutor(
			config, http.MethodGet, req.URL(),
		)
		if err != nil {
			return err
		}
		err = executor.Stream(remotecommand.StreamOptions{
			Stdin:  os.Stdin,
			Stdout: os.Stdout,
			Stderr: os.Stderr,
			Tty:    true,
		})

		fmt.Println(err)
		return err
	}

	interrupt.Chain(nil, func() {
		fmt.Println("done!")
	}).Run(fn)
}
