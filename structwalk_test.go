package structwalk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFieldValue(t *testing.T) {
	assert := assert.New(t)
	v := &SomeStruct{
		Foo: "foo",
		Bar: &struct {
			Baz   int
			Array []int
		}{
			Baz: 5,
		},
	}
	vv, ok := FieldValue("Foo", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal("foo", vv.(string))
	}

	vv, ok = FieldValue("Foo.Bar.Baz", v)
	assert.False(ok)
	assert.Nil(vv)

	vv, ok = FieldValue("Bar.Baz", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal(5, vv.(int))
	}
}

func TestSetFieldValue(t *testing.T) {
	assert := assert.New(t)

	v := &SomeStruct{
		Foo: "foo",
		Bar: &struct {
			Baz   int
			Array []int
		}{
			Baz:   5,
			Array: []int{1, 2},
		},
	}

	{
		SetFieldValue("Foo", "bar", v)

		vv, ok := FieldValue("Foo", v)
		assert.True(ok)
		if assert.NotNil(vv) {
			assert.Equal("bar", vv.(string))
		}
	}

	{
		SetFieldValue("Bar.Baz", 10, v)

		vv, ok := FieldValue("Bar.Baz", v)
		assert.True(ok)
		if assert.NotNil(vv) {
			assert.Equal(10, vv.(int))
		}
	}

	{
		SetFieldValue("Bar.Array", []int{3, 4}, v)

		vv, ok := FieldValue("Bar.Array", v)
		assert.True(ok)
		if assert.NotNil(vv) {
			assert.EqualValues([]int{3, 4}, vv.([]int))
		}
	}
}

func TestFieldValueMap(t *testing.T) {
	assert := assert.New(t)
	v := map[string]interface{}{
		"First": &SomeStruct{
			Foo: "foo",
			Bar: &struct {
				Baz   int
				Array []int
			}{
				Baz: 5,
			},
		},
		"Second": 5,
		"Third": &struct {
			Baz int
		}{
			Baz: 5,
		},
	}
	vv, ok := FieldValue("First.Foo", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal("foo", vv.(string))
	}

	vv, ok = FieldValue("First.Foo.Bar.Baz", v)
	assert.False(ok)
	assert.Nil(vv)

	vv, ok = FieldValue("First.Bar.Baz", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal(5, vv.(int))
	}

	vv, ok = FieldValue("Second", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal(5, vv.(int))
	}

	vv, ok = FieldValue("Third.Baz", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal(5, vv.(int))
	}
}

func TestGetterValue(t *testing.T) {
	assert := assert.New(t)
	v := &SomeDecorated{}

	vv, ok := GetterValue("Foo", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		// expect []byte when getting values of Getters like Foo() string
		assert.EqualValues([]byte("foo"), vv.([]byte))
	}

	vv, ok = GetterValue("Foo.Bar.Baz", v)
	assert.False(ok)
	assert.Nil(vv)

	vv, ok = GetterValue("Bar.Baz", v)
	assert.True(ok)
	if assert.NotNil(vv) {
		assert.Equal(5, vv.(int))
	}
}

func TestFieldList(t *testing.T) {
	assert := assert.New(t)
	foo := &SomeStruct{}
	list := FieldList(foo)
	assert.Equal([]string{
		"Bar.Array",
		"Bar.Baz",
		"Foo",
	}, list)
}

func TestFieldListNoSort(t *testing.T) {
	assert := assert.New(t)
	foo := &SomeStruct{}
	list := FieldListNoSort(foo)
	assert.Equal([]string{
		"Foo",
		"Bar.Baz",
		"Bar.Array",
	}, list)
}

func TestMapFieldList(t *testing.T) {
	assert := assert.New(t)

	foo := map[string]interface{}{
		"Foo":  1,
		"Null": nil,
		"Bar": map[string]interface{}{
			"Baz": 5,
		},
	}
	assert.Equal([]string{
		"Bar.Baz",
		"Foo",
		"Null",
	}, FieldList(foo))

	var str = "aaa"
	bar := map[string]*string{
		"Kek": nil,
		"Lol": &str,
	}
	assert.Equal([]string{
		"Kek",
		"Lol",
	}, FieldList(bar))
}

func TestGetterList(t *testing.T) {
	assert := assert.New(t)
	foo := &SomeDecorated{}
	list := GetterList(foo)
	assert.Equal([]string{
		"Bar.Baz",
		"Foo",
		"FooBytes",
	}, list)
}

type SomeDecorated struct{}

func (s SomeDecorated) Foo() string {
	return "foo"
}

func (s SomeDecorated) FooBytes() []byte {
	return []byte("foo")
}

func (s SomeDecorated) Bar() SomeStruct {
	return SomeStruct{}
}

type SomeStruct struct {
	Foo string
	Bar *struct {
		Baz   int
		Array []int
	}
}

func (s SomeStruct) Baz() int {
	return 5
}
