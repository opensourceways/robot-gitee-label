package main

import "github.com/opensourceways/community-robot-lib/giteeclient"

func (bot *robot) handleSquashLabel(action string, prInfo giteeclient.PRInfo, commits uint, cfg SquashConfig) error {
	if !cfg.needCheckCommits() {
		return nil
	}

	if action != giteeclient.PRActionOpened && action != giteeclient.PRActionChangedSourceBranch {
		return nil
	}

	exceeded := commits > cfg.CommitsThreshold
	hasSquashLabel := prInfo.HasLabel(cfg.SquashCommitLabel)
	org, repo, number := prInfo.Org, prInfo.Repo, prInfo.Number

	if exceeded && !hasSquashLabel {
		return bot.cli.AddPRLabel(org, repo, number, cfg.SquashCommitLabel)
	}

	if !exceeded && hasSquashLabel {
		return bot.cli.RemovePRLabel(org, repo, number, cfg.SquashCommitLabel)
	}

	return nil
}
