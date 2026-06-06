package lib

import (
	"errors"
	"fmt"
	"strings"

	"github.com/mattn/go-sqlite3"
)

func ParseSQLiteError(err error) []FormError {
	var sqliteErr sqlite3.Error
	if !errors.As(err, &sqliteErr) {
		return nil
	}

	switch sqliteErr.ExtendedCode {
	case sqlite3.ErrConstraintUnique:
		return parseSQLiteUniqueError(err.Error())
	case sqlite3.ErrConstraintForeignKey:
		return []FormError{
			{
				Field:   "",
				Rule:    "foreign_key",
				Message: "Referenced record does not exist",
			},
		}
	case sqlite3.ErrConstraintNotNull:
		return []FormError{
			{
				Field:   "",
				Rule:    "required",
				Message: "A required field is missing",
			},
		}
	case sqlite3.ErrConstraintCheck:
		return []FormError{
			{
				Field:   "",
				Rule:    "invalid",
				Message: "Invalid value",
			},
		}
	case sqlite3.ErrConstraintPrimaryKey, sqlite3.ErrConstraintRowID:
		return []FormError{
			{
				Field:   "",
				Rule:    "conflict",
				Message: "Record already exists",
			},
		}
	default:
		return []FormError{
			{
				Field:   "",
				Rule:    "database_error",
				Message: "Internal server error",
			},
		}
	}
}

func parseSQLiteUniqueError(msg string) []FormError {
	const prefix = "UNIQUE constraint failed: "

	if !strings.HasPrefix(msg, prefix) {
		return []FormError{
			{
				Field:   "",
				Rule:    "unique",
				Message: "Record already exists",
			},
		}
	}

	targets := strings.Split(strings.TrimPrefix(msg, prefix), ", ")
	errors := make([]FormError, 0, len(targets))

	for _, target := range targets {
		parts := strings.Split(target, ".")
		field := parts[len(parts)-1]

		errors = append(errors, FormError{
			Field:   field,
			Rule:    "unique",
			Message: fmt.Sprintf("%s already exists", field),
		})
	}

	return errors
}
