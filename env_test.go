package env

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type Config struct {
	Some        string `env:"somevar"`
	Other       bool   `env:"othervar"`
	Port        int    `env:"PORT"`
	Int64Val    int64  `env:"INT64VAL"`
	UintVal     uint   `env:"UINTVAL"`
	Uint64Val   uint64 `env:"UINT64VAL"`
	NotAnEnv    string
	DatabaseURL string          `env:"DATABASE_URL" envDefault:"postgres://localhost:5432/db"`
	Strings     []string        `env:"STRINGS"`
	SepStrings  []string        `env:"SEPSTRINGS" envSeparator:":"`
	Numbers     []int           `env:"NUMBERS"`
	Numbers64   []int64         `env:"NUMBERS64"`
	UNumbers64  []uint64        `env:"UNUMBERS64"`
	Bools       []bool          `env:"BOOLS"`
	Duration    time.Duration   `env:"DURATION"`
	Float32     float32         `env:"FLOAT32"`
	Float64     float64         `env:"FLOAT64"`
	Float32s    []float32       `env:"FLOAT32S"`
	Float64s    []float64       `env:"FLOAT64S"`
	Durations   []time.Duration `env:"DURATIONS"`
}

type ParentStruct struct {
	InnerStruct *InnerStruct
	unexported  *InnerStruct
	Ignored     *http.Client
}

type InnerStruct struct {
	Inner  string `env:"innervar"`
	Number uint   `env:"innernum"`
}

type TestAgainst struct {
	setenv       func(key, val string)
	run          func(data interface{}) error
	runWithFuncs func(data interface{}, c CustomParsers) error
}

func TestEnv(t *testing.T) {
	cases := map[string]TestAgainst{
		"default": {
			setenv: func(key, val string) {
				os.Setenv(key, val)
			},
			run:          Parse,
			runWithFuncs: ParseWithFuncs,
		},
		"prefix": {
			setenv: func(key, val string) {
				os.Setenv("PREFIX_"+key, val)
			},
			run: func(data interface{}) error {
				return PrefixedParse(data, "PREFIX_")
			},
			runWithFuncs: func(data interface{}, c CustomParsers) error {
				return PrefixedParseWithFuncs(data, c, "PREFIX_")
			},
		},
	}

	wrap := func(f func(*testing.T, TestAgainst), a TestAgainst) func(*testing.T) {
		return func(t *testing.T) {
			f(t, a)
		}
	}

	for n, c := range cases {
		t.Run(n, func(t *testing.T) {
			t.Run("ParseEnv", wrap(testParsesEnv, c))
			t.Run("ParseEnvInner", wrap(testParsesEnvInner, c))
			t.Run("ParsesEnvInnerNil", wrap(testParsesEnvInnerNil, c))
			t.Run("ParseEnvInnerInvalid", wrap(testParsesEnvInnerInvalid, c))
			t.Run("EmptyVars", wrap(testEmptyVars, c))
			t.Run("PassAnInvalidPtr", wrap(testPassAnInvalidPtr, c))
			t.Run("PassReference", wrap(testPassReference, c))
			t.Run("InvalidBool", wrap(testInvalidBool, c))
			t.Run("InvalidInt", wrap(testInvalidInt, c))
			t.Run("InvalidUint", wrap(testInvalidUint, c))
			t.Run("InvalidFloat32", wrap(testInvalidFloat32, c))
			t.Run("InvalidFloat64", wrap(testInvalidFloat64, c))
			t.Run("InvalidInt64", wrap(testInvalidInt64, c))
			t.Run("InvalidUint64", wrap(testInvalidUint64, c))
			t.Run("InvalidInt64Slice", wrap(testInvalidInt64Slice, c))
			t.Run("InvalidUInt64Slice", wrap(testInvalidUInt64Slice, c))
			t.Run("InvalidFloat32Slice", wrap(testInvalidFloat32Slice, c))
			t.Run("InvalidFloat64Slice", wrap(testInvalidFloat64Slice, c))
			t.Run("InvalidBoolsSlice", wrap(testInvalidBoolsSlice, c))
			t.Run("InvalidDuration", wrap(testInvalidDuration, c))
			t.Run("InvalidDurations", wrap(testInvalidDurations, c))
			t.Run("ParsesDefaultconfig", wrap(testParsesDefaultConfig, c))
			t.Run("ParseStructWithoutEnvTag", wrap(testParseStructWithoutEnvTag, c))
			t.Run("ParseStructWithInvalidFieldKind", wrap(testParseStructWithInvalidFieldKind, c))
			t.Run("UnsupportedSliceType", wrap(testUnsupportedSliceType, c))
			t.Run("BadSeparator", wrap(testBadSeparator, c))
			t.Run("NoErrorRequiredSet", wrap(testNoErrorRequiredSet, c))
			t.Run("ErrorRequiredNotSet", wrap(testErrorRequiredNotSet, c))
			t.Run("CustomParser", wrap(testCustomParser, c))
			t.Run("ParseWithFuncsNoPtr", wrap(testParseWithFuncsNoPtr, c))
			t.Run("ParseWithFuncsInvalidType", wrap(testParseWithFuncsInvalidType, c))
			t.Run("CustomParserError", wrap(testCustomParserError, c))
			t.Run("UnsupportedStructType", wrap(testUnsupportedStructType, c))
			t.Run("EmptyOption", wrap(testEmptyOption, c))
			t.Run("ErrorOptionNotRecognized", wrap(testErrorOptionNotRecognized, c))
		})
	}
}

