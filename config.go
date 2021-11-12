package main

import (
	"regexp"

	libconfig "github.com/opensourceways/community-robot-lib/config"
)

type configuration struct {
	ConfigItems []botConfig `json:"config_items,omitempty"`
}

func (c *configuration) configFor(org, repo string) *botConfig {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	v := make([]libconfig.IPluginForRepo, len(items))

	for i := range items {
		v[i] = &items[i]
	}

	if i := libconfig.FindConfig(org, repo, v); i >= 0 {
		return &items[i]
	}

	return nil
}

func (c *configuration) Validate() error {
	if c == nil {
		return nil
	}

	items := c.ConfigItems
	for i := range items {
		if err := items[i].validate(); err != nil {
			return err
		}
	}

	return nil
}

func (c *configuration) SetDefault() {
	if c == nil {
		return
	}

	Items := c.ConfigItems
	for i := range Items {
		Items[i].setDefault()
	}
}

type botConfig struct {
	libconfig.PluginForRepo

	// ClearLabels specifies labels that should be removed when the codes of PR are changed.
	ClearLabels []string `json:"clear_labels,omitempty"`

	// ClearLabelsByRegexp specifies a expression which can match a list of labels that
	// should be removed when the codes of PR are changed.
	ClearLabelsByRegexp string `json:"clear_labels_by_regexp,omitempty"`
	clearLabelsByRegexp *regexp.Regexp

	// AllowCreatingLabelsByCollaborator is a tag which will lead to create unavailable labels
	// by collaborator if it is true.
	AllowCreatingLabelsByCollaborator bool `json:"allow_creating_labels_by_collaborator,omitempty"`
}

func (c *botConfig) setDefault() {
}

func (c *botConfig) validate() error {
	if c.ClearLabelsByRegexp != "" {
		v, err := regexp.Compile(c.ClearLabelsByRegexp)
		if err != nil {
			return err
		}
		c.clearLabelsByRegexp = v
	}
	return c.PluginForRepo.Validate()
}
