package mediawiki

import (
	"time"
)

type Editor struct {
	Identifier        int64      `json:"identifier,omitempty"`
	IsAnonymous       bool       `json:"is_anonymous,omitempty"`
	IsBot             bool       `json:"is_bot,omitempty"`
	IsAdmin           bool       `json:"is_admin,omitempty"`
	IsPatroller       bool       `json:"is_patroller,omitempty"`
	HasAdvancedRights bool       `json:"has_advanced_rights,omitempty"`
	Name              string     `json:"name,omitempty"`
	EditCount         int64      `json:"edit_count,omitempty"`
	DateStarted       *time.Time `json:"date_started,omitempty"`
	Groups            []string   `json:"groups,omitempty"`
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
	Damaging  *Score `json:"damaging,omitempty"`
	Goodfaith *Score `json:"goodfaith,omitempty"`
}

type Size struct {
	Value int64  `json:"value"`
	Unit  string `json:"unit_text"`
}

type Version struct {
	Identifier          int64    `json:"identifier"`
	Editor              *Editor  `json:"editor,omitempty"`
	Comment             string   `json:"comment,omitempty"`
	Tags                []string `json:"tags,omitempty"`
	HasTagNeedsCitation bool     `json:"has_tag_needs_citation,omitempty"`
	IsMinorEdit         bool     `json:"is_minor_edit,omitempty"`
	IsFlaggedStable     bool     `json:"is_flagged_stable,omitempty"`
	Scores              *Scores  `json:"scores,omitempty"`
	Size                *Size    `json:"size,omitempty"`
	NumberOfCharacters  int64    `json:"number_of_characters,omitempty"`
	Event               Event    `json:"event"`
}

// TODO: Should Type and Level be enumerations?
// TODO: Should Expiry be time.Time?

type Protection struct {
	Type   string `json:"type"`
	Level  string `json:"level"`
	Expiry string `json:"expiry,omitempty"`
}

type Namespace struct {
	Identifier int64 `json:"identifier"`
}

type InLanguage struct {
	Identifier string `json:"identifier"`
}

// TODO: Should we parse Aspects?

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
	URL        string `json:"url,omitempty"`
}

type ArticleBody struct {
	HTML     string `json:"html"`
	WikiText string `json:"wikitext"`
}

type License struct {
	Identifier string `json:"identifier"`
	Name       string `json:"name"`
	URL        string `json:"url"`
}

type Visibility struct {
	Text    bool `json:"text"`
	Editor  bool `json:"editor"`
	Comment bool `json:"comment"`
}

type Image struct {
	ContentURL string `json:"content_url"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
}

// TODO: Should Type be enumeration?

type Event struct {
	Identifier    string     `json:"identifier"`
	Type          string     `json:"type"`
	DateCreated   time.Time  `json:"date_created"`
	DatePublished *time.Time `json:"date_published,omitempty"`
	Partition     int        `json:"partition,omitempty"`
	Offset        int64      `json:"offset,omitempty"`
}

type Link struct {
	URL    string  `json:"url"`
	Text   string  `json:"text,omitempty"`
	Images []Image `json:"images,omitempty"`
}

// TODO: Should Type be enumeration?

type InfoBox struct {
	Name     string    `json:"name,omitempty"`
	Type     string    `json:"type"`
	Value    string    `json:"value,omitempty"`
	Values   []string  `json:"values,omitempty"`
	HasParts []InfoBox `json:"has_parts,omitempty"`
	Images   []Image   `json:"images,omitempty"`
	Links    []Link    `json:"links,omitempty"`
}

// Article is a Wikimedia Enterprise HTML dump article.
type Article struct {
	Name                   string       `json:"name"`
	Identifier             int64        `json:"identifier"`
	Abstract               string       `json:"abstract,omitempty"`
	WatchersCount          int64        `json:"watchers_count,omitempty"`
	DateCreated            time.Time    `json:"date_created"`
	DateModified           time.Time    `json:"date_modified"`
	DatePreviouslyModified *time.Time   `json:"date_previously_modified,omitempty"`
	Protection             []Protection `json:"protection,omitempty"`
	Version                Version      `json:"version"`
	PreviousVersion        *Version     `json:"previous_version,omitempty"`
	URL                    string       `json:"url"`
	Namespace              Namespace    `json:"namespace"`
	InLanguage             InLanguage   `json:"in_language"`
	MainEntity             *EntityRef   `json:"main_entity,omitempty"`
	AdditionalEntities     []EntityRef  `json:"additional_entities,omitempty"`
	Categories             []Category   `json:"categories,omitempty"`
	Templates              []Template   `json:"templates,omitempty"`
	Redirects              []Redirect   `json:"redirects,omitempty"`
	IsPartOf               IsPartOf     `json:"is_part_of"`
	ArticleBody            ArticleBody  `json:"article_body"`
	License                []License    `json:"license,omitempty"`
	Visibility             *Visibility  `json:"visibility,omitempty"`
	Image                  *Image       `json:"image,omitempty"`
	Event                  Event        `json:"event"`
	InfoBox                []InfoBox    `json:"infobox,omitempty"`
}
