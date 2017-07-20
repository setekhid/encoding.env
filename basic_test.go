package env_test

import (
	"strings"
	"testing"

	"github.com/setekhid/encoding.env"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBasic(t *testing.T) {
	t.Parallel()

	type cType struct {
		CA float32 `env:"ca,omitempty"`
		CB uint16  `env:"cb"`
	}

	type abcType struct {
		A string              `env:"a,omitempty"`
		B int                 `env:"b"`
		C cType               `env:"c"`
		D [][]int             `env:"blabla"`
		E map[string][]string `env:"e"`
		F string
		G map[string]int
	}

	obj := abcType{
		B: 123,
		C: cType{
			CA: 4.3,
		},
		D: [][]int{{1, 2, 3}},
		E: map[string][]string{"abc": {"bb", "cc"}},
		F: "ajfieoaf",
	}

	envs, err := env.Marshal(&obj)
	require.NoError(t, err)

	lines := strings.Split(string(envs), "\n")
	for _, line := range lines {
		t.Log(line)
	}

	obj_ := abcType{
		D: [][]int{make([]int, 3)},
		E: map[string][]string{"abc": make([]string, 2)},
	}

	err = env.Unmarshal(envs, &obj_)
	require.NoError(t, err)

	assert.Equal(t, obj, obj_)
}
