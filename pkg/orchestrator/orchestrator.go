package orchestrator

import (
	"errors"
	"github.com/SAP/jenkins-library/pkg/log"
	"os"
	"time"
)

type Orchestrator int

const (
	Unknown Orchestrator = iota
	AzureDevOps
	GitHubActions
	Jenkins
)

type OrchestratorSpecificConfigProviding interface {
	InitOrchestratorProvider(settings *OrchestratorSettings)
	OrchestratorType() string
	OrchestratorVersion() string
	GetStageName() string
	GetBranch() string
	GetBuildUrl() string
	GetBuildId() string
	GetJobUrl() string
	GetJobName() string
	GetCommit() string
	GetPullRequestConfig() PullRequestConfig
	GetRepoUrl() string
	IsPullRequest() bool
	GetLog() ([]byte, error)
	GetPipelineStartTime() time.Time
	GetBuildStatus() string
}

type PullRequestConfig struct {
	Branch string
	Base   string
	Key    string
}

type OrchestratorSettings struct {
	JenkinsUser  string
	JenkinsToken string
	AzureToken   string
}

func NewOrchestratorSpecificConfigProvider() (OrchestratorSpecificConfigProviding, error) {
	switch DetectOrchestrator() {
	case AzureDevOps:
		return &AzureDevOpsConfigProvider{}, nil
	case GitHubActions:
		return &GitHubActionsConfigProvider{}, nil
	case Jenkins:
		return &JenkinsConfigProvider{}, nil
	default:
		return &UnknownOrchestratorConfigProvider{}, errors.New("unable to detect a supported orchestrator (Azure DevOps, GitHub Actions, Jenkins)")
	}
}

func DetectOrchestrator() Orchestrator {
	if isAzure() {
		return Orchestrator(AzureDevOps)
	} else if isGitHubActions() {
		return Orchestrator(GitHubActions)
	} else if isJenkins() {
		return Orchestrator(Jenkins)
	} else {
		return Orchestrator(Unknown)
	}
}

func (o Orchestrator) String() string {
	return [...]string{"Unknown", "AzureDevOps", "GitHubActions", "Jenkins"}[o]
}

func areIndicatingEnvVarsSet(envVars []string) bool {
	for _, v := range envVars {
		if truthy(v) {
			return true
		}
	}
	return false
}

// Checks if var is set and neither empty nor false
func truthy(key string) bool {
	val, exists := os.LookupEnv(key)
	if !exists {
		return false
	}
	if len(val) == 0 || val == "no" || val == "false" || val == "off" || val == "0" {
		return false
	}

	return true
}

// Wrapper function to read env variable and set default value
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	log.Entry().Warnf("Could not read env variable %v using fallback value %v", key, fallback)
	return fallback
}
