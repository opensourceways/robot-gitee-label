package main

import sdk "github.com/opensourceways/go-gitee/gitee"

func (bot *robot) handleSquashLabel(e *sdk.PullRequestEvent, commits uint, cfg SquashConfig) error {
	if cfg.unableCheckingSquash() {
		return nil
	}

	action := sdk.GetPullRequestAction(e)
	if action != sdk.ActionOpen && action != sdk.PRActionChangedSourceBranch {
		return nil
	}

	labels := e.GetPRLabelSet()
	hasSquashLabel := labels.Has(cfg.SquashCommitLabel)
	exceeded := commits > cfg.CommitsThreshold
	org, repo := e.GetOrgRepo()
	number := e.GetPRNumber()

	if exceeded && !hasSquashLabel {
		return bot.cli.AddPRLabel(org, repo, number, cfg.SquashCommitLabel)
	}

	if !exceeded && hasSquashLabel {
		return bot.cli.RemovePRLabel(org, repo, number, cfg.SquashCommitLabel)
	}

	return nil
}
