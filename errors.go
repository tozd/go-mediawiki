package mediawiki

import (
	"gitlab.com/tozd/go/errors"
)

var (
	ErrUnexpectedType = errors.Base("unexpected type")
	ErrInvalidValue   = errors.Base("invalid value")
	ErrNotFound       = errors.Base("not found")
	ErrJSONDecode     = errors.Base("cannot decode json")
	ErrSQLParse       = errors.Base("cannot parse SQL")
)
