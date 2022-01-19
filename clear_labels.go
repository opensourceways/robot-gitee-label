package main

import (
	"fmt"
	"strings"

	sdk "github.com/opensourceways/go-gitee/gitee"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleClearLabel(e *sdk.PullRequestEvent, cfg *botConfig) error {
	action := sdk.GetPullRequestAction(e)
	if action != sdk.PRActionChangedSourceBranch {
		return nil
	}

	labels := e.GetPRLabelSet()
	toRemove := getClearLabels(labels, cfg)
	if len(toRemove) == 0 {
		return nil
	}

	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()

	if err := bot.cli.RemovePRLabels(org, repo, number, toRemove); err != nil {
		return err
	}

	comment := fmt.Sprintf(
		"This pull request source branch has changed, so removes the following label(s): %s.",
		strings.Join(toRemove, ", "),
	)

	return bot.cli.CreatePRComment(org, repo, number, comment)
}

func getClearLabels(labels sets.String, cfg *botConfig) []string {
	var r []string

	all := labels
	if len(cfg.ClearLabels) > 0 {
		v := all.Intersection(sets.NewString(cfg.ClearLabels...))
		if v.Len() > 0 {
			r = v.UnsortedList()
			all = all.Difference(v)
		}
	}

	exp := cfg.clearLabelsByRegexp
	if exp != nil {
		for k := range all {
			if exp.MatchString(k) {
				r = append(r, k)
			}
		}
	}

	return r
}
