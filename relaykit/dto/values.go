package dto

import (
	"encoding/json"
	"math"
	"strconv"
)

type StringValue string

func (s *StringValue) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err == nil {
		*s = StringValue(str)
		return nil
	}

	var raw json.Number
	if err := json.Unmarshal(data, &raw); err == nil {
		*s = StringValue(raw.String())
		return nil
	}

	return json.Unmarshal(data, &str)
}

func (s StringValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s))
}

type IntValue int

func parseIntValue(value string) (IntValue, error) {
	if integer, err := strconv.Atoi(value); err == nil {
		return IntValue(integer), nil
	}

	number, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, err
	}
	intLimit := math.Ldexp(1, strconv.IntSize-1)
	if math.IsNaN(number) || math.IsInf(number, 0) || number < -intLimit || number >= intLimit {
		return 0, &strconv.NumError{Func: "ParseIntValue", Num: value, Err: strconv.ErrRange}
	}
	return IntValue(number), nil
}

func (i *IntValue) UnmarshalJSON(data []byte) error {
	var n int
	if err := json.Unmarshal(data, &n); err == nil {
		*i = IntValue(n)
		return nil
	}

	var number json.Number
	if err := json.Unmarshal(data, &number); err == nil {
		value, err := parseIntValue(number.String())
		if err != nil {
			return err
		}
		*i = value
		return nil
	}

	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}
	value, err := parseIntValue(s)
	if err != nil {
		return err
	}
	*i = value
	return nil
}

func (i IntValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(int(i))
}

type BoolValue bool

func (b *BoolValue) UnmarshalJSON(data []byte) error {
	var boolean bool
	if err := json.Unmarshal(data, &boolean); err == nil {
		*b = BoolValue(boolean)
		return nil
	}
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	if str == "true" {
		*b = BoolValue(true)
	} else if str == "false" {
		*b = BoolValue(false)
	} else {
		return json.Unmarshal(data, &boolean)
	}
	return nil
}
func (b BoolValue) MarshalJSON() ([]byte, error) {
	return json.Marshal(bool(b))
}
