package service

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

type ghClient interface {
	FetchRepos(ctx context.Context, username string) ([]GithubRepo, error)
}

type repo interface {
	GetRepositories() ([]GithubRepo, error)
	UpsertRepositories(ghRepos []GithubRepo) error
}

type GithubRepo struct {
	Name  string `json:"name"`
	Stars int    `json:"stargazers_count"`
	Forks int    `json:"forks_count"`
}

type Diff struct {
	Name  string
	Stars int
	Forks int
}

func (d *Diff) String() string {
	var parts []string
	if d.Stars > 0 {
		parts = append(parts, fmt.Sprintf("%s: +%d stars", d.Name, d.Stars))
	}

	if d.Forks > 0 {
		parts = append(parts, fmt.Sprintf("+%d forks", d.Forks))
	}
	return strings.Join(parts, ", ")
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

	ghRepos, err := s.ghClient.FetchRepos(context.TODO(), ghUsername)
	if err != nil {
		return nil, fmt.Errorf("could not get repos from GitHub: %w", err)
	}

	diffs := calcDiffs(ghRepos, dbRepos)

	if err := s.repo.UpsertRepositories(ghRepos); err != nil {
		return nil, fmt.Errorf("could not upsert repos in db: %w", err)
	}
	return diffs, nil
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
			if ghRepo.Stars > 0 || ghRepo.Forks > 0 {
				diffs = append(diffs, Diff{
					Name:  ghRepo.Name,
					Stars: ghRepo.Stars,
					Forks: ghRepo.Forks,
				})
			}
		} else {
			starsDiff := ghRepo.Stars - dbRepo.Stars
			forksDiff := ghRepo.Forks - dbRepo.Forks

			if starsDiff > 0 || forksDiff > 0 {
				diffs = append(diffs, Diff{
					Name:  ghRepo.Name,
					Stars: starsDiff,
					Forks: forksDiff,
				})
			}
		}
	}
	return diffs
}
