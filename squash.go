package main

import sdk "github.com/opensourceways/go-gitee/gitee"

func (bot *robot) handleSquashLabel(e *sdk.PullRequestEvent, commits uint, cfg SquashConfig) error {
	if cfg.unableCheckingSquash() {
		return nil
	}

	action := sdk.GetPullRequestAction(e)
	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()
	labels := e.GetPRLabelSet()

	if action != sdk.PRActionOpened && action != sdk.PRActionChangedSourceBranch {
		return nil
	}

	exceeded := commits > cfg.CommitsThreshold
	hasSquashLabel := labels.Has(cfg.SquashCommitLabel)

	if exceeded && !hasSquashLabel {
		return bot.cli.AddPRLabel(org, repo, number, cfg.SquashCommitLabel)
	}

	if !exceeded && hasSquashLabel {
		return bot.cli.RemovePRLabel(org, repo, number, cfg.SquashCommitLabel)
	}

	return nil
}
