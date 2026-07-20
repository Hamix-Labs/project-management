package domain

import (
	"fmt"
	"unicode/utf8"
)

// Field size limits for project context items. Kept below the default HTTP
// request body cap (1 MiB) so oversize titles/bodies fail with a clear 400
// instead of a generic 413.
const (
	MaxProjectContextBodyBytes        = 512 << 10 // 512 KiB
	MaxProjectContextTitleChars       = 200
	MaxProjectContextDescriptionChars = 400
)

// ValidateProjectContextTitle reports ErrInvalidInput when title is empty or
// longer than MaxProjectContextTitleChars (Unicode code points).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidateProjectContextTitle(title string) error {
	if title == "" {
		return fmt.Errorf("%w: context title required", ErrInvalidInput)
	}
	if utf8.RuneCountInString(title) > MaxProjectContextTitleChars {
		return fmt.Errorf("%w: context title must be %d characters or fewer", ErrInvalidInput, MaxProjectContextTitleChars)
	}
	return nil
}

// ValidateProjectContextBody reports ErrInvalidInput when body is empty or
// larger than MaxProjectContextBodyBytes (UTF-8 byte length).
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidateProjectContextBody(body string) error {
	if body == "" {
		return fmt.Errorf("%w: context body required", ErrInvalidInput)
	}
	if len(body) > MaxProjectContextBodyBytes {
		return fmt.Errorf("%w: context body must be %d bytes or smaller", ErrInvalidInput, MaxProjectContextBodyBytes)
	}
	return nil
}

// ValidateProjectContextDescription reports ErrInvalidInput when description is
// longer than MaxProjectContextDescriptionChars (Unicode code points). Empty is
// allowed so existing rows and non-import creates stay valid.
//
//funclogmeasure:skip category=hot-path reason="Pure helper without I/O; operation trace is emitted by the calling chokepoint."
func ValidateProjectContextDescription(description string) error {
	if utf8.RuneCountInString(description) > MaxProjectContextDescriptionChars {
		return fmt.Errorf("%w: context description must be %d characters or fewer", ErrInvalidInput, MaxProjectContextDescriptionChars)
	}
	return nil
}
