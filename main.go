package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	coreV1 "k8s.io/api/core/v1"
	metaV1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	coreV1Types "k8s.io/client-go/kubernetes/typed/core/v1"
	"sigs.k8s.io/yaml"

	_ "k8s.io/client-go/plugin/pkg/client/auth"
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

var secretsClient coreV1Types.SecretInterface

func main() {

	secretfile := flag.NewFlagSet("local", flag.ExitOnError)
	fname := secretfile.String("f", "", "specify file name with -f")

	k8ssecname := flag.NewFlagSet("k8s", flag.ExitOnError)
	secname := k8ssecname.String("s", "", "secret name with -s")
	namespace := k8ssecname.String("n", "", "namespace with -n")
	kubeconfig := k8ssecname.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "(optional) absolute path to the kubeconfig file")

	if len(os.Args) < 2 {
		fmt.Println("expected 'k8s', 'local' and 'read' subcommands")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "local":
		secretfile.Parse(os.Args[2:])
		if *fname == "" {
			secretfile.Usage()
			os.Exit(1)
		}
		fromSecretFile(*fname)
	case "k8s":
		k8ssecname.Parse(os.Args[2:])
		if *secname == "" || *namespace == "" {
			k8ssecname.Usage()
			os.Exit(1)
		}
		fromKubeSecret(*secname, *namespace, *kubeconfig)
	case "read":
		fromStdInput()
	default:
		fmt.Println("expected 'k8s', 'local' and 'read' subcommands")
		os.Exit(1)
	}
}

func fromSecretFile(name string) {

	bytes, err := ioutil.ReadFile(name)
	if err != nil {
		panic(err.Error())
	}

	var secretSpec coreV1.Secret
	err = yaml.Unmarshal(bytes, &secretSpec)
	if err != nil {
		panic(err.Error())
	}

	secretName := secretSpec.ObjectMeta.Name
	secretData := secretSpec.Data

	fmt.Println("secretName:", secretName)
	for key, value := range secretData {
		fmt.Printf("%s=%s\n", key, value)
	}

}

func fromKubeSecret(secname, namespace, kubeconfig string) *coreV1.Secret {

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}
	clientset, err := kubernetes.NewForConfig(config)

	if err != nil {
		panic(err.Error())
	}

	secretsClient = clientset.CoreV1().Secrets(namespace)

	secret, err := secretsClient.Get(context.TODO(), secname, metaV1.GetOptions{})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("data: ")

	for key, value := range secret.Data {
		fmt.Printf(" %s=%s\n", key, value)
	}
	return secret

}

func fromStdInput() {
	var out []byte
	scanner := bufio.NewReader(os.Stdin)
	for {
		input, err := scanner.ReadByte()
		if err != nil && err == io.EOF {
			break
		}
		out = append(out, input)
	}

	var secretSpec coreV1.Secret
	err := yaml.Unmarshal(out, &secretSpec)
	if err != nil {
		panic(err.Error())
	}

	secretName := secretSpec.ObjectMeta.Name
	secretData := secretSpec.Data

	fmt.Println("secretName:", secretName)
	for key, value := range secretData {
		fmt.Printf("%s=%s\n", key, value)
	}

}
