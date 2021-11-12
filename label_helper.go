package main

import (
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

type iRepoLabelHelper interface {
	getLabelsOfRepo() ([]string, error)
	isCollaborator(string) (bool, error)
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

type labelHelper interface {
	addLabel([]string) error
	removeLabel([]string) error
	getCurrentLabels() sets.String
	addComment(string) error

	iRepoLabelHelper
}

type issueLabelHelper struct {
	*repoLabelHelper

	number string
	labels sets.String
}

func (h *issueLabelHelper) addLabel(label []string) error {
	return h.cli.AddMultiIssueLabel(h.org, h.repo, h.number, label)
}

func (h *issueLabelHelper) removeLabel(label []string) error {
	return nil
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

func (h *prLabelHelper) addLabel(label []string) error {
	return h.cli.AddMultiPRLabel(h.org, h.repo, h.number, label)
}

func (h *prLabelHelper) removeLabel(label []string) error {
	return h.cli.RemovePRLabels(h.org, h.repo, h.number, label)
}

func (h *prLabelHelper) getCurrentLabels() sets.String {
	return h.labels
}

func (h *prLabelHelper) addComment(comment string) error {
	return h.cli.CreatePRComment(h.org, h.repo, h.number, comment)
}

func getIntersection(a sets.String, b []string) []string {
	return a.Intersection(sets.NewString(b...)).UnsortedList()
}

type setsHelper struct {
	m map[string]string
	s sets.String
}

func (h *setsHelper) intersection(h1 *setsHelper) []string {
	return h.s.Intersection(h1.s).UnsortedList()
}

func (h *setsHelper) difference(h1 *setsHelper) []string {
	return h.s.Difference(h1.s).UnsortedList()
}

func (h *setsHelper) origin(data []string) []string {
	r := make([]string, 0, len(data))
	for _, item := range data {
		if v, ok := h.m[item]; ok {
			r = append(r, v)
		}
	}
	return r
}

func (h *setsHelper) count() int {
	return len(h.m)
}

func (h *setsHelper) toList() []string {
	return h.s.UnsortedList()
}

func (h *setsHelper) differenceSlice(data []string) []string {
	return h.s.Difference(sets.NewString(data...)).UnsortedList()
}

func newSetsHelper(data []string) *setsHelper {
	m := map[string]string{}
	v := make([]string, len(data))
	for i := range data {
		v[i] = strings.ToLower(data[i])
		m[v[i]] = data[i]
	}

	return &setsHelper{
		m: m,
		s: sets.NewString(v...),
	}
}
