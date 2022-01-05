package mediawiki

import (
	"time"
)

type Editor struct {
	Identifier  int64 `json:",omitempty"`
	IsAnonymous bool  `json:"is_anonymous,omitempty"`
	IsBot       bool  `json:"is_bot,omitempty"`
	Name        string
	EditCount   int64     `json:"edit_count,omitempty"`
	DateStarted time.Time `json:"date_started,omitempty"`
	Groups      []string
}

type Probability struct {
	False float64
	True  float64
}

type Score struct {
	Prediction  bool
	Probability Probability
}

type Scores struct {
	Damaging  Score
	Goodfaith Score
}

type Version struct {
	Identifier      int64
	Editor          Editor
	Comment         string   `json:",omitempty"`
	Tags            []string `json:",omitempty"`
	IsMinorEdit     bool     `json:"is_minor_edit,omitempty"`
	IsFlaggedStable bool     `json:"is_flagged_stable,omitempty"`
	Scores          Scores   `json:",omitempty"`
}

// TODO: Should Type and Level be enumerations?
// TODO: Should Expiry be time.Time?
type Protection struct {
	Type   string
	Level  string
	Expiry string
}

type Namespace struct {
	Identifier int64
	Name       string
}

type InLanguage struct {
	Identifier string
	Name       string
}

type EntityRef struct {
	Identifier string
	URL        string   `json:"url"`
	Aspects    []string `json:",omitempty"`
}

type Category struct {
	Name string
	URL  string `json:"url"`
}

type Template struct {
	Name string
	URL  string `json:"url"`
}

type Redirect struct {
	Name string
	URL  string `json:"url"`
}

type IsPartOf struct {
	Identifier string
	Name       string
}

type ArticleBody struct {
	HTML     string `json:"html"`
	WikiText string `json:"wikitext"`
}

type License struct {
	Identifier string
	Name       string
	URL        string `json:"url"`
}

type Article struct {
	Name               string
	Identifier         int64
	DateModified       time.Time    `json:"date_modified"`
	Protection         []Protection `json:",omitempty"`
	Version            Version
	URL                string `json:"url"`
	Namespace          Namespace
	InLanguage         InLanguage  `json:"in_language"`
	MainEntity         EntityRef   `json:"main_entity"`
	AdditionalEntities []EntityRef `json:"additional_entities,omitempty"`
	Categories         []Category  `json:",omitempty"`
	Templates          []Template  `json:",omitempty"`
	Redirects          []Redirect  `json:",omitempty"`
	IsPartOf           IsPartOf    `json:"is_part_of"`
	ArticleBody        ArticleBody `json:"article_body"`
	License            []License
}
