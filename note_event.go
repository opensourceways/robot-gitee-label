package main

import (
	"fmt"
	"strings"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
)

func (bot *robot) handleLabels(
	e giteeclient.NoteEventWrapper,
	toAdd []string,
	toRemove []string,
	cfg *botConfig,
	log *logrus.Entry,
) error {

	lh := genLabelHelper(e, bot.cli)
	if lh == nil {
		return nil
	}

	if v := getIntersectionBetweenSlices(toAdd, toRemove); len(v) > 0 {
		return lh.addComment(fmt.Sprintf(
			"conflict labels(%s) exit", strings.Join(v, ", "),
		))
	}

	merr := utils.NewMultiErrors()

	if len(toRemove) > 0 {
		if _, err := removeLabels(lh, toRemove); err != nil {
			merr.AddError(err)
		}
	}

	if len(toAdd) > 0 {
		err := addLabels(lh, toAdd, e.GetCommenter(), cfg, log)
		if err != nil {
			merr.AddError(err)
		}
	}
	return merr.Err()
}

func genLabelHelper(e giteeclient.NoteEventWrapper, cli iClient) labelHelper {
	org, repo := e.GetOrgRep()
	rlh := &repoLabelHelper{
		cli:  cli,
		org:  org,
		repo: repo,
	}

	if e.IsPullRequest() {
		ne := giteeclient.NewPRNoteEvent(e.NoteEvent)
		return &prLabelHelper{
			number:          ne.GetPRNumber(),
			labels:          nil,
			repoLabelHelper: rlh,
		}

	}

	if e.IsIssue() {
		ne := giteeclient.NewIssueNoteEvent(e.NoteEvent)
		return &issueLabelHelper{
			number:          ne.GetIssueNumber(),
			labels:          nil,
			repoLabelHelper: rlh,
		}
	}
	return nil
}

func addLabels(h labelHelper, toAdd []string, commenter string, cfg *botConfig, log *logrus.Entry) error {
	labels, err := h.getCurrentLabels()
	if err != nil {
		return err
	}

	ls := getDifference(toAdd, labels)
	if len(ls) == 0 {
		return nil
	}

	canAdd, missing, err := getLabelsCanAdd(h, ls, commenter, cfg, log)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()

	if len(canAdd) > 0 {
		if err := h.addLabel(canAdd); err != nil {
			merr.AddError(err)
		}
	}

	if len(missing) > 0 {
		msg := fmt.Sprintf(
			"The label(s) `%s` cannot be applied, because the repository doesn't have them",
			strings.Join(missing, ", "),
		)

		if err := h.addComment(msg); err != nil {
			merr.AddError(err)
		}
	}

	return merr.Err()
}

func getLabelsCanAdd(
	h labelHelper,
	toAdd []string,
	commenter string,
	cfg *botConfig,
	log *logrus.Entry,
) ([]string, []string, error) {

	repoLabels, err := h.getLabelsOfRepo()
	if err != nil {
		return nil, nil, err
	}

	missing := getDifference(toAdd, repoLabels)
	if len(missing) == 0 {
		return toAdd, nil, nil
	}

	if !cfg.AllowCreatingLabelsByCollaborator {
		return getIntersection(repoLabels, toAdd), missing, nil
	}

	v, err := h.isCollaborator(commenter)
	if v {
		return toAdd, nil, nil
	}
	if err != nil {
		log.WithError(err).Error("check whether is collaborator")
	}
	return getIntersection(repoLabels, toAdd), missing, nil
}

func removeLabels(lh labelHelper, toRemove []string) (removed []string, err error) {
	labels, err := lh.getCurrentLabels()
	if err != nil {
		return
	}

	ls := getIntersection(labels, toRemove)
	if len(ls) == 0 {
		return nil, nil
	}

	return ls, lh.removeLabel(ls)
}
