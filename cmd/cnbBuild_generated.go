// Code generated by piper's step-generator. DO NOT EDIT.

package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/SAP/jenkins-library/pkg/config"
	"github.com/SAP/jenkins-library/pkg/log"
	"github.com/SAP/jenkins-library/pkg/piperenv"
	"github.com/SAP/jenkins-library/pkg/splunk"
	"github.com/SAP/jenkins-library/pkg/telemetry"
	"github.com/SAP/jenkins-library/pkg/validation"
	"github.com/spf13/cobra"
)

type cnbBuildOptions struct {
	ContainerImageName        string   `json:"containerImageName,omitempty"`
	ContainerImageTag         string   `json:"containerImageTag,omitempty"`
	ContainerRegistryURL      string   `json:"containerRegistryUrl,omitempty"`
	Buildpacks                []string `json:"buildpacks,omitempty"`
	BuildEnvVars              []string `json:"buildEnvVars,omitempty"`
	Path                      string   `json:"path,omitempty"`
	ProjectDescriptor         string   `json:"projectDescriptor,omitempty"`
	DockerConfigJSON          string   `json:"dockerConfigJSON,omitempty"`
	CustomTLSCertificateLinks []string `json:"customTlsCertificateLinks,omitempty"`
	AdditionalTags            []string `json:"additionalTags,omitempty"`
}

type cnbBuildCommonPipelineEnvironment struct {
	container struct {
		registryURL  string
		imageNameTag string
	}
}

func (p *cnbBuildCommonPipelineEnvironment) persist(path, resourceName string) {
	content := []struct {
		category string
		name     string
		value    interface{}
	}{
		{category: "container", name: "registryUrl", value: p.container.registryURL},
		{category: "container", name: "imageNameTag", value: p.container.imageNameTag},
	}

	errCount := 0
	for _, param := range content {
		err := piperenv.SetResourceParameter(path, resourceName, filepath.Join(param.category, param.name), param.value)
		if err != nil {
			log.Entry().WithError(err).Error("Error persisting piper environment.")
			errCount++
		}
	}
	if errCount > 0 {
		log.Entry().Fatal("failed to persist Piper environment")
	}
}

