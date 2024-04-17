package repo

import (
	"reflect"
	"time"

	"github.com/mitchellh/mapstructure"

	"github.com/axiomesh/axiom-kit/types"
)

var coinNumberType = reflect.TypeOf(types.CoinNumber{})

var durationType = reflect.TypeOf(Duration(5))

type Duration time.Duration

func (d *Duration) MarshalText() (text []byte, err error) {
	return []byte(time.Duration(*d).String()), nil
}

func (d *Duration) UnmarshalText(b []byte) error {
	x, err := time.ParseDuration(string(b))
	if err != nil {
		return err
	}
	*d = Duration(x)
	return nil
}

func StringToTimeDurationHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != durationType {
			return data, nil
		}

		d, err := time.ParseDuration(data.(string))
		if err != nil {
			return nil, err
		}
		return Duration(d), nil
	}
}

func (d *Duration) ToDuration() time.Duration {
	return time.Duration(*d)
}

func (d *Duration) String() string {
	return time.Duration(*d).String()
}

func StringToCoinNumberHookFunc() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data any) (any, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}
		if t != coinNumberType {
			return data, nil
		}

		return types.ParseCoinNumber(data.(string))
	}
}
