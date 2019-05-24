package pulse

import (
	"strings"

	"github.com/ghetzel/go-stockutil/maputil"
	"github.com/ghetzel/go-stockutil/sliceutil"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/go-stockutil/typeutil"
)

const FilterSeparator = `;`
const FieldValueSeparator = `/`

func F(filters interface{}) (f Filter) {
	for _, flt := range sliceutil.Stringify(filters) {
		f = append(f, flt)
	}

	return
}

type Filter []string

func (self Filter) String() string {
	return strings.Join(self, FilterSeparator)
}

func (self Filter) IsMatch(in interface{}) bool {
	if len(self) == 0 {
		return true
	}

	data := maputil.M(in).MapNative()
	data, _ = maputil.CoalesceMap(data, `.`)

	for k, v := range data {
		if self.IsFieldMatch(k, v) {
			return true
		}
	}

	return false
}

func (self Filter) IsFieldMatch(k string, v interface{}) bool {
	if len(self) == 0 {
		return true
	}

	var included bool
	value := typeutil.V(v)

	for _, flt := range self {
		if field, vpair := stringutil.SplitPair(flt, FieldValueSeparator); field == k {
			op, cmp := stringutil.SplitPairTrailing(vpair, `:`)
			cmpI := typeutil.Int(cmp)

			if field == `` || cmp == `` {
				continue
			}

			switch op {
			case `contains`:
				included = strings.Contains(value.String(), cmp)
			case `gt`:
				included = (value.Int() > cmpI)
			case `lt`:
				included = (value.Int() < cmpI)
			case `gte`:
				included = (value.Int() >= cmpI)
			case `lte`:
				included = (value.Int() <= cmpI)
			case `not`:
				included = (value.String() != cmp)
			default:
				included = (value.String() == cmp)
			}

			if included {
				return true
			}
		}
	}

	return false
}

func (self Filter) Apply(in map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{})
	data, _ := maputil.CoalesceMap(in, `.`)

	for k, v := range data {
		if self.IsFieldMatch(k, v) {
			maputil.DeepSet(out, strings.Split(k, `.`), v)
		}
	}

	return out
}
