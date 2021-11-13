package main

import (
	"fmt"
	"strings"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"k8s.io/apimachinery/pkg/util/sets"
)

func (bot *robot) handleClearLabel(action string, prInfo giteeclient.PRInfo, cfg *botConfig) error {
	if action != giteeclient.PRActionChangedSourceBranch {
		return nil
	}

	toRemove := getClearLabels(prInfo.Labels, cfg)
	if len(toRemove) == 0 {
		return nil
	}

	err := bot.cli.RemovePRLabels(prInfo.Org, prInfo.Repo, prInfo.Number, toRemove)
	if err != nil {
		return err
	}

	comment := fmt.Sprintf(
		"This pull request source branch has changed, so removes the following label(s): %s.",
		strings.Join(toRemove, ", "),
	)

	return bot.cli.CreatePRComment(prInfo.Org, prInfo.Repo, prInfo.Number, comment)
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
