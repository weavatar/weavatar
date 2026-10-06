package rule

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/go-rio/rio"
	"github.com/libtnb/validator"
)

var (
	_ validator.FallibleRule = (*Exists)(nil)
	_ validator.FallibleRule = (*NotExists)(nil)
)

// identifier restricts table and column arguments, which are spliced into SQL.
var identifier = regexp.MustCompile(`^[a-z_][a-z0-9_]*$`)

// Exists passes when the value matches any of the listed columns:
// exists:users,phone,email. Empty values pass; presence is required's job.
type Exists struct {
	db *rio.DB
}

func NewExists(db *rio.DB) *Exists {
	return &Exists{db: db}
}

func (r *Exists) Signature() string { return "exists" }

func (r *Exists) Message() string { return "{field} 不存在" }

func (r *Exists) CheckArgs(args []string) error { return checkTableArgs(args) }

func (r *Exists) Validate(f *validator.Field) (bool, error) {
	n, empty, err := countMatches(f, r.db)
	if empty || err != nil {
		return empty, err
	}
	return n > 0, nil
}

// NotExists passes when the value matches none of the listed columns:
// not_exists:users,phone. Empty values pass; presence is required's job.
type NotExists struct {
	db *rio.DB
}

func NewNotExists(db *rio.DB) *NotExists {
	return &NotExists{db: db}
}

func (r *NotExists) Signature() string { return "not_exists" }

func (r *NotExists) Message() string { return "{field} 已存在" }

func (r *NotExists) CheckArgs(args []string) error { return checkTableArgs(args) }

func (r *NotExists) Validate(f *validator.Field) (bool, error) {
	n, empty, err := countMatches(f, r.db)
	if empty || err != nil {
		return empty, err
	}
	return n == 0, nil
}

func checkTableArgs(args []string) error {
	if len(args) < 2 {
		return errors.New("want a table and at least one column")
	}
	for _, arg := range args {
		if !identifier.MatchString(arg) {
			return fmt.Errorf("%q is not a lowercase identifier", arg)
		}
	}
	return nil
}

// countMatches counts rows whose listed columns equal the field value; empty
// reports a blank value that was not looked up.
func countMatches(f *validator.Field, db *rio.DB) (n int64, empty bool, err error) {
	if validator.IsEmptyValue(f.Reflect()) {
		return 0, true, nil
	}
	value, ok := f.Value[any]()
	if !ok {
		return 0, false, nil
	}
	if db == nil {
		return 0, false, errors.New("rule: no database configured")
	}

	args := f.Attrs()
	columns := args[1:]
	conds := make([]string, len(columns))
	values := make([]any, len(columns))
	for i, column := range columns {
		conds[i] = column + " = ?"
		values[i] = value
	}

	n, err = rio.Raw[int64]("SELECT count(*) FROM "+args[0]).
		Where(strings.Join(conds, " OR "), values...).
		Value(f.Context(), db)
	if err != nil {
		return 0, false, fmt.Errorf("rule: count %s: %w", args[0], err)
	}
	return n, false, nil
}
