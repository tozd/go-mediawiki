package mediawiki

import (
	"bytes"
	"encoding/json"
	"math/big"
	"regexp"
	"strconv"
	"time"

	"gitlab.com/tozd/go/errors"
	"gitlab.com/tozd/go/x"
)

type EntityType int

const (
	Item EntityType = iota
	Property
)

var TimeRegex *regexp.Regexp

func init() {
	TimeRegex = regexp.MustCompile(`^([+-]\d{4,})-(\d{2})-(\d{2})T(\d{2}):(\d{2}):(\d{2})Z$`)
}

func (t EntityType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch t {
	case Item:
		buffer.WriteString("item")
	case Property:
		buffer.WriteString("property")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *EntityType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "item":
		*t = Item
	case "property":
		*t = Property
	default:
		return errors.Errorf("unknown entity type: %s", s)
	}
	return nil
}

type WikiBaseEntityType int

const (
	ItemType WikiBaseEntityType = iota
	PropertyType
	LexemeType
	FormType
	SenseType
)

func (t WikiBaseEntityType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch t {
	case ItemType:
		buffer.WriteString("item")
	case PropertyType:
		buffer.WriteString("property")
	case LexemeType:
		buffer.WriteString("lexeme")
	case FormType:
		buffer.WriteString("form")
	case SenseType:
		buffer.WriteString("sense")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *WikiBaseEntityType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "item":
		*t = ItemType
	case "property":
		*t = PropertyType
	case "lexeme":
		*t = LexemeType
	case "form":
		*t = FormType
	case "sense":
		*t = SenseType
	default:
		return errors.Errorf("unknown wikibase entity type: %s", s)
	}
	return nil
}

type StatementType int

const (
	StatementT StatementType = iota
)

func (t StatementType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch t { //nolint:gocritic
	case StatementT:
		buffer.WriteString("statement")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *StatementType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "statement":
		*t = StatementT
	default:
		return errors.Errorf("unknown statement type: %s", s)
	}
	return nil
}

type StatementRank int

const (
	Preferred StatementRank = iota
	Normal
	Deprecated
)

func (r StatementRank) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch r {
	case Preferred:
		buffer.WriteString("preferred")
	case Normal:
		buffer.WriteString("normal")
	case Deprecated:
		buffer.WriteString("deprecated")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (r *StatementRank) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "preferred":
		*r = Preferred
	case "normal":
		*r = Normal
	case "deprecated":
		*r = Deprecated
	default:
		return errors.Errorf("unknown statement rank: %s", s)
	}
	return nil
}

type SnakType int

const (
	Value SnakType = iota
	SomeValue
	NoValue
)

func (t SnakType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch t {
	case Value:
		buffer.WriteString("value")
	case SomeValue:
		buffer.WriteString("somevalue")
	case NoValue:
		buffer.WriteString("novalue")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *SnakType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "value":
		*t = Value
	case "somevalue":
		*t = SomeValue
	case "novalue":
		*t = NoValue
	default:
		return errors.Errorf("unknown snak type: %s", s)
	}
	return nil
}

type DataType int

const (
	WikiBaseItem DataType = iota
	ExternalID
	String
	Quantity
	Time
	GlobeCoordinate
	CommonsMedia
	MonolingualText
	URL
	GeoShape
	WikiBaseLexeme
	WikiBaseSense
	WikiBaseProperty
	Math
	MusicalNotation
	WikiBaseForm
	TabularData
)

func (t DataType) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch t {
	case WikiBaseItem:
		buffer.WriteString("wikibase-item")
	case ExternalID:
		buffer.WriteString("external-id")
	case String:
		buffer.WriteString("string")
	case Quantity:
		buffer.WriteString("quantity")
	case Time:
		buffer.WriteString("time")
	case GlobeCoordinate:
		buffer.WriteString("globe-coordinate")
	case CommonsMedia:
		buffer.WriteString("commonsMedia")
	case MonolingualText:
		buffer.WriteString("monolingualtext")
	case URL:
		buffer.WriteString("url")
	case GeoShape:
		buffer.WriteString("geo-shape")
	case WikiBaseLexeme:
		buffer.WriteString("wikibase-lexeme")
	case WikiBaseSense:
		buffer.WriteString("wikibase-sense")
	case WikiBaseProperty:
		buffer.WriteString("wikibase-property")
	case Math:
		buffer.WriteString("math")
	case MusicalNotation:
		buffer.WriteString("musical-notation")
	case WikiBaseForm:
		buffer.WriteString("wikibase-form")
	case TabularData:
		buffer.WriteString("tabular-data")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *DataType) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "wikibase-item":
		*t = WikiBaseItem
	case "external-id":
		*t = ExternalID
	case "string":
		*t = String
	case "quantity":
		*t = Quantity
	case "time":
		*t = Time
	case "globe-coordinate":
		*t = GlobeCoordinate
	case "commonsMedia":
		*t = CommonsMedia
	case "monolingualtext":
		*t = MonolingualText
	case "url":
		*t = URL
	case "geo-shape":
		*t = GeoShape
	case "wikibase-lexeme":
		*t = WikiBaseLexeme
	case "wikibase-sense":
		*t = WikiBaseSense
	case "wikibase-property":
		*t = WikiBaseProperty
	case "math":
		*t = Math
	case "musical-notation":
		*t = MusicalNotation
	case "wikibase-form":
		*t = WikiBaseForm
	case "tabular-data":
		*t = TabularData
	default:
		return errors.Errorf("unknown data type: %s", s)
	}
	return nil
}

