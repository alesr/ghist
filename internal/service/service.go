package service

import (
	"context"
	"fmt"
	"log/slog"
)

type ghClient interface {
	FetchRepos(ctx context.Context, username string) ([]GithubRepo, error)
}

type repo interface {
	GetRepositories() ([]GithubRepo, error)
	UpsertRepositories(ghRepos []GithubRepo) error
}

type GithubRepo struct {
	Name     string `json:"name"`
	Stars    int    `json:"stargazers_count"`
	Forks    int    `json:"forks_count"`
	Watchers int    `json:"watchers_count"`
}

type service struct {
	logger   *slog.Logger
	repo     repo
	ghClient ghClient
}

func New(logger *slog.Logger, ghCli ghClient, repo repo) *service {
	return &service{
		logger:   logger.WithGroup("ghist-service"),
		repo:     repo,
		ghClient: ghCli,
	}
}

func (s *service) Run(ctx context.Context, ghUsername string) ([]Diff, error) {
	dbRepos, err := s.repo.GetRepositories()
	if err != nil {
		return nil, fmt.Errorf("could not get repos from db: %w", err)
	}

	if err := s.repo.UpsertRepositories(dbRepos); err != nil {
		return nil, fmt.Errorf("could not upsert repos in db: %w", err)
	}

	ghRepos, err := s.ghClient.FetchRepos(context.TODO(), "alesr")
	if err != nil {
		return nil, fmt.Errorf("could not get repos from Github: %w", err)
	}
	return calcDiffs(ghRepos, dbRepos), nil
}

type Diff struct {
	Name     string
	Stars    *int
	Forks    *int
	Watchers *int
}

func calcDiffs(ghRepos []GithubRepo, dbRepos []GithubRepo) []Diff {
	var diffs []Diff

	for _, ghRepo := range ghRepos {
		var dbRepo *GithubRepo
		for i, r := range dbRepos {
			if r.Name == ghRepo.Name {
				dbRepo = &dbRepos[i]
				break
			}
		}

		if dbRepo == nil {
			diffs = append(diffs, Diff{
				Name:     ghRepo.Name,
				Stars:    &ghRepo.Stars,
				Forks:    &ghRepo.Forks,
				Watchers: &ghRepo.Watchers,
			})
		} else {
			starsDiff := ghRepo.Stars - dbRepo.Stars
			forksDiff := ghRepo.Forks - dbRepo.Forks
			watchersDiff := ghRepo.Watchers - dbRepo.Watchers

			if starsDiff != 0 || forksDiff != 0 || watchersDiff != 0 {
				diffs = append(diffs, Diff{
					Name:     ghRepo.Name,
					Stars:    pointerOrNil(starsDiff),
					Forks:    pointerOrNil(forksDiff),
					Watchers: pointerOrNil(watchersDiff),
				})
			}
		}
	}
	return diffs
}

var zero = 0

func pointerOrNil(i int) *int {
	if i == 0 {
		return &zero
	}
	return &i
}
