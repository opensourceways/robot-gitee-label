package main

import (
	"fmt"
	"strings"

	"github.com/opensourceways/community-robot-lib/giteeclient"
	"github.com/opensourceways/community-robot-lib/utils"
	"github.com/sirupsen/logrus"
	"k8s.io/apimachinery/pkg/util/sets"
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

	add := newSetsHelper(toAdd)
	remove := newSetsHelper(toRemove)
	if v := add.intersection(remove); len(v) > 0 {
		return lh.addComment(fmt.Sprintf(
			"conflict labels(%s) exit", strings.Join(add.origin(v), ", "),
		))
	}

	merr := utils.NewMultiErrors()

	if remove.count() > 0 {
		if _, err := removeLabels(lh, remove); err != nil {
			merr.AddError(err)
		}
	}

	if add.count() > 0 {
		err := addLabels(lh, add, e.GetCommenter(), cfg)
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
			labels:          ne.GetPRLabels(),
			repoLabelHelper: rlh,
		}
	}

	if e.IsIssue() {
		ne := giteeclient.NewIssueNoteEvent(e.NoteEvent)
		return &issueLabelHelper{
			number:          ne.GetIssueNumber(),
			labels:          ne.GetIssueLabels(),
			repoLabelHelper: rlh,
		}
	}
	return nil
}

func addLabels(lh labelHelper, toAdd *setsHelper, commenter string, cfg *botConfig) error {
	canAdd, missing, err := checkLabesToAdd(lh, toAdd, commenter, cfg)
	if err != nil {
		return err
	}

	merr := utils.NewMultiErrors()

	if len(canAdd) > 0 {
		ls := sets.NewString(canAdd...).Difference(lh.getCurrentLabels())
		if ls.Len() > 0 {
			if err := lh.addLabel(ls.UnsortedList()); err != nil {
				merr.AddError(err)
			}
		}
	}

	if len(missing) > 0 {
		msg := fmt.Sprintf(
			"The label(s) `%s` cannot be applied, because the repository doesn't have them",
			strings.Join(missing, ", "),
		)

		if err := lh.addComment(msg); err != nil {
			merr.AddError(err)
		}
	}

	return merr.Err()
}

func checkLabesToAdd(
	h labelHelper,
	toAdd *setsHelper,
	commenter string,
	cfg *botConfig,
) ([]string, []string, error) {

	v, err := h.getLabelsOfRepo()
	if err != nil {
		return nil, nil, err
	}
	repoLabels := newSetsHelper(v)

	missing := toAdd.difference(repoLabels)
	if len(missing) == 0 {
		return repoLabels.origin(toAdd.toList()), nil, nil
	}

	var canAdd []string
	if len(missing) < toAdd.count() {
		canAdd = repoLabels.origin(toAdd.differenceSlice(missing))
	}

	missing = toAdd.origin(missing)

	if !cfg.AllowCreatingLabelsByCollaborator {
		return canAdd, missing, nil
	}

	b, err := h.isCollaborator(commenter)
	if err != nil {
		return nil, nil, err
	}
	if b {
		return append(canAdd, missing...), nil, nil
	}
	return canAdd, missing, nil
}

func removeLabels(lh labelHelper, toRemove *setsHelper) ([]string, error) {
	v, err := lh.getLabelsOfRepo()
	if err != nil {
		return nil, err
	}
	repoLabels := newSetsHelper(v)

	labels := repoLabels.intersection(toRemove)
	if len(labels) == 0 {
		return nil, nil
	}

	ls := lh.getCurrentLabels().Intersection(
		sets.NewString(repoLabels.origin(labels)...)).UnsortedList()

	if len(ls) == 0 {
		return nil, nil
	}
	return ls, lh.removeLabel(ls)
}
