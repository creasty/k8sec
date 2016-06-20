package command

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/dtan4/k8sec/k8s"
	"k8s.io/kubernetes/pkg/api"
	client "k8s.io/kubernetes/pkg/client/unversioned"
)

type LoadCommand struct {
	Meta
}

func (c *LoadCommand) Run(args []string) int {
	var (
		arguments  []string
		filename   string
		kubeconfig string
		kubeClient *client.Client
		namespace  string
	)

	flags := flag.NewFlagSet("list", flag.ContinueOnError)
	flags.Usage = func() {}

	flags.StringVar(&filename, "f", "", "Path to save (Default: flush to stdout)")
	flags.StringVar(&kubeconfig, "kubeconfig", "", "Path to the kubeconfig file (Default: ~/.kube/config)")
	flags.StringVar(&namespace, "namespace", "", "Namespace scope (Default: default)")

	if err := flags.Parse(args[0:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for 0 < flags.NArg() {
		arguments = append(arguments, flags.Arg(0))
		flags.Parse(flags.Args()[1:])
	}

	if len(arguments) != 1 {
		fmt.Fprintln(os.Stderr, "Variable name must be specified.")
		return 1
	}

	name := arguments[0]

	if namespace == "" {
		namespace = api.NamespaceDefault
	}

	data := map[string][]byte{}

	if filename != "" {
		f, err := os.Open(filename)

		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return 1
		}

		defer f.Close()

		sc := bufio.NewScanner(f)

		for sc.Scan() {
			line := sc.Text()
			ary := strings.SplitN(line, "=", 2)

			if len(ary) != 2 {
				fmt.Fprintln(os.Stderr, "Line should be key=value format. Given line: "+line)
				return 1
			}

			k, v := ary[0], ary[1]

			_v, err := strconv.Unquote(v)

			if err != nil {
				// Parse as is
				_v = v
			}

			data[k] = []byte(_v)
		}
	}

	kubeClient, err := k8s.NewKubeClient(kubeconfig)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	s, err := kubeClient.Secrets(namespace).Get(name)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	for k, v := range data {
		s.Data[k] = v
	}

	_, err = kubeClient.Secrets(namespace).Update(s)

	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	return 1
}

func (c *LoadCommand) Synopsis() string {
	return ""
}

func (c *LoadCommand) Help() string {
	helpText := `

`
	return strings.TrimSpace(helpText)
}
