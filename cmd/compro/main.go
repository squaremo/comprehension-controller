package main

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"

	generate "github.com/squaremo/comprehension-controller/api/v1alpha1"
	"github.com/squaremo/comprehension-controller/internal/eval"
)

type opts struct {
	filename string
}

const longDesc = `
Command-line tool that evaluates Comprehension objects.
`

func main() {
	opts := &opts{}

	cmd := &cobra.Command{
		Use:   "compro",
		Short: "Comprehension generator",
		Long:  longDesc,
		Example: `
# Run the comprehension specified in file.yaml and print the results as YAML
compro -f file.yaml
`,
		RunE: opts.runE,
	}

	cmd.Flags().StringVarP(&opts.filename, "file", "f", "-", "the path to a file containing a Comprehension object specification")

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error: %s\n", err.Error())
		os.Exit(1)
	}
}

func (o *opts) runE(cmd *cobra.Command, args []string) error {
	var input []byte
	var err error
	if o.filename == "-" {
		input, err = io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
	} else {
		input, err = os.ReadFile(o.filename)
		if err != nil {
			return err
		}
	}

	// TODO: load this "properly", with the schema and kind and all
	// that, through some k8s.io library.
	var compro generate.Comprehension
	if err := yaml.Unmarshal(input, &compro); err != nil {
		return err
	}

	// TODO: supply a Kubernetes client with the get-it-from-the-env
	// constructor.
	ev := eval.Evaluator{}
	outs, err := ev.Eval(&compro.Spec)
	if err != nil {
		return err
	}
	for i := range outs {
		println("---")
		bs, err := yaml.Marshal(outs[i])
		if err != nil {
			return err
		}
		print(string(bs))
	}

	return nil
}
