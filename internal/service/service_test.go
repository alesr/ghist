package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcDiffs(t *testing.T) {
	givenDBRepos := []GithubRepo{
		{
			Name:     "foo",
			Stars:    1,
			Forks:    2,
			Watchers: 3,
		},
		{
			Name:     "bar",
			Stars:    1,
			Forks:    2,
			Watchers: 3,
		},
		{
			Name:     "qux",
			Stars:    1,
			Forks:    2,
			Watchers: 3,
		},
	}

	givenGhRepos := []GithubRepo{
		{
			Name:     "foo",
			Stars:    2,
			Forks:    2,
			Watchers: 3,
		},
		{
			Name:     "bar",
			Stars:    1,
			Forks:    2,
			Watchers: 3,
		},
		{
			Name:     "qux",
			Stars:    1,
			Forks:    2,
			Watchers: 2,
		},
	}

	got := calcDiffs(givenGhRepos, givenDBRepos)

	expected := []Diff{
		{
			Name:     "foo",
			Stars:    intsToPtr(1),
			Forks:    intsToPtr(0),
			Watchers: intsToPtr(0),
		},
		{
			Name:     "qux",
			Stars:    intsToPtr(0),
			Forks:    intsToPtr(0),
			Watchers: intsToPtr(-1),
		},
	}

	assert.ElementsMatch(t, expected, got)
}

func intsToPtr(i int) *int {
	return &i
}
