package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/massdriver-cloud/massdriver-cli/pkg/bundle"
	"github.com/massdriver-cloud/massdriver-cli/pkg/generator"
	"github.com/massdriver-cloud/massdriver-cli/pkg/provisioners/terraform"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
)

var bundleCmd = &cobra.Command{
	Use:   "bundle",
	Short: "Bundle development tools",
	Long:  ``,
}

var bundleBuildCmd = &cobra.Command{
	Use:   "build [Path to bundle.yaml]",
	Short: "Builds bundle JSON Schemas",
	Args:  cobra.ExactArgs(1),
	RunE:  runBundleBuild,
}

var bundleGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a new bundle",
	RunE:  runBundleGenerate,
}

var bundlePublishCmd = &cobra.Command{
	Use:          "publish [Path to bundle.yaml]",
	Short:        "Publish a bundle to Massdriver",
	Args:         cobra.ExactArgs(1),
	RunE:         runBundlePublish,
	SilenceUsage: true,
}

func init() {
	rootCmd.AddCommand(bundleCmd)

	bundleCmd.AddCommand(bundleBuildCmd)
	bundleBuildCmd.Flags().StringP("output", "o", "", "Path to output directory (default is bundle.yaml directory)")

	bundleCmd.AddCommand(bundleGenerateCmd)
	bundleGenerateCmd.Flags().StringP("template-dir", "t", "./generators/xo-bundle-template", "Path to template directory")
	bundleGenerateCmd.Flags().StringP("bundle-dir", "b", "./bundles", "Path to bundle directory")

	bundleCmd.AddCommand(bundlePublishCmd)
	bundlePublishCmd.Flags().StringP("api-key", "k", os.Getenv("MASSDRIVER_API_KEY"), "Massdriver API key (can also be set via MASSDRIVER_API_KEY environment variable)")
}

func runBundleBuild(cmd *cobra.Command, args []string) error {
	var err error
	bundlePath := args[0]

	// default the output to the path of the bundle.yaml file
	output, err := cmd.Flags().GetString("output")
	if err != nil {
		log.Error().Err(err).Str("bundle", bundlePath).Msg("an error occurred while building bundle")
		return err
	}
	if output == "" {
		output = filepath.Dir(bundlePath)
	}

	log.Info().Str("bundle", bundlePath).Msg("building bundle")

	bundle, err := bundle.ParseBundle(bundlePath)
	if err != nil {
		log.Error().Err(err).Str("bundle", bundlePath).Msg("an error occurred while building bundle")
		return err
	}

	err = bundle.GenerateSchemas(output)
	if err != nil {
		log.Error().Err(err).Str("bundle", bundlePath).Msg("an error occurred while generating bundle schema files")
		return err
	}

	for _, step := range bundle.Steps {
		switch step.Provisioner {
		case "terraform":
			err = terraform.GenerateFiles(output, step.Path)
			if err != nil {
				log.Error().Err(err).Str("bundle", bundlePath).Str("provisioner", step.Provisioner).Msg("an error occurred while generating provisioner files")
				return err
			}
		case "exec":
			// No-op (Golang doesn't not fallthrough unless explicitly stated)
		default:
			log.Error().Str("bundle", bundlePath).Msg("unknown provisioner: " + step.Provisioner)
			return fmt.Errorf("unknown provisioner: %v", step.Provisioner)
		}
	}

	log.Info().Str("bundle", bundlePath).Str("output", output).Msg("bundle built")

	return err
}

func runBundleGenerate(cmd *cobra.Command, args []string) error {
	var err error

	bundleDir, err := cmd.Flags().GetString("bundle-dir")
	if err != nil {
		return err
	}

	templateDir, err := cmd.Flags().GetString("template-dir")
	if err != nil {
		return err
	}

	templateData := &generator.TemplateData{
		BundleDir:   bundleDir,
		TemplateDir: templateDir,
	}

	err = generator.RunPrompt(templateData)
	if err != nil {
		return err
	}

	err = generator.Generate(*templateData)
	if err != nil {
		return err
	}

	return nil
}

func runBundlePublish(cmd *cobra.Command, args []string) error {
	var err error
	bundlePath := args[0]

	apiKey, err := cmd.Flags().GetString("api-key")
	if err != nil {
		return err
	}
	if apiKey == "" {
		fmt.Print("Error: Massdriver API Key must be specified\n\n")
		cmd.Usage()
		os.Exit(1)
	}

	b, err := bundle.ParseBundle(bundlePath)
	if err != nil {
		return err
	}

	uploadURL, err := b.Publish(apiKey)
	if err != nil {
		return err
	}

	var buf bytes.Buffer
	err = bundle.TarGzipBundle(bundlePath, &buf)
	if err != nil {
		return err
	}

	err = bundle.UploadToPresignedS3URL(uploadURL, &buf)
	if err != nil {
		return err
	}

	return nil
}
