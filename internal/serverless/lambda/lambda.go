package lambda

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/lambda"

	"github.com/benthosdev/benthos/v4/internal/component/output"
	"github.com/benthosdev/benthos/v4/internal/config"
	"github.com/benthosdev/benthos/v4/internal/docs"
	"github.com/benthosdev/benthos/v4/internal/filepath/ifs"
	"github.com/benthosdev/benthos/v4/internal/serverless"
)

var handler *serverless.Handler

// Run executes Benthos as an AWS Lambda function. Configuration can be stored
// within the environment variable BENTHOS_CONFIG.
func Run() {
	// A list of default config paths to check for if not explicitly defined
	defaultPaths := []string{
		"./benthos.yaml",
		"./config.yaml",
		"/benthos.yaml",
		"/etc/benthos/config.yaml",
		"/etc/benthos.yaml",
	}
	if path := os.Getenv("BENTHOS_CONFIG_PATH"); path != "" {
		defaultPaths = append([]string{path}, defaultPaths...)
	}

	conf := config.New()
	conf.Metrics.Type = "none"
	conf.Logger.Format = "json"

	var err error
	if conf.Output, err = output.FromYAML(`
switch:
  retry_until_success: false
  cases:
    - check: 'errored()'
      output:
        reject: "processing failed due to: ${! error() }"
    - output:
        sync_response: {}
`); err != nil {
		fmt.Fprintf(os.Stderr, "Config init error: %v\n", err)
		os.Exit(1)
	}

	if confStr := os.Getenv("BENTHOS_CONFIG"); len(confStr) > 0 {
		confBytes, err := config.ReplaceEnvVariables([]byte(confStr), os.LookupEnv)
		if err != nil {
			// TODO: Make this configurable somehow maybe, along with linting
			// errors.
			var errEnvMissing *config.ErrMissingEnvVars
			if errors.As(err, &errEnvMissing) {
				confBytes = errEnvMissing.BestAttempt
			} else {
				fmt.Fprintf(os.Stderr, "Configuration file read error: %v\n", err)
				os.Exit(1)
			}
		}

		confNode, err := docs.UnmarshalYAML(confBytes)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Configuration file parse error: %v\n", err)
			os.Exit(1)
		}

		if err = conf.FromAny(docs.DeprecatedProvider, confNode); err != nil {
			fmt.Fprintf(os.Stderr, "Configuration file read error: %v\n", err)
			os.Exit(1)
		}
	} else {
		// Iterate default config paths
		for _, path := range defaultPaths {
			if _, err := ifs.OS().Stat(path); err == nil {
				if _, err = config.ReadYAMLFileLinted(ifs.OS(), path, false, docs.NewLintConfig(), &conf); err != nil {
					fmt.Fprintf(os.Stderr, "Configuration file read error: %v\n", err)
					os.Exit(1)
				}
				break
			}
		}
	}

	if handler, err = serverless.NewHandler(conf); err != nil {
		fmt.Fprintf(os.Stderr, "Initialisation error: %v\n", err)
		os.Exit(1)
	}

	lambda.Start(handler.Handle)
	if err = handler.Close(time.Second * 30); err != nil {
		fmt.Fprintf(os.Stderr, "Shut down error: %v\n", err)
		os.Exit(1)
	}
}