func testParsesEnv(t *testing.T, a TestAgainst) {
	a.setenv("somevar", "somevalue")
	a.setenv("othervar", "true")
	a.setenv("PORT", "8080")
	a.setenv("STRINGS", "string1,string2,string3")
	a.setenv("SEPSTRINGS", "string1:string2:string3")
	a.setenv("NUMBERS", "1,2,3,4")
	a.setenv("NUMBERS64", "1,2,2147483640,-2147483640")
	a.setenv("UNUMBERS64", "1,2,214748364011,9147483641")
	a.setenv("BOOLS", "t,TRUE,0,1")
	a.setenv("DURATION", "1s")
	a.setenv("FLOAT32", "3.40282346638528859811704183484516925440e+38")
	a.setenv("FLOAT64", "1.797693134862315708145274237317043567981e+308")
	a.setenv("FLOAT32S", "1.0,2.0,3.0")
	a.setenv("FLOAT64S", "1.0,2.0,3.0")
	a.setenv("UINTVAL", "44")
	a.setenv("UINT64VAL", "6464")
	a.setenv("INT64VAL", "-7575")
	a.setenv("DURATIONS", "1s,2s,3s")

	defer os.Clearenv()

	cfg := Config{}
	assert.NoError(t, a.run(&cfg))
	assert.Equal(t, "somevalue", cfg.Some)
	assert.Equal(t, true, cfg.Other)
	assert.Equal(t, 8080, cfg.Port)
	assert.Equal(t, uint(44), cfg.UintVal)
	assert.Equal(t, int64(-7575), cfg.Int64Val)
	assert.Equal(t, uint64(6464), cfg.Uint64Val)
	assert.Equal(t, []string{"string1", "string2", "string3"}, cfg.Strings)
	assert.Equal(t, []string{"string1", "string2", "string3"}, cfg.SepStrings)
	assert.Equal(t, []int{1, 2, 3, 4}, cfg.Numbers)
	assert.Equal(t, []int64{1, 2, 2147483640, -2147483640}, cfg.Numbers64)
	assert.Equal(t, []uint64{1, 2, 214748364011, 9147483641}, cfg.UNumbers64)
	assert.Equal(t, []bool{true, true, false, true}, cfg.Bools)
	d1, _ := time.ParseDuration("1s")
	assert.Equal(t, d1, cfg.Duration)
	f32 := float32(3.40282346638528859811704183484516925440e+38)
	assert.Equal(t, f32, cfg.Float32)
	f64 := float64(1.797693134862315708145274237317043567981e+308)
	assert.Equal(t, f64, cfg.Float64)
	assert.Equal(t, []float32{float32(1.0), float32(2.0), float32(3.0)}, cfg.Float32s)
	assert.Equal(t, []float64{float64(1.0), float64(2.0), float64(3.0)}, cfg.Float64s)
	d2, _ := time.ParseDuration("2s")
	d3, _ := time.ParseDuration("3s")
	assert.Equal(t, []time.Duration{d1, d2, d3}, cfg.Durations)
}

func testParsesEnvInner(t *testing.T, a TestAgainst) {
	a.setenv("innervar", "someinnervalue")
	defer os.Clearenv()
	cfg := ParentStruct{
		InnerStruct: &InnerStruct{},
		unexported:  &InnerStruct{},
	}
	assert.NoError(t, a.run(&cfg))
	assert.Equal(t, "someinnervalue", cfg.InnerStruct.Inner)
}

func testParsesEnvInnerNil(t *testing.T, a TestAgainst) {
	a.setenv("innervar", "someinnervalue")
	defer os.Clearenv()
	cfg := ParentStruct{}
	assert.NoError(t, a.run(&cfg))
}

func testParsesEnvInnerInvalid(t *testing.T, a TestAgainst) {
	a.setenv("innernum", "-547")
	defer os.Clearenv()
	cfg := ParentStruct{
		InnerStruct: &InnerStruct{},
	}
	assert.Error(t, a.run(&cfg))
}

