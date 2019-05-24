package pulse

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFilterApply(t *testing.T) {
	assert := require.New(t)
	original := map[string]interface{}{
		`x`: map[string]interface{}{
			`y`: map[string]interface{}{
				`z`: 123,
			},
		},
		`a`: map[string]interface{}{
			`b`: map[string]interface{}{
				`c`: true,
			},
		},
	}

	flt := Filter{`x.y.z/123`}

	assert.EqualValues(map[string]interface{}{
		`x`: map[string]interface{}{
			`y`: map[string]interface{}{
				`z`: 123,
			},
		},
	}, flt.Apply(original))

	flt = append(flt, `a.b.c/false`)

	assert.EqualValues(map[string]interface{}{
		`x`: map[string]interface{}{
			`y`: map[string]interface{}{
				`z`: 123,
			},
		},
	}, flt.Apply(original))

	flt = append(flt[0:1], `a.b.c/true`)
	assert.EqualValues(original, flt.Apply(original))

	flt = Filter{`x.y.z/gte:123`}
	assert.EqualValues(map[string]interface{}{
		`x`: map[string]interface{}{
			`y`: map[string]interface{}{
				`z`: 123,
			},
		},
	}, flt.Apply(original))

	flt = Filter{`x.y.z/gt:123`}
	assert.Empty(flt.Apply(original))
}
