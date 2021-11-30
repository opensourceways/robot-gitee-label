package main

import (
	"github.com/opensourceways/community-robot-lib/utils"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

type iRepoLabelHelper interface {
	getLabelsOfRepo() ([]string, error)
	isCollaborator(string) (bool, error)
	createLabelsOfRepo(missing []string) error
}

type repoLabelHelper struct {
	cli  iClient
	org  string
	repo string
}

func (h *repoLabelHelper) isCollaborator(commenter string) (bool, error) {
	return h.cli.IsCollaborator(h.org, h.repo, commenter)
}

func (h *repoLabelHelper) getLabelsOfRepo() ([]string, error) {
	labels, err := h.cli.GetRepoLabels(h.org, h.repo)
	if err != nil {
		return nil, err
	}

	r := make([]string, len(labels))
	for i, item := range labels {
		r[i] = item.Name
	}
	return r, nil
}

func (h *repoLabelHelper) createLabelsOfRepo(labels []string) error {
	mErr := utils.MultiError{}
	for _, v := range labels {
		mErr.AddError(h.cli.CreateRepoLabel(h.org, h.repo, v, ""))
	}
	return mErr.Err()
}

type labelHelper interface {
	addLabels([]string) error
	removeLabels([]string) error
	getCurrentLabels() sets.String
	addComment(string) error

	iRepoLabelHelper
}

type issueLabelHelper struct {
	*repoLabelHelper

	number string
	labels sets.String
}

func (h *issueLabelHelper) addLabels(label []string) error {
	return h.cli.AddMultiIssueLabel(h.org, h.repo, h.number, label)
}

func (h *issueLabelHelper) removeLabels(label []string) error {
	return h.cli.RemoveIssueLabels(h.org, h.repo, h.number, label)
}

func (h *issueLabelHelper) getCurrentLabels() sets.String {
	return h.labels
}

func (h *issueLabelHelper) addComment(comment string) error {
	return h.cli.CreateIssueComment(h.org, h.repo, h.number, comment)
}

type prLabelHelper struct {
	*repoLabelHelper

	number int32
	labels sets.String
}

func (h *prLabelHelper) addLabels(label []string) error {
	return h.cli.AddMultiPRLabel(h.org, h.repo, h.number, label)
}

func (h *prLabelHelper) removeLabels(label []string) error {
	return h.cli.RemovePRLabels(h.org, h.repo, h.number, label)
}

func (h *prLabelHelper) getCurrentLabels() sets.String {
	return h.labels
}

func (h *prLabelHelper) addComment(comment string) error {
	return h.cli.CreatePRComment(h.org, h.repo, h.number, comment)
}

type labelSet struct {
	m map[string]string
	s sets.String
}

func (ls *labelSet) count() int {
	return len(ls.m)
}

func (ls *labelSet) toList() []string {
	return ls.s.UnsortedList()
}

func (ls *labelSet) origin(data []string) []string {
	r := make([]string, 0, len(data))
	for _, item := range data {
		if v, ok := ls.m[item]; ok {
			r = append(r, v)
		}
	}
	return r
}

func (ls *labelSet) intersection(ls1 *labelSet) []string {
	return ls.s.Intersection(ls1.s).UnsortedList()
}

func (ls *labelSet) difference(ls1 *labelSet) []string {
	return ls.s.Difference(ls1.s).UnsortedList()
}

func newLabelSet(data []string) *labelSet {
	m := map[string]string{}
	v := make([]string, len(data))
	for i := range data {
		v[i] = strings.ToLower(data[i])
		m[v[i]] = data[i]
	}

	return &labelSet{
		m: m,
		s: sets.NewString(v...),
	}
}
