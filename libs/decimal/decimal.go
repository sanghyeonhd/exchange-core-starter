package decimal

import (
	"errors"
	"strconv"
	"strings"
)

var (
	ErrScaleMismatch = errors.New("decimal scale mismatch")
	ErrInvalid       = errors.New("invalid decimal")
	ErrOverflow      = errors.New("decimal overflow")
)

type Decimal struct {
	Value int64
	Scale int32
}

func New(value int64, scale int32) Decimal {
	return Decimal{Value: value, Scale: scale}
}

func Parse(s string, scale int32) (Decimal, error) {
	if scale < 0 {
		return Decimal{}, ErrInvalid
	}

	s = strings.TrimSpace(s)
	if s == "" {
		return Decimal{}, ErrInvalid
	}

	negative := false
	if s[0] == '-' {
		negative = true
		s = s[1:]
	}
	if s == "" {
		return Decimal{}, ErrInvalid
	}

	parts := strings.Split(s, ".")
	if len(parts) > 2 {
		return Decimal{}, ErrInvalid
	}

	whole := parts[0]
	fraction := ""
	if len(parts) == 2 {
		fraction = parts[1]
	}
	if whole == "" {
		whole = "0"
	}
	if len(fraction) > int(scale) {
		return Decimal{}, ErrInvalid
	}
	for len(fraction) < int(scale) {
		fraction += "0"
	}

	raw := whole + fraction
	if raw == "" {
		return Decimal{}, ErrInvalid
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return Decimal{}, ErrInvalid
	}
	if negative {
		value = -value
	}
	return Decimal{Value: value, Scale: scale}, nil
}

func (d Decimal) Add(other Decimal) (Decimal, error) {
	if d.Scale != other.Scale {
		return Decimal{}, ErrScaleMismatch
	}
	return Decimal{Value: d.Value + other.Value, Scale: d.Scale}, nil
}

func (d Decimal) Sub(other Decimal) (Decimal, error) {
	if d.Scale != other.Scale {
		return Decimal{}, ErrScaleMismatch
	}
	return Decimal{Value: d.Value - other.Value, Scale: d.Scale}, nil
}

func (d Decimal) String() string {
	value := d.Value
	sign := ""
	if value < 0 {
		sign = "-"
		value = -value
	}

	if d.Scale == 0 {
		return sign + strconv.FormatInt(value, 10)
	}

	pow := int64(1)
	for i := int32(0); i < d.Scale; i++ {
		pow *= 10
	}
	whole := value / pow
	fraction := value % pow
	fractionText := strconv.FormatInt(fraction, 10)
	for len(fractionText) < int(d.Scale) {
		fractionText = "0" + fractionText
	}
	return sign + strconv.FormatInt(whole, 10) + "." + fractionText
}