// CnbBuildCommand Executes a Cloud Native Buildpacks build for creating a Docker container.
func CnbBuildCommand() *cobra.Command {
	const STEP_NAME = "cnbBuild"

	metadata := cnbBuildMetadata()
	var stepConfig cnbBuildOptions
	var startTime time.Time
	var commonPipelineEnvironment cnbBuildCommonPipelineEnvironment
	var logCollector *log.CollectorHook

	var createCnbBuildCmd = &cobra.Command{
		Use:   STEP_NAME,
		Short: "Executes a Cloud Native Buildpacks build for creating a Docker container.",
		Long:  `Executes a Cloud Native Buildpacks build for creating a Docker container.`,
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			startTime = time.Now()
			log.SetStepName(STEP_NAME)
			log.SetVerbose(GeneralConfig.Verbose)

			GeneralConfig.GitHubAccessTokens = ResolveAccessTokens(GeneralConfig.GitHubTokens)

			path, _ := os.Getwd()
			fatalHook := &log.FatalHook{CorrelationID: GeneralConfig.CorrelationID, Path: path}
			log.RegisterHook(fatalHook)

			err := PrepareConfig(cmd, &metadata, STEP_NAME, &stepConfig, config.OpenPiperFile)
			if err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}
			log.RegisterSecret(stepConfig.DockerConfigJSON)

			if len(GeneralConfig.HookConfig.SentryConfig.Dsn) > 0 {
				sentryHook := log.NewSentryHook(GeneralConfig.HookConfig.SentryConfig.Dsn, GeneralConfig.CorrelationID)
				log.RegisterHook(&sentryHook)
			}

			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				logCollector = &log.CollectorHook{CorrelationID: GeneralConfig.CorrelationID}
				log.RegisterHook(logCollector)
			}

			validation, err := validation.New(validation.WithJSONNamesForStructFields(), validation.WithPredefinedErrorMessages())
			if err != nil {
				return err
			}
			if err = validation.ValidateStruct(stepConfig); err != nil {
				log.SetErrorCategory(log.ErrorConfiguration)
				return err
			}

			return nil
		},
		Run: func(_ *cobra.Command, _ []string) {
			telemetryData := telemetry.CustomData{}
			telemetryData.ErrorCode = "1"
			handler := func() {
				config.RemoveVaultSecretFiles()
				commonPipelineEnvironment.persist(GeneralConfig.EnvRootPath, "commonPipelineEnvironment")
				telemetryData.Duration = fmt.Sprintf("%v", time.Since(startTime).Milliseconds())
				telemetryData.ErrorCategory = log.GetErrorCategory().String()
				telemetry.Send(&telemetryData)
				if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
					splunk.Send(&telemetryData, logCollector)
				}
			}
			log.DeferExitHandler(handler)
			defer handler()
			telemetry.Initialize(GeneralConfig.NoTelemetry, STEP_NAME)
			if len(GeneralConfig.HookConfig.SplunkConfig.Dsn) > 0 {
				splunk.Initialize(GeneralConfig.CorrelationID,
					GeneralConfig.HookConfig.SplunkConfig.Dsn,
					GeneralConfig.HookConfig.SplunkConfig.Token,
					GeneralConfig.HookConfig.SplunkConfig.Index,
					GeneralConfig.HookConfig.SplunkConfig.SendLogs)
			}
			cnbBuild(stepConfig, &telemetryData, &commonPipelineEnvironment)
			telemetryData.ErrorCode = "0"
			log.Entry().Info("SUCCESS")
		},
	}

	addCnbBuildFlags(createCnbBuildCmd, &stepConfig)
	return createCnbBuildCmd
}

func addCnbBuildFlags(cmd *cobra.Command, stepConfig *cnbBuildOptions) {
	cmd.Flags().StringVar(&stepConfig.ContainerImageName, "containerImageName", os.Getenv("PIPER_containerImageName"), "Name of the container which will be built")
	cmd.Flags().StringVar(&stepConfig.ContainerImageTag, "containerImageTag", os.Getenv("PIPER_containerImageTag"), "Tag of the container which will be built")
	cmd.Flags().StringVar(&stepConfig.ContainerRegistryURL, "containerRegistryUrl", os.Getenv("PIPER_containerRegistryUrl"), "Container registry where the image should be pushed to")
	cmd.Flags().StringSliceVar(&stepConfig.Buildpacks, "buildpacks", []string{}, "List of custom buildpacks to use in the form of '<hostname>/<repo>[:<tag>]'.")
	cmd.Flags().StringSliceVar(&stepConfig.BuildEnvVars, "buildEnvVars", []string{}, "List of custom environment variables used during a build in the form of 'KEY=VALUE'.")
	cmd.Flags().StringVar(&stepConfig.Path, "path", os.Getenv("PIPER_path"), "The path should either point to a directory with your sources or an artifact in zip format.")
	cmd.Flags().StringVar(&stepConfig.ProjectDescriptor, "projectDescriptor", `project.toml`, "Path to the project.toml file (see https://buildpacks.io/docs/reference/config/project-descriptor/ for the reference). Parameters passed to the cnbBuild step will take precedence over the parameters set in the project.toml file.")
	cmd.Flags().StringVar(&stepConfig.DockerConfigJSON, "dockerConfigJSON", os.Getenv("PIPER_dockerConfigJSON"), "Path to the file `.docker/config.json` - this is typically provided by your CI/CD system. You can find more details about the Docker credentials in the [Docker documentation](https://docs.docker.com/engine/reference/commandline/login/).")
	cmd.Flags().StringSliceVar(&stepConfig.CustomTLSCertificateLinks, "customTlsCertificateLinks", []string{}, "List containing download links of custom TLS certificates. This is required to ensure trusted connections to registries with custom certificates.")
	cmd.Flags().StringSliceVar(&stepConfig.AdditionalTags, "additionalTags", []string{}, "List of tags which will be additionally pushed to the registry, e.g. \"latest\".")

	cmd.MarkFlagRequired("containerImageName")
	cmd.MarkFlagRequired("containerImageTag")
	cmd.MarkFlagRequired("containerRegistryUrl")
	cmd.MarkFlagRequired("dockerConfigJSON")
}

