package main

import (
	"fmt"
	"strings"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	libconfig "github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
)

func (bot *robot) handleClearLabel(e *sdk.PullRequestEvent, pc libconfig.PluginConfig, log *logrus.Entry) error {
	if giteeclient.GetPullRequestAction(e) != giteeclient.PRActionChangedSourceBranch {
		return nil
	}

	prInfo := giteeclient.GetPRInfoByPREvent(e)

	cfg, err := bot.getConfig(pc, prInfo.Org, prInfo.Repo)
	if err != nil {
		return err
	}
	if len(cfg.ClearLabels) == 0 {
		return nil
	}

	toRemove := getIntersection(prInfo.Labels, cfg.ClearLabels)
	if len(toRemove) == 0 {
		return nil
	}

	err = bot.cli.RemovePRLabels(prInfo.Org, prInfo.Repo, prInfo.Number, toRemove)
	if err != nil {
		return err
	}

	comment := fmt.Sprintf(
		"This pull request source branch has changed, so removes the following label(s): %s.",
		strings.Join(toRemove, ", "),
	)

	return bot.cli.CreatePRComment(prInfo.Org, prInfo.Repo, prInfo.Number, comment)
}