type TimePrecision int

const (
	BillionYears TimePrecision = iota
	HoundredMillionYears
	TenMillionYears
	MillionYears
	HoundredMillenniums
	TenMillenniums
	Millennium
	Century
	Decade
	Year
	Month
	Day
	Hour
	Minute
	Second
)

type CalendarModel int

const (
	Gregorian CalendarModel = iota
	Julian
)

func (t CalendarModel) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	switch t {
	case Gregorian:
		buffer.WriteString("https://www.wikidata.org/wiki/Q1985727")
	case Julian:
		buffer.WriteString("https://www.wikidata.org/wiki/Q1985786")
	}
	buffer.WriteString(`"`)
	return buffer.Bytes(), nil
}

func (t *CalendarModel) UnmarshalJSON(b []byte) error {
	var s string
	err := json.Unmarshal(b, &s)
	if err != nil {
		return errors.WithStack(err)
	}
	switch s {
	case "https://www.wikidata.org/wiki/Q1985727":
		*t = Gregorian
	case "http://www.wikidata.org/entity/Q1985727":
		*t = Gregorian
	case "https://www.wikidata.org/wiki/Q1985786":
		*t = Julian
	case "http://www.wikidata.org/entity/Q1985786":
		*t = Julian
	default:
		return errors.Errorf("unknown calendar model: %s", s)
	}
	return nil
}

type ErrorValue string

type StringValue string

type WikiBaseEntityIDValue struct {
	Type WikiBaseEntityType
	ID   string
}

type GlobeCoordinateValue struct {
	Latitude  float64
	Longitude float64
	Precision float64
	Globe     string
}

type MonolingualTextValue struct {
	Language string
	Text     string
}

type QuantityValue struct {
	Amount     big.Float
	UpperBound *big.Float `json:"upperBound"`
	LowerBound *big.Float `json:"lowerBound"`
	Unit       string
}

type TimeValue struct {
	Time      time.Time
	Precision TimePrecision
	Calendar  CalendarModel
}

type DataValue struct {
	Value interface{}
}

func (v DataValue) MarshalJSON() ([]byte, error) {
	return nil, nil
}

func parseTime(t string) (time.Time, errors.E) {
	match := TimeRegex.FindStringSubmatch(t)
	if match == nil {
		return time.Time{}, errors.Errorf(`unable to parse time "%s"`, t)
	}
	year, err := strconv.ParseInt(match[1], 10, 0) //nolint:gomnd
	if err != nil {
		return time.Time{}, errors.WithMessagef(err, `unable to parse year "%s"`, t)
	}
	if year < 0 {
		// Wikidata uses historical numbering, in which year 0 is undefined,
		// but Go uses astronomical numbering, so we add 1 here.
		year++
	}
	month, err := strconv.ParseInt(match[2], 10, 0) //nolint:gomnd
	if err != nil {
		return time.Time{}, errors.WithMessagef(err, `unable to parse month "%s"`, t)
	}
	if month == 0 {
		// Wikidata uses 0 when month is unknown or insignificant.
		// Go does not support this, so we set it to 1 here.
		month = 1
	}
	day, err := strconv.ParseInt(match[3], 10, 0) //nolint:gomnd
	if err != nil {
		return time.Time{}, errors.WithMessagef(err, `unable to parse day "%s"`, t)
	}
	if day == 0 {
		// Wikidata uses 0 when day is unknown or insignificant.
		// Go does not support this, so we set it to 1 here.
		day = 1
	}
	hour, err := strconv.ParseInt(match[4], 10, 0) //nolint:gomnd
	if err != nil {
		return time.Time{}, errors.WithMessagef(err, `unable to parse hour "%s"`, t)
	}
	minute, err := strconv.ParseInt(match[5], 10, 0) //nolint:gomnd
	if err != nil {
		return time.Time{}, errors.WithMessagef(err, `unable to parse minute "%s"`, t)
	}
	second, err := strconv.ParseInt(match[6], 10, 0) //nolint:gomnd
	if err != nil {
		return time.Time{}, errors.WithMessagef(err, `unable to parse second "%s"`, t)
	}
	return time.Date(int(year), time.Month(month), int(day), int(hour), int(minute), int(second), 0, time.UTC), nil
}

