package main

import (
	"time"

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

	//ClearLabels specifies labels that should be removed when the codes of PR are changed.
	ClearLabels []string `json:"clear_labels,omitempty"`

	//LabelsToValidate specifies config of label that will be validated
	LabelsToValidate []validateLabelConfig `json:"labels_to_validate,omitempty"`
}

func (c *botConfig) setDefault() {
}

func (c *botConfig) validate() error {
	return c.PluginForRepo.Validate()
}

type validateLabelConfig struct {
	// Label is the label name to be validated
	Label string `json:"label" required:"true"`

	// ActiveTime is the time in hours that the label becomes invalid after it from created
	ActiveTime int `json:"active_time" required:"true"`
}

func (vc validateLabelConfig) isExpiry(t time.Time) bool {
	return t.Add(time.Duration(vc.ActiveTime) * time.Hour).Before(time.Now())
}


