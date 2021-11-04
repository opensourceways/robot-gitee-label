package main

import (
	"fmt"
	"strings"

	libconfig "github.com/opensourceways/community-robot-lib/config"
	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/sirupsen/logrus"
)

func (bot *robot) handleLabels(e giteeclient.NoteEventWrapper, mLabels *match, pc libconfig.PluginConfig, log *logrus.Entry) error {
	org, repo := e.GetOrgRep()
	_, err := bot.getConfig(pc, org, repo)
	if err != nil {
		return err
	}

	var nh noteHandler
	if e.IsPullRequest() {
		number := giteeclient.NewPRNoteEvent(e.NoteEvent).GetPRNumber()
		nh = &prNoteHandle{org: org, repo: repo, client: bot.cli, number: number}
	} else if e.IsIssue() {
		number := giteeclient.NewIssueNoteEvent(e.NoteEvent).GetIssueNumber()
		nh = &issueNoteHandle{client: bot.cli, org: org, repo: repo, number: number}
	} else {
		return nil
	}

	removeLabels := mLabels.getRemoveLabels()
	if len(removeLabels) > 0 {
		if _, err := handleRemoveLabels(nh, removeLabels); err != nil {
			log.Error(err)
		}
	}

	addLabels := mLabels.getAddLabels()
	if len(addLabels) <= 0 {
		return nil
	}

	repoLabels, err := bot.getRepoLabelsMap(org, repo)
	if err != nil {
		return err
	}
	isCollaborator, _ := bot.cli.IsCollaborator(org, repo, e.GetCommenter())

	return handleAddLabels(nh, repoLabels, addLabels, isCollaborator)
}

func (bot *robot) handleClearLabel(handle *prNoteHandle, cfg *botConfig) error {
	labels := cfg.ClearLabels
	if len(labels) == 0 {
		return nil
	}

	removed, err := handleRemoveLabels(handle, labels)
	if err != nil || len(removed) == 0 {
		return err
	}

	comment := fmt.Sprintf(
		"This pull request source branch has changed, label(s): %s has been removed.", strings.Join(removed, ","))

	return handle.addComment(comment)
}

func handleAddLabels(nh noteHandler, repoLabels map[string]string, labelsToAdd []string, isCollaborator bool) error {
	noteLabels, err := nh.getLabels()
	if err != nil {
		return err
	}

	var canNotAddLabels []string
	var canAddLabels []string

	for _, labelToAdd := range labelsToAdd {
		if _, ok := noteLabels[labelToAdd]; ok {
			continue
		}
		if _, ok := repoLabels[labelToAdd]; !ok && !isCollaborator {
			canNotAddLabels = append(canNotAddLabels, labelToAdd)
		} else {
			canAddLabels = append(canAddLabels, labelToAdd)
		}
	}

	if len(canAddLabels) > 0 {
		if err := nh.addLabel(canAddLabels); err != nil {
			return err
		}
	}

	if len(canNotAddLabels) > 0 {
		msg := fmt.Sprintf(
			"The label(s) `%s` cannot be applied, because the repository doesn't have them",
			strings.Join(canNotAddLabels, ", "),
		)
		return nh.addComment(msg)
	}

	return nil
}

func handleRemoveLabels(nh noteHandler, removeLabels []string) (removed []string, err error) {
	eventSubjectLabels, err := nh.getLabels()
	if err != nil {
		return
	}

	var rmFails []string
	for _, rmLabel := range removeLabels {
		if label, ok := eventSubjectLabels[rmLabel]; ok {
			if err := nh.removeLabel(label); err != nil {
				rmFails = append(rmFails, fmt.Sprintf("%s labe remove fiald with err: %s", label, err.Error()))
				continue
			}
			removed = append(removed, label)
		}
	}

	if len(rmFails) > 0 {
		err = fmt.Errorf("failed remove some labels, details : %s", strings.Join(rmFails, ";"))
	}
	return
}
