package main

import (
	"fmt"

	"github.com/opensourceways/community-robot-lib/config"
	framework "github.com/opensourceways/community-robot-lib/robot-gitee-framework"
	"github.com/opensourceways/community-robot-lib/utils"
	sdk "github.com/opensourceways/go-gitee/gitee"
	"github.com/sirupsen/logrus"
)

const botName = "label"

type iClient interface {
	GetRepoLabels(owner, repo string) ([]sdk.Label, error)
	GetIssueLabels(org, repo, number string) ([]sdk.Label, error)
	GetPRLabels(org, repo string, number int32) ([]sdk.Label, error)

	AddIssueLabel(org, repo, number, label string) error
	RemoveIssueLabels(org, repo, number string, label []string) error

	AddMultiIssueLabel(org, repo, number string, label []string) error
	AddMultiPRLabel(org, repo string, number int32, label []string) error
	AddPRLabel(org, repo string, number int32, label string) error
	RemovePRLabel(org, repo string, number int32, label string) error
	RemovePRLabels(org, repo string, number int32, labels []string) error
	CreateRepoLabel(org, repo, label, color string) error

	CreatePRComment(org, repo string, number int32, comment string) error
	CreateIssueComment(org, repo string, number string, comment string) error

	IsCollaborator(owner, repo, login string) (bool, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewConfig() config.Config {
	return &configuration{}
}

func (bot *robot) getConfig(cfg config.Config, org, repo string) (*botConfig, error) {
	c, ok := cfg.(*configuration)
	if !ok {
		return nil, fmt.Errorf("can't convert to configuration")
	}

	if bc := c.configFor(org, repo); bc != nil {
		return bc, nil
	}

	return nil, fmt.Errorf("no config for this repo:%s/%s", org, repo)
}

func (bot *robot) RegisterEventHandler(p framework.HandlerRegitster) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
	p.RegisterNoteEventHandler(bot.handleNoteEvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, pc config.Config, log *logrus.Entry) error {
	org, repo := e.GetOrgRepo()

	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()
	if err = bot.handleClearLabel(e, cfg); err != nil {
		merr.AddError(err)
	}

	commits := uint(e.GetPullRequest().GetCommits())

	err = bot.handleSquashLabel(e, commits, cfg.SquashConfig)
	if err != nil {
		merr.AddError(err)
	}

	return merr.Err()
}

func (bot *robot) handleNoteEvent(e *sdk.NoteEvent, pc config.Config, log *logrus.Entry) error {
	if !e.IsCreatingCommentEvent() {
		log.Debug("Event is not a creation of a comment, skipping.")
		return nil
	}

	org, repo := e.GetOrgRepo()
	cfg, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	toAdd, toRemove := getMatchedLabels(e.GetCommenter())
	if len(toAdd) == 0 && len(toRemove) == 0 {
		log.Debug("invalid comment, skipping.")
		return nil
	}

	return bot.handleLabels(e, toAdd, toRemove, cfg, log)
}