// retrieve step metadata
func cnbBuildMetadata() config.StepData {
	var theMetaData = config.StepData{
		Metadata: config.StepMetadata{
			Name:        "cnbBuild",
			Aliases:     []config.Alias{},
			Description: "Executes a Cloud Native Buildpacks build for creating a Docker container.",
		},
		Spec: config.StepSpec{
			Inputs: config.StepInputs{
				Secrets: []config.StepSecrets{
					{Name: "dockerConfigJsonCredentialsId", Description: "Jenkins 'Secret file' credentials ID containing Docker config.json (with registry credential(s)) in the following format:\n\n```json\n{\n    \"auths\": {\n            \"$server\": {\n                    \"auth\": \"base64($username + ':' + $password)\"\n            }\n    }\n}\n```\n\nExample:\n\n```json\n{\n    \"auths\": {\n            \"example.com\": {\n                    \"auth\": \"dXNlcm5hbWU6cGFzc3dvcmQ=\"\n            }\n    }\n}\n```\n", Type: "jenkins"},
				},
				Parameters: []config.StepParameters{
					{
						Name:        "containerImageName",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   true,
						Aliases:     []config.Alias{{Name: "dockerImageName"}},
						Default:     os.Getenv("PIPER_containerImageName"),
					},
					{
						Name: "containerImageTag",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "artifactVersion",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "artifactVersion"}},
						Default:   os.Getenv("PIPER_containerImageTag"),
					},
					{
						Name: "containerRegistryUrl",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "container/registryUrl",
							},
						},
						Scope:     []string{"GENERAL", "PARAMETERS", "STAGES", "STEPS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{{Name: "dockerRegistryUrl"}},
						Default:   os.Getenv("PIPER_containerRegistryUrl"),
					},
					{
						Name:        "buildpacks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "buildEnvVars",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "path",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     os.Getenv("PIPER_path"),
					},
					{
						Name:        "projectDescriptor",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     `project.toml`,
					},
					{
						Name: "dockerConfigJSON",
						ResourceRef: []config.ResourceReference{
							{
								Name:  "commonPipelineEnvironment",
								Param: "custom/dockerConfigJSON",
							},

							{
								Name: "dockerConfigJsonCredentialsId",
								Type: "secret",
							},

							{
								Name:    "dockerConfigFileVaultSecretName",
								Type:    "vaultSecretFile",
								Default: "docker-config",
							},
						},
						Scope:     []string{"PARAMETERS"},
						Type:      "string",
						Mandatory: true,
						Aliases:   []config.Alias{},
						Default:   os.Getenv("PIPER_dockerConfigJSON"),
					},
					{
						Name:        "customTlsCertificateLinks",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
					{
						Name:        "additionalTags",
						ResourceRef: []config.ResourceReference{},
						Scope:       []string{"PARAMETERS", "STAGES", "STEPS"},
						Type:        "[]string",
						Mandatory:   false,
						Aliases:     []config.Alias{},
						Default:     []string{},
					},
				},
			},
			Containers: []config.Container{
				{Image: "paketobuildpacks/builder:full"},
			},
			Outputs: config.StepOutputs{
				Resources: []config.StepResources{
					{
						Name: "commonPipelineEnvironment",
						Type: "piperEnvironment",
						Parameters: []map[string]interface{}{
							{"Name": "container/registryUrl"},
							{"Name": "container/imageNameTag"},
						},
					},
				},
			},
		},
	}
	return theMetaData
}