func (v *DataValue) UnmarshalJSON(b []byte) error {
	var t struct {
		Type  string
		Error string
	}
	err := json.Unmarshal(b, &t)
	if err != nil {
		return errors.WithStack(err)
	}
	if t.Error != "" {
		v.Value = ErrorValue(t.Error)
		return nil
	}
	switch t.Type {
	case "string":
		var t struct {
			Type  string
			Value string
		}
		err := x.UnmarshalWithoutUnknownFields(b, &t)
		if err != nil {
			return err
		}
		v.Value = StringValue(t.Value)
	case "wikibase-entityid":
		var t struct {
			Type  string
			Value struct {
				Type WikiBaseEntityType `json:"entity-type"`
				ID   string
				// Not available for all entity types. Not recommended to be used. We ignore it.
				NumericID int `json:"numeric-id"`
			}
		}
		err := x.UnmarshalWithoutUnknownFields(b, &t)
		if err != nil {
			return err
		}
		v.Value = WikiBaseEntityIDValue{
			Type: t.Value.Type,
			ID:   t.Value.ID,
		}
	case "globecoordinate":
		var t struct {
			Type  string
			Value struct {
				Latitude  float64
				Longitude float64
				// Altitude is deprecated and no longer used. We ignore it.
				Altitude  float64
				Precision float64
				Globe     string
			}
		}
		err := x.UnmarshalWithoutUnknownFields(b, &t)
		if err != nil {
			return err
		}
		v.Value = GlobeCoordinateValue{
			Latitude:  t.Value.Latitude,
			Longitude: t.Value.Longitude,
			Precision: t.Value.Precision,
			Globe:     t.Value.Globe,
		}
	case "monolingualtext":
		var t struct {
			Type  string
			Value MonolingualTextValue
		}
		err := x.UnmarshalWithoutUnknownFields(b, &t)
		if err != nil {
			return err
		}
		v.Value = t.Value
	case "quantity":
		var t struct {
			Type  string
			Value QuantityValue
		}
		err := x.UnmarshalWithoutUnknownFields(b, &t)
		if err != nil {
			return err
		}
		v.Value = t.Value
	case "time":
		var t struct {
			Type  string
			Value struct {
				Time      string
				Precision TimePrecision
				Calendar  CalendarModel `json:"calendarmodel"`
				// Defined and declared not used, but sometimes still set. We ignore it.
				Timezone int64
				// Defined and declared not used, but sometimes still set. We ignore it.
				Before int64
				// Defined and declared not used, but sometimes still set. We ignore it.
				After int64
			}
		}
		err := x.UnmarshalWithoutUnknownFields(b, &t)
		if err != nil {
			return err
		}
		parsedTime, err := parseTime(t.Value.Time)
		if err != nil {
			return err
		}
		v.Value = TimeValue{
			Time:      parsedTime,
			Precision: t.Value.Precision,
			Calendar:  t.Value.Calendar,
		}
	default:
		return errors.Errorf(`unknown data value type "%s": %s`, t.Type, string(b))
	}
	return nil
}

type LanguageValue struct {
	Language string
	Value    string
}

type SiteLink struct {
	Site   string
	Title  string
	Badges []string
	URL    string
}

type Snak struct {
	Hash      string
	SnakType  SnakType `json:"snaktype"`
	Property  string
	DataType  DataType  `json:"datatype"`
	DataValue DataValue `json:"datavalue"`
}

type Reference struct {
	Hash       string
	Snaks      map[string][]Snak
	SnaksOrder []string `json:"snaks-order"`
}

type Statement struct {
	ID              string
	Type            StatementType
	MainSnak        Snak `json:"mainsnak"`
	Rank            StatementRank
	Qualifiers      map[string][]Snak
	QualifiersOrder []string `json:"qualifiers-order"`
	References      []Reference
}

type Entity struct {
	ID           string
	Type         EntityType
	DataType     string `json:"datatype"`
	Labels       map[string]LanguageValue
	Descriptions map[string]LanguageValue
	Aliases      map[string][]LanguageValue
	Claims       map[string][]Statement
	SiteLinks    map[string]SiteLink
	LastRevID    int64 `json:"lastrevid"`
}
