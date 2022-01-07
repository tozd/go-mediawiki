package mediawiki

import (
	"time"
)

type Editor struct {
	Identifier  int64      `json:"identifier,omitempty"`
	IsAnonymous bool       `json:"is_anonymous,omitempty"`
	IsBot       bool       `json:"is_bot,omitempty"`
	Name        string     `json:"name,omitempty"`
	EditCount   int64      `json:"edit_count,omitempty"`
	DateStarted *time.Time `json:"date_started,omitempty"`
	Groups      []string   `json:"groups,omitempty"`
}

type Probability struct {
	False float64 `json:"false"`
	True  float64 `json:"true"`
}

type Score struct {
	Prediction  bool        `json:"prediction"`
	Probability Probability `json:"probability"`
}

type Scores struct {
	Damaging  Score `json:"damaging"`
	Goodfaith Score `json:"goodfaith"`
}

type Version struct {
	Identifier      int64    `json:"identifier"`
	Editor          Editor   `json:"editor"`
	Comment         string   `json:"comment,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	IsMinorEdit     bool     `json:"is_minor_edit,omitempty"`
	IsFlaggedStable bool     `json:"is_flagged_stable,omitempty"`
	Scores          *Scores  `json:"scores,omitempty"`
}

// TODO: Should Type and Level be enumerations?
// TODO: Should Expiry be time.Time?
type Protection struct {
	Type   string `json:"type"`
	Level  string `json:"level"`
	Expiry string `json:"expiry,omitempty"`
}

type Namespace struct {
	Identifier int64  `json:"identifier"`
	Name       string `json:"name"`
}

type InLanguage struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type EntityRef struct {
	Identifier string   `json:"identifier"`
	URL        string   `json:"url"`
	Aspects    []string `json:"aspects,omitempty"`
}

type Category struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Template struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Redirect struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type IsPartOf struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
}

type ArticleBody struct {
	HTML     string `json:"html"`
	WikiText string `json:"wikitext"`
}

type License struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	URL        string `json:"url,omitempty"`
}

// Article is a Wikimedia Enterprise HTML dump article.
type Article struct {
	Name               string       `json:"name"`
	Identifier         int64        `json:"identifier"`
	DateModified       time.Time    `json:"date_modified"`
	Protection         []Protection `json:"protection,omitempty"`
	Version            Version      `json:"version"`
	URL                string       `json:"url"`
	Namespace          Namespace    `json:"namespace"`
	InLanguage         InLanguage   `json:"in_language"`
	MainEntity         EntityRef    `json:"main_entity"`
	AdditionalEntities []EntityRef  `json:"additional_entities,omitempty"`
	Categories         []Category   `json:"categories,omitempty"`
	Templates          []Template   `json:"templates,omitempty"`
	Redirects          []Redirect   `json:"redirects,omitempty"`
	IsPartOf           IsPartOf     `json:"is_part_of"`
	ArticleBody        ArticleBody  `json:"article_body"`
	License            []License    `json:"license,omitempty"`
}
