// Copyright 2018 The Terraformer Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package github

import (
	"context"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/Perruer/unclick/terraformutils"
	githubAPI "github.com/google/go-github/v35/github"
)

type RepositoriesGenerator struct {
	GithubService
}

// Generate TerraformResources from github API,
func (g *RepositoriesGenerator) InitResources() error {
	ctx := context.Background()
	client, err := g.createClient()
	if err != nil {
		return err
	}

	owner := g.GetArgs()["owner"].(string)
	list, err := repositoryLister(ctx, client, owner)
	if err != nil {
		log.Println(err)
		return nil
	}
	for page := 1; page != 0; {
		repos, resp, err := list(page)
		if err != nil {
			log.Println(err)
			return nil
		}
		for _, repo := range repos {
			resource := terraformutils.NewSimpleResource(
				repo.GetName(),
				repo.GetName(),
				"github_repository",
				"github",
				[]string{},
			)
			resource.SlowQueryRequired = true
			g.Resources = append(g.Resources, resource)
			g.Resources = append(g.Resources, g.createRepositoryWebhookResources(ctx, client, repo)...)
			g.Resources = append(g.Resources, g.createRepositoryBranchProtectionResources(ctx, client, repo)...)
			g.Resources = append(g.Resources, g.createRepositoryCollaboratorResources(ctx, client, repo)...)
			g.Resources = append(g.Resources, g.createRepositoryDeployKeyResources(ctx, client, repo)...)
		}
		page = resp.NextPage
	}

	return nil
}

func (g *RepositoriesGenerator) createRepositoryWebhookResources(ctx context.Context, client *githubAPI.Client, repo *githubAPI.Repository) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	hooks, _, err := client.Repositories.ListHooks(ctx, g.GetArgs()["owner"].(string), repo.GetName(), nil)
	if err != nil {
		log.Println(err)
	}
	for _, hook := range hooks {
		resources = append(resources, terraformutils.NewResource(
			strconv.FormatInt(hook.GetID(), 10),
			repo.GetName()+"_"+strconv.FormatInt(hook.GetID(), 10),
			"github_repository_webhook",
			"github",
			map[string]string{
				"repository": repo.GetName(),
			},
			[]string{},
			map[string]interface{}{},
		))
	}
	return resources
}

func (g *RepositoriesGenerator) createRepositoryBranchProtectionResources(ctx context.Context, client *githubAPI.Client, repo *githubAPI.Repository) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	branches, _, err := client.Repositories.ListBranches(ctx, g.GetArgs()["owner"].(string), repo.GetName(), nil)
	if err != nil {
		log.Println(err)
	}
	for _, branch := range branches {
		if branch.GetProtected() {
			resources = append(resources, terraformutils.NewSimpleResource(
				repo.GetName()+":"+branch.GetName(),
				repo.GetName()+"_"+branch.GetName(),
				"github_branch_protection",
				"github",
				[]string{},
			))
		}
	}
	return resources
}

func (g *RepositoriesGenerator) createRepositoryCollaboratorResources(ctx context.Context, client *githubAPI.Client, repo *githubAPI.Repository) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	collaborators, _, err := client.Repositories.ListCollaborators(ctx, g.GetArgs()["owner"].(string), repo.GetName(), nil)
	if err != nil {
		log.Println(err)
	}
	for _, collaborator := range collaborators {
		resources = append(resources, terraformutils.NewSimpleResource(
			repo.GetName()+":"+collaborator.GetLogin(),
			repo.GetName()+":"+collaborator.GetLogin(),
			"github_repository_collaborator",
			"github",
			[]string{},
		))
	}
	return resources
}

func (g *RepositoriesGenerator) createRepositoryDeployKeyResources(ctx context.Context, client *githubAPI.Client, repo *githubAPI.Repository) []terraformutils.Resource {
	resources := []terraformutils.Resource{}
	deployKeys, _, err := client.Repositories.ListKeys(ctx, g.GetArgs()["owner"].(string), repo.GetName(), nil)
	if err != nil {
		log.Println(err)
	}
	for _, key := range deployKeys {
		resources = append(resources, terraformutils.NewSimpleResource(
			repo.GetName()+":"+strconv.FormatInt(key.GetID(), 10),
			repo.GetName()+":"+key.GetTitle(),
			"github_repository_deploy_key",
			"github",
			[]string{},
		))
	}
	return resources
}

// PostGenerateHook for connect between resources
func (g *RepositoriesGenerator) PostConvertHook() error {
	for _, repo := range g.Resources {
		if repo.InstanceInfo.Type != "github_repository" {
			continue
		}
		for i, member := range g.Resources {
			if member.InstanceInfo.Type != "github_repository_webhook" {
				continue
			}
			if member.InstanceState.Attributes["repository"] == repo.InstanceState.Attributes["name"] {
				g.Resources[i].Item["repository"] = "${github_repository." + repo.ResourceName + ".name}"
			}
		}
		for i, branch := range g.Resources {
			if branch.InstanceInfo.Type != "github_branch_protection" {
				continue
			}
			if branch.InstanceState.Attributes["repository"] == repo.InstanceState.Attributes["name"] {
				g.Resources[i].Item["repository"] = "${github_repository." + repo.ResourceName + ".name}"
			}
		}
		for i, collaborator := range g.Resources {
			if collaborator.InstanceInfo.Type != "github_repository_collaborator" {
				continue
			}
			if collaborator.InstanceState.Attributes["repository"] == repo.InstanceState.Attributes["name"] {
				g.Resources[i].Item["repository"] = "${github_repository." + repo.ResourceName + ".name}"
			}
		}
		for i, key := range g.Resources {
			if key.InstanceInfo.Type != "github_repository_deploy_key" {
				continue
			}
			if key.InstanceState.Attributes["repository"] == repo.InstanceState.Attributes["name"] {
				g.Resources[i].Item["repository"] = "${github_repository." + repo.ResourceName + ".name}"
			}
		}
	}
	return nil
}

// repositoryLister returns a function that lists one page of the owner's
// repositories. The owner may be an organization or a personal account;
// Terraformer only handled organizations. For the account the token belongs
// to, private repositories are listed too.
func repositoryLister(ctx context.Context, client *githubAPI.Client, owner string) (func(page int) ([]*githubAPI.Repository, *githubAPI.Response, error), error) {
	byOrg := func(page int) ([]*githubAPI.Repository, *githubAPI.Response, error) {
		return client.Repositories.ListByOrg(ctx, owner, &githubAPI.RepositoryListByOrgOptions{
			ListOptions: githubAPI.ListOptions{PerPage: 100, Page: page},
		})
	}
	if _, _, err := client.Organizations.Get(ctx, owner); err == nil {
		return byOrg, nil
	} else if resp, ok := err.(*githubAPI.ErrorResponse); !ok || resp.Response == nil || resp.Response.StatusCode != http.StatusNotFound {
		return nil, err
	}

	user := ""
	if me, _, err := client.Users.Get(ctx, ""); err == nil && strings.EqualFold(me.GetLogin(), owner) {
		log.Printf("%s is the authenticated user: listing its private repositories too", owner)
	} else {
		user = owner
	}
	return func(page int) ([]*githubAPI.Repository, *githubAPI.Response, error) {
		opt := &githubAPI.RepositoryListOptions{ListOptions: githubAPI.ListOptions{PerPage: 100, Page: page}}
		if user == "" {
			opt.Affiliation = "owner"
		} else {
			opt.Type = "owner"
		}
		return client.Repositories.List(ctx, user, opt)
	}, nil
}
