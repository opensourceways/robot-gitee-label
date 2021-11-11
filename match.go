package main

import (
	"regexp"
	"strings"
)

var (
	labelRegex       = regexp.MustCompile(`(?m)^/(kind|priority|sig)\s*(.*?)\s*$`)
	removeLabelRegex = regexp.MustCompile(`(?m)^/remove-(kind|priority|sig)\s*(.*?)\s*$`)
)

type match struct {
	adds    [][]string
	removes [][]string
}

func (m match) getAddLabels() []string {
	return getLabelsFromMatch(m.adds)
}

func (m match) getRemoveLabels() []string {
	return getLabelsFromMatch(m.removes)
}

func getLabelsFromMatch(lm [][]string) []string {
	var labels []string

	for _, v := range lm {
		for _, label := range strings.Split(v[0], " ")[1:] {
			label = strings.ToLower(v[1] + "/" + strings.TrimSpace(label))
			labels = append(labels, label)
		}
	}

	return labels
}

func genMachLabels(comment string) *match {
	addLabelMatches := labelRegex.FindAllStringSubmatch(comment, -1)
	removeLabelMatches := removeLabelRegex.FindAllStringSubmatch(comment, -1)

	if len(addLabelMatches) == 0 && len(removeLabelMatches) == 0 {
		return nil
	}

	return &match{
		adds:    addLabelMatches,
		removes: removeLabelMatches,
	}
}
