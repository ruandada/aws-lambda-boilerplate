package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/ruandada/aws-lambda-boilerplate/aws-lambda-golang-docker-starter/internal/lambdaapp"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "bootstrap-cli",
		Short: "AWS Lambda Go bootstrap CLI",
		Long:  "A Go + Docker AWS Lambda bootstrap CLI with local debug commands.",
		// Default to Lambda runtime to fit container startup.
		RunE: func(_ *cobra.Command, _ []string) error {
			startLambdaRuntime()
			return nil
		},
	}

	rootCmd.AddCommand(newDevCmd())
	rootCmd.AddCommand(newTestEventCmd())
	rootCmd.AddCommand(newLambdaCmd())

	return rootCmd
}

func newDevCmd() *cobra.Command {
	var port int
	cmd := &cobra.Command{
		Use:   "dev",
		Short: "Start local HTTP server",
		RunE: func(_ *cobra.Command, _ []string) error {
			return lambdaapp.StartLocalServer(port)
		},
	}

	defaultPort := 3000
	if envPort, ok := os.LookupEnv("PORT"); ok {
		parsed, err := strconv.Atoi(envPort)
		if err == nil && parsed > 0 {
			defaultPort = parsed
		}
	}
	cmd.Flags().IntVarP(&port, "port", "p", defaultPort, "local HTTP port")
	return cmd
}

func newTestEventCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "test-event <event-json-file>",
		Short: "Invoke Lambda handler with local event file",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			eventPayload, err := loadEventFile(args[0])
			if err != nil {
				return err
			}

			result, err := lambdaapp.HandleRawEvent(context.Background(), eventPayload)
			if err != nil {
				return err
			}

			if result == nil {
				fmt.Println("Handler completed with no return value.")
				return nil
			}

			body, err := json.MarshalIndent(result, "", "  ")
			if err != nil {
				fmt.Println(result)
				return nil
			}

			fmt.Println(string(body))
			return nil
		},
	}
	return cmd
}

func newLambdaCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "lambda",
		Short: "Start AWS Lambda runtime",
		RunE: func(_ *cobra.Command, _ []string) error {
			startLambdaRuntime()
			return nil
		},
	}
}

func startLambdaRuntime() {
	lambda.Start(lambda.NewHandler(lambdaapp.HandleRawEvent))
}

func loadEventFile(path string) ([]byte, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read event file %q: %w", path, err)
	}

	if len(content) == 0 {
		return nil, errors.New("event file is empty")
	}

	return content, nil
}