func testEmptyVars(t *testing.T, a TestAgainst) {
	cfg := Config{}
	assert.NoError(t, a.run(&cfg))
	assert.Equal(t, "", cfg.Some)
	assert.Equal(t, false, cfg.Other)
	assert.Equal(t, 0, cfg.Port)
	assert.Equal(t, uint(0), cfg.UintVal)
	assert.Equal(t, uint64(0), cfg.Uint64Val)
	assert.Equal(t, int64(0), cfg.Int64Val)
	assert.Equal(t, 0, len(cfg.Strings))
	assert.Equal(t, 0, len(cfg.SepStrings))
	assert.Equal(t, 0, len(cfg.Numbers))
	assert.Equal(t, 0, len(cfg.Bools))
}

func testPassAnInvalidPtr(t *testing.T, a TestAgainst) {
	var thisShouldBreak int
	assert.Error(t, a.run(&thisShouldBreak))
}

func testPassReference(t *testing.T, a TestAgainst) {
	cfg := Config{}
	assert.Error(t, a.run(cfg))
}

func testInvalidBool(t *testing.T, a TestAgainst) {
	a.setenv("othervar", "should-be-a-bool")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidInt(t *testing.T, a TestAgainst) {
	a.setenv("PORT", "should-be-an-int")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidUint(t *testing.T, a TestAgainst) {
	a.setenv("UINTVAL", "-44")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidFloat32(t *testing.T, a TestAgainst) {
	a.setenv("FLOAT32", "AAA")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidFloat64(t *testing.T, a TestAgainst) {
	a.setenv("FLOAT64", "AAA")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidUint64(t *testing.T, a TestAgainst) {
	a.setenv("UINT64VAL", "AAA")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidInt64(t *testing.T, a TestAgainst) {
	a.setenv("INT64VAL", "AAA")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidInt64Slice(t *testing.T, a TestAgainst) {
	type config struct {
		BadFloats []int64 `env:"BADINTS"`
	}

	a.setenv("BADINTS", "A,2,3")
	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testInvalidUInt64Slice(t *testing.T, a TestAgainst) {
	type config struct {
		BadFloats []uint64 `env:"BADINTS"`
	}

	a.setenv("BADFLOATS", "A,2,3")
	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testInvalidFloat32Slice(t *testing.T, a TestAgainst) {
	type config struct {
		BadFloats []float32 `env:"BADFLOATS"`
	}

	a.setenv("BADFLOATS", "A,2.0,3.0")
	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testInvalidFloat64Slice(t *testing.T, a TestAgainst) {
	type config struct {
		BadFloats []float64 `env:"BADFLOATS"`
	}

	a.setenv("BADFLOATS", "A,2.0,3.0")
	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testInvalidBoolsSlice(t *testing.T, a TestAgainst) {
	type config struct {
		BadBools []bool `env:"BADBOOLS"`
	}

	a.setenv("BADBOOLS", "t,f,TRUE,faaaalse")
	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testInvalidDuration(t *testing.T, a TestAgainst) {
	a.setenv("DURATION", "should-be-a-valid-duration")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testInvalidDurations(t *testing.T, a TestAgainst) {
	a.setenv("DURATIONS", "1s,contains-an-invalid-duration,3s")
	defer os.Clearenv()

	cfg := Config{}
	assert.Error(t, a.run(&cfg))
}

func testParsesDefaultConfig(t *testing.T, a TestAgainst) {
	cfg := Config{}
	assert.NoError(t, a.run(&cfg))
	assert.Equal(t, "postgres://localhost:5432/db", cfg.DatabaseURL)
}

func testParseStructWithoutEnvTag(t *testing.T, a TestAgainst) {
	cfg := Config{}
	assert.NoError(t, a.run(&cfg))
	assert.Empty(t, cfg.NotAnEnv)
}

func testParseStructWithInvalidFieldKind(t *testing.T, a TestAgainst) {
	type config struct {
		WontWorkByte byte `env:"BLAH"`
	}
	a.setenv("BLAH", "a")
	cfg := config{}
	assert.Error(t, a.run(&cfg))
}

func testUnsupportedSliceType(t *testing.T, a TestAgainst) {
	type config struct {
		WontWork []map[int]int `env:"WONTWORK"`
	}

	a.setenv("WONTWORK", "1,2,3")
	defer os.Clearenv()

	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testBadSeparator(t *testing.T, a TestAgainst) {
	type config struct {
		WontWork []int `env:"WONTWORK" envSeparator:":"`
	}

	cfg := &config{}
	a.setenv("WONTWORK", "1,2,3,4")
	defer os.Clearenv()

	assert.Error(t, a.run(cfg))
}

func testNoErrorRequiredSet(t *testing.T, a TestAgainst) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required"`
	}

	cfg := &config{}

	a.setenv("IS_REQUIRED", "val")
	defer os.Clearenv()
	assert.NoError(t, a.run(cfg))
	assert.Equal(t, "val", cfg.IsRequired)
}

func testErrorRequiredNotSet(t *testing.T, a TestAgainst) {
	type config struct {
		IsRequired string `env:"IS_REQUIRED,required"`
	}

	cfg := &config{}
	assert.Error(t, a.run(cfg))
}

func testCustomParser(t *testing.T, a TestAgainst) {
	type foo struct {
		name string
	}

	type config struct {
		Var foo `env:"VAR"`
	}

	a.setenv("VAR", "test")

	customParserFunc := func(v string) (interface{}, error) {
		return foo{name: v}, nil
	}

	cfg := &config{}
	err := a.runWithFuncs(cfg, map[reflect.Type]ParserFunc{
		reflect.TypeOf(foo{}): customParserFunc,
	})

	assert.NoError(t, err)
	assert.Equal(t, cfg.Var.name, "test")
}

func testParseWithFuncsNoPtr(t *testing.T, a TestAgainst) {
	type foo struct{}
	err := a.runWithFuncs(foo{}, nil)
	assert.Error(t, err)
	assert.Equal(t, err, ErrNotAStructPtr)
}

func testParseWithFuncsInvalidType(t *testing.T, a TestAgainst) {
	var c int
	err := a.runWithFuncs(&c, nil)
	assert.Error(t, err)
	assert.Equal(t, err, ErrNotAStructPtr)
}

func testCustomParserError(t *testing.T, a TestAgainst) {
	type foo struct {
		name string
	}

	type config struct {
		Var foo `env:"VAR"`
	}

	a.setenv("VAR", "test")

	customParserFunc := func(v string) (interface{}, error) {
		return nil, errors.New("something broke")
	}

	cfg := &config{}
	err := a.runWithFuncs(cfg, map[reflect.Type]ParserFunc{
		reflect.TypeOf(foo{}): customParserFunc,
	})

	assert.Empty(t, cfg.Var.name, "Var.name should not be filled out when parse errors")
	assert.Error(t, err)
	assert.Equal(t, err.Error(), "Custom parser error: something broke")
}

func testUnsupportedStructType(t *testing.T, a TestAgainst) {
	type config struct {
		Foo http.Client `env:"FOO"`
	}

	a.setenv("FOO", "foo")

	cfg := &config{}
	err := a.run(cfg)

	assert.Error(t, err)
	assert.Equal(t, ErrUnsupportedType, err)
}
func testEmptyOption(t *testing.T, a TestAgainst) {
	type config struct {
		Var string `env:"VAR,"`
	}

	cfg := &config{}

	a.setenv("VAR", "val")
	defer os.Clearenv()
	assert.NoError(t, a.run(cfg))
	assert.Equal(t, "val", cfg.Var)
}

func testErrorOptionNotRecognized(t *testing.T, a TestAgainst) {
	type config struct {
		Var string `env:"VAR,not_supported!"`
	}

	cfg := &config{}
	assert.Error(t, a.run(cfg))

}

func ExampleParse() {
	type config struct {
		Home         string `env:"HOME"`
		Port         int    `env:"PORT" envDefault:"3000"`
		IsProduction bool   `env:"PRODUCTION"`
	}
	os.Setenv("HOME", "/tmp/fakehome")
	cfg := config{}
	Parse(&cfg)
	fmt.Println(cfg)
	// Output: {/tmp/fakehome 3000 false}
}

func ExampleParseRequiredField() {
	type config struct {
		Home         string `env:"HOME"`
		Port         int    `env:"PORT" envDefault:"3000"`
		IsProduction bool   `env:"PRODUCTION"`
		SecretKey    string `env:"SECRET_KEY,required"`
	}
	os.Setenv("HOME", "/tmp/fakehome")
	cfg := config{}
	err := Parse(&cfg)
	fmt.Println(err)
	// Output: Required environment variable SECRET_KEY is not set
}

func ExampleParseMultipleOptions() {
	type config struct {
		Home         string `env:"HOME"`
		Port         int    `env:"PORT" envDefault:"3000"`
		IsProduction bool   `env:"PRODUCTION"`
		SecretKey    string `env:"SECRET_KEY,required,option1"`
	}
	os.Setenv("HOME", "/tmp/fakehome")
	cfg := config{}
	err := Parse(&cfg)
	fmt.Println(err)
	// Output: Env tag option option1 not supported.
}
