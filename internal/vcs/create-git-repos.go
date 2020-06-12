package vcs

import (
	"context"
	"fmt"
	"github.com/machinebox/graphql"
	"os/exec"
	"strings"
)

// InitializeRepository Creates and initializes a github repository for the given url
// repositoryUrl is expected to be in the format "github.com/{ownerName}/{repositoryName}"
func InitializeRepository(repositoryUrl string, githubApiKey string) {

	ownerName, repositoryName, pErr := parseRepositoryUrl(repositoryUrl)
	if pErr != nil {
		fmt.Printf("error creating repository: %s\n", pErr.Error())
		return
	}

	isOrgOwned, ownerId, iErr := isOrganizationOwned(ownerName, githubApiKey)
	if iErr != nil {
		fmt.Printf("error creating repository: %s\n", iErr.Error())
		return
	}

	if isOrgOwned {
		r := graphql.NewRequest(createOrganizationRepositoryMutation)
		r.Var("repoName", repositoryName)
		r.Var("repoDescription", fmt.Sprintf("Repository for %s", repositoryName))
		r.Var("ownerId", ownerId)
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubApiKey))

		if err := createRepository(r, repositoryName); err != nil {
			fmt.Printf("error creating repository: %s\n", err.Error())
			return
		}
	} else {
		r := graphql.NewRequest(createPersonalRepositoryMutation)
		r.Var("repoName", repositoryName)
		r.Var("repoDescription", fmt.Sprintf("Repository for %s", repositoryName))
		r.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubApiKey))

		if err := createRepository(r, repositoryName); err != nil {
			fmt.Printf("error creating repository: %s\n", err.Error())
			return
		}
	}

	if err := doInitialCommit(ownerName, repositoryName); err != nil {
		fmt.Printf("error initializing repository: %s\n", err.Error())
	}
}

// parseRepositoryUrl extracts the owner name and repository name from a repository url.
// repositoryUrl is expected to be in the format "github.com/{ownerName}/{repositoryName}"
func parseRepositoryUrl(repositoryUrl string) (string, string, error) {
	if len(repositoryUrl) == 0 {
		return "","", fmt.Errorf("invalid repository url.  expected format \"github.com/{ownerName}/{repositoryName}\"")
	}

	segments := strings.Split(repositoryUrl, "/")
	if len(segments) != 3 {
		return "","", fmt.Errorf("invalid repository url.  expected format \"github.com/{ownerName}/{repositoryName}\"")
	}

	ownerName := segments[1]
	repositoryName := segments[2]

	fmt.Printf("found owner %s, repository %s\n", ownerName, repositoryName)

	return ownerName, repositoryName, nil
}

const createPersonalRepositoryMutation = `mutation ($repoName: String!, $repoDescription: String!) {
		createRepository(
			input: {
				name:$repoName, 
				visibility: PRIVATE, 
				description: $repoDescription
			}) 
		{
			clientMutationId
		}
	}`

const createOrganizationRepositoryMutation = `mutation ($repoName: String!, $repoDescription: String!, $ownerId: String!) {
		createRepository(
			input: {
				name:$repoName, 
				visibility: PRIVATE, 
				description: $repoDescription
				ownerId: $ownerId
			}) 
		{
			clientMutationId
		}
	}`

// createRepository will create a new repository in github
func createRepository(request *graphql.Request, repositoryName string) error {
	fmt.Printf("Creating repository for module: %s\n", repositoryName)

	c := graphql.NewClient("https://api.github.com/graphql")
	ctx := context.Background()
	if err := c.Run(ctx, request, nil); err != nil {
		return err
	}

	fmt.Printf("Repository successfully created for module: %s\n", repositoryName)

	return nil
}

const getOrganizationQuery = `query ($ownerName: String!) {
		organization(login: $ownerName) {
			id
		}
	}`

type organizationQueryResponse struct {
	Organization struct {
		Id string
	}
}

// isOrganizationOwned will determine if ownerName is an organization.
// If ownerName is an organization it's id will be returned.
func isOrganizationOwned(ownerName string, githubApiKey string) (bool, string, error) {
	oRequest := graphql.NewRequest(getOrganizationQuery)
	oRequest.Var("ownerName", ownerName)
	oRequest.Header.Add("Authorization", fmt.Sprintf("Bearer %s", githubApiKey))

	var oResponse organizationQueryResponse
	c := graphql.NewClient("https://api.github.com/graphql")
	ctx := context.Background()
	if err := c.Run(ctx, oRequest, &oResponse); err != nil {

		notAnOrgMessage := fmt.Sprintf("graphql: Could not resolve to an Organization with the login of '%s'.", ownerName)
		if err.Error() == notAnOrgMessage {
			return false, "", nil
		}
		return false, "", err
	}
	organizationId := oResponse.Organization.Id

	return true, organizationId, nil
}

type initialCommands struct {
	description string
	command     string
	args        []string
}

// doInitialCommit runs the git commands that initialize and do the first commit to a repository.
func doInitialCommit(ownerName string, repositoryName string) error {
	fmt.Printf("Initializing repository for module: %s\n", repositoryName)

	remoteOrigin := fmt.Sprintf("git@github.com:%s/%s.git", ownerName, repositoryName)
	fmt.Printf("remote origin: %s\n", remoteOrigin)
	commands := []initialCommands{
		{
			description: "git init",
			command:     "git",
			args:        []string{"init"},
		},
		{
			description: "git add .",
			command:     "git",
			args:        []string{"add", "."},
		},
		{
			description: "git commit -m \"initial commit by zero\"",
			command:     "git",
			args:        []string{"commit", "-m", "initial commit by zero"},
		},
		{
			description: fmt.Sprintf("git remote add origin %s", remoteOrigin),
			command:     "git",
			args:        []string{"remote", "add", "origin", remoteOrigin},
		},
		{
			description: "git push -u origin master",
			command:     "git",
			args:        []string{"push", "-u", "origin", "master"},
		},
	}

	for _, command := range commands {
		fmt.Printf(">> %s\n", command.description)

		cmd := exec.Command(command.command, command.args...)
		cmd.Dir = "./" + repositoryName
		out, err := cmd.CombinedOutput()
		if err != nil {
			fmt.Printf("ERROR: failed to run %s: %s\n", command.description, err.Error())
			// this is a partial failure.  some commands may have exec'ed successfully.
			break
		} else {
			response := string(out)
			if len(response) > 0 {
				fmt.Println(response)
			}
		}
	}

	fmt.Printf("Repository successfully initialized for module: %s\n", repositoryName)

	return nil
}
