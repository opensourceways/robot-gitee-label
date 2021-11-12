package main

import "k8s.io/apimachinery/pkg/util/sets"

type iRepoLabelHelper interface {
	getLabelsOfRepo() (sets.String, error)
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

func (h *repoLabelHelper) getLabelsOfRepo() (sets.String, error) {
	labels, err := h.cli.GetRepoLabels(h.org, h.repo)
	if err != nil {
		return nil, err
	}

	r := sets.NewString()
	for _, item := range labels {
		r.Insert(item.Name)
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

func getIntersectionBetweenSlices(a, b []string) []string {
	return sets.NewString(a...).Intersection(
		sets.NewString(b...)).UnsortedList()
}

func getDifference(a []string, b sets.String) []string {
	return sets.NewString(a...).Difference(b).UnsortedList()
}
