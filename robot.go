package main

import (
	"fmt"

	sdk "gitee.com/openeuler/go-gitee/gitee"
	libconfig "github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	libplugin "github.com/opensourceways/community-robot-lib/giteeplugin"
	"github.com/sirupsen/logrus"
)

const botName = "label"

type iClient interface {
	GetRepoLabels(owner, repo string) ([]sdk.Label, error)
	GetIssueLabels(org, repo, number string) ([]sdk.Label, error)
	GetPRLabels(org, repo string, number int32) ([]sdk.Label, error)

	AddIssueLabel(org, repo, number, label string) error
	RemoveIssueLabel(org, repo, number, label string) error

	AddMultiIssueLabel(org, repo, number string, label []string) error
	AddMultiPRLabel(org, repo string, number int32, label []string) error
	RemovePRLabel(org, repo string, number int32, label string) error

	CreatePRComment(org, repo string, number int32, comment string) error
	CreateIssueComment(org, repo string, number string, comment string) error
	ListPROperationLogs(org, repo string, number int32) ([]sdk.OperateLog, error)

	IsCollaborator(owner, repo, login string) (bool, error)
}

func newRobot(cli iClient) *robot {
	return &robot{cli: cli}
}

type robot struct {
	cli iClient
}

func (bot *robot) NewPluginConfig() libconfig.PluginConfig {
	return &configuration{}
}

func (bot *robot) getConfig(cfg libconfig.PluginConfig) (*configuration, error) {
	if c, ok := cfg.(*configuration); ok {
		return c, nil
	}
	return nil, fmt.Errorf("can't convert to configuration")
}

func (bot *robot) RegisterEventHandler(p libplugin.HandlerRegitster) {
	p.RegisterPullRequestHandler(bot.handlePREvent)
	p.RegisterNoteEventHandler(bot.handleNoteEvent)
}

func (bot *robot) handlePREvent(e *sdk.PullRequestEvent, cfg libconfig.PluginConfig, log *logrus.Entry) error {
	prInfo := giteeclient.GetPRInfoByPREvent(e)

	cfgForRepo := bot.getConfigItem(prInfo.Org, prInfo.Repo, cfg, log)
	if cfgForRepo == nil {
		return nil
	}
	prHandle := &prNoteHandle{client: bot.cli, org: prInfo.Org, repo: prInfo.Repo, number: prInfo.Number}

	switch giteeclient.GetPullRequestAction(e) {
	case giteeclient.PRActionChangedSourceBranch:
		return bot.handleClearLabel(prHandle, cfgForRepo)
	case "edited":
		return bot.handleValidatingLabel(prHandle, prInfo, cfgForRepo, log)
	}

	return nil
}

func (bot *robot) handleNoteEvent(e *sdk.NoteEvent, cfg libconfig.PluginConfig, log *logrus.Entry) error {
	ne := giteeclient.NewNoteEventWrapper(e)
	if !ne.IsCreatingCommentEvent() {
		log.Debug("Event is not a creation of a comment, skipping.")
		return nil
	}

	matchLabels := genMachLabels(ne.GetComment())
	if matchLabels == nil {
		log.Debug("comment content needn't handle, skipping.")
		return nil
	}

	return bot.handleLabels(ne, matchLabels, cfg, log)
}

func (bot *robot) getConfigItem(org, repo string, cfg libconfig.PluginConfig, log *logrus.Entry) *botConfig {
	config, err := bot.getConfig(cfg)
	if err != nil {
		return nil
	}

	cfgItem := config.configFor(org, repo)
	if cfgItem == nil {
		log.Info(fmt.Sprintf("no %s robot config for this repo:%s/%s", botName, org, repo))
		return nil
	}

	return cfgItem
}

func (bot *robot) getRepoLabelsMap(org, repo string) (map[string]string, error) {
	repoLabels, err := bot.cli.GetRepoLabels(org, repo)
	if err != nil {
		return nil, err
	}
	return labelsTransformMap(repoLabels), nil
}
