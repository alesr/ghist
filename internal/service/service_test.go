package service

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCalcDiffs(t *testing.T) {
	givenDBRepos := []GithubRepo{
		{
			Name:  "foo",
			Stars: 1,
			Forks: 2,
		},
		{
			Name:  "bar",
			Stars: 1,
			Forks: 2,
		},
		{
			Name:  "qux",
			Stars: 1,
			Forks: 2,
		},
	}

	givenGhRepos := []GithubRepo{
		{
			Name:  "foo",
			Stars: 2, // +1
			Forks: 2,
		},
		{
			Name:  "bar", // no change
			Stars: 1,
			Forks: 2,
		},
		{
			Name:  "qux",
			Stars: 1,
			Forks: 2,
		},
	}

	got := calcDiffs(givenGhRepos, givenDBRepos)

	expected := []Diff{
		{
			Name:  "foo",
			Stars: 1,
		},
	}

	for _, d := range got {
		fmt.Println(d.String())
	}
	assert.ElementsMatch(t, expected, got)
}

func intsToPtr(i int) *int {
	return &i
}
