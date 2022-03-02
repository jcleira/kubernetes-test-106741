package main

import (
	"log"
	"net/http"

	"github.com/moby/term"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/kubernetes/scheme"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
)

func main() {
	config, err := clientcmd.RESTConfigFromKubeConfig([]byte(db.GetKubeConfig(env, region)))
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
	restclient, err := rest.RESTClientFor(config)
	if err != nil {
		log.Fatal(err)
	}

	fn := func() error {
		req := restclient.Get().
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
		return executor.Stream(remotecommand.StreamOptions{
			Stdin:             t,
			Stdout:            t,
			Stderr:            t,
			Tty:               true,
			TerminalSizeQueue: t,
		})
	}

	inFd, _ := term.GetFdInfo(t.conn)
	state, err := term.SaveState(inFd)
	return interrupt.Chain(nil, func() {
		term.RestoreTerminal(inFd, state)
	}).Run(fn)
}
