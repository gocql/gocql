package gocql

import (
	"bytes"
	"math"
	"math/big"
	"reflect"
	"strings"
	"testing"
	"time"

	"speter.net/go/exp/math/dec/inf"
)

var marshalTests = []struct {
	Info  *TypeInfo
	Data  []byte
	Value interface{}
}{
	{
		&TypeInfo{Type: TypeVarchar},
		[]byte("hello world"),
		[]byte("hello world"),
	},
	{
		&TypeInfo{Type: TypeVarchar},
		[]byte("hello world"),
		"hello world",
	},
	{
		&TypeInfo{Type: TypeVarchar},
		[]byte(nil),
		[]byte(nil),
	},
	{
		&TypeInfo{Type: TypeVarchar},
		[]byte("hello world"),
		MyString("hello world"),
	},
	{
		&TypeInfo{Type: TypeVarchar},
		[]byte("HELLO WORLD"),
		CustomString("hello world"),
	},
	{
		&TypeInfo{Type: TypeBlob},
		[]byte("hello\x00"),
		[]byte("hello\x00"),
	},
	{
		&TypeInfo{Type: TypeBlob},
		[]byte(nil),
		[]byte(nil),
	},
	{
		&TypeInfo{Type: TypeTimeUUID},
		[]byte{0x3d, 0xcd, 0x98, 0x0, 0xf3, 0xd9, 0x11, 0xbf, 0x86, 0xd4, 0xb8, 0xe8, 0x56, 0x2c, 0xc, 0xd0},
		func() UUID {
			x, _ := UUIDFromBytes([]byte{0x3d, 0xcd, 0x98, 0x0, 0xf3, 0xd9, 0x11, 0xbf, 0x86, 0xd4, 0xb8, 0xe8, 0x56, 0x2c, 0xc, 0xd0})
			return x
		}(),
	},
	{
		&TypeInfo{Type: TypeInt},
		[]byte("\x00\x00\x00\x00"),
		0,
	},
	{
		&TypeInfo{Type: TypeInt},
		[]byte("\x01\x02\x03\x04"),
		int(16909060),
	},
	{
		&TypeInfo{Type: TypeInt},
		[]byte("\x80\x00\x00\x00"),
		int32(math.MinInt32),
	},
	{
		&TypeInfo{Type: TypeInt},
		[]byte("\x7f\xff\xff\xff"),
		int32(math.MaxInt32),
	},
	{
		&TypeInfo{Type: TypeBigInt},
		[]byte("\x00\x00\x00\x00\x00\x00\x00\x00"),
		0,
	},
	{
		&TypeInfo{Type: TypeBigInt},
		[]byte("\x01\x02\x03\x04\x05\x06\x07\x08"),
		72623859790382856,
	},
	{
		&TypeInfo{Type: TypeBigInt},
		[]byte("\x80\x00\x00\x00\x00\x00\x00\x00"),
		int64(math.MinInt64),
	},
	{
		&TypeInfo{Type: TypeBigInt},
		[]byte("\x7f\xff\xff\xff\xff\xff\xff\xff"),
		int64(math.MaxInt64),
	},
	{
		&TypeInfo{Type: TypeBoolean},
		[]byte("\x00"),
		false,
	},
	{
		&TypeInfo{Type: TypeBoolean},
		[]byte("\x01"),
		true,
	},
	{
		&TypeInfo{Type: TypeFloat},
		[]byte("\x40\x49\x0f\xdb"),
		float32(3.14159265),
	},
	{
		&TypeInfo{Type: TypeDouble},
		[]byte("\x40\x09\x21\xfb\x53\xc8\xd4\xf1"),
		float64(3.14159265),
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x00\x00"),
		inf.NewDec(0, 0),
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x00\x64"),
		inf.NewDec(100, 0),
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x02\x19"),
		decimalize("0.25"),
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x13\xD5\a;\x20\x14\xA2\x91"),
		decimalize("-0.0012095473475870063"), // From the iconara/cql-rb test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x13*\xF8\xC4\xDF\xEB]o"),
		decimalize("0.0012095473475870063"), // From the iconara/cql-rb test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x12\xF2\xD8\x02\xB6R\x7F\x99\xEE\x98#\x99\xA9V"),
		decimalize("-1042342234234.123423435647768234"), // From the iconara/cql-rb test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\r\nJ\x04\"^\x91\x04\x8a\xb1\x18\xfe"),
		decimalize("1243878957943.1234124191998"), // From the datastax/python-driver test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x06\xe5\xde]\x98Y"),
		decimalize("-112233.441191"), // From the datastax/python-driver test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x14\x00\xfa\xce"),
		decimalize("0.00000000000000064206"), // From the datastax/python-driver test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\x00\x00\x00\x14\xff\x052"),
		decimalize("-0.00000000000000064206"), // From the datastax/python-driver test suite
	},
	{
		&TypeInfo{Type: TypeDecimal},
		[]byte("\xff\xff\xff\x9c\x00\xfa\xce"),
		inf.NewDec(64206, -100), // From the datastax/python-driver test suite
	},
	{
		&TypeInfo{Type: TypeTimestamp},
		[]byte("\x00\x00\x01\x40\x77\x16\xe1\xb8"),
		time.Date(2013, time.August, 13, 9, 52, 3, 0, time.UTC),
	},
	{
		&TypeInfo{Type: TypeTimestamp},
		[]byte("\x00\x00\x01\x40\x77\x16\xe1\xb8"),
		int64(1376387523000),
	},
	{
		&TypeInfo{Type: TypeList, Elem: &TypeInfo{Type: TypeInt}},
		[]byte("\x00\x02\x00\x04\x00\x00\x00\x01\x00\x04\x00\x00\x00\x02"),
		[]int{1, 2},
	},
	{
		&TypeInfo{Type: TypeList, Elem: &TypeInfo{Type: TypeInt}},
		[]byte("\x00\x02\x00\x04\x00\x00\x00\x01\x00\x04\x00\x00\x00\x02"),
		[2]int{1, 2},
	},
	{
		&TypeInfo{Type: TypeSet, Elem: &TypeInfo{Type: TypeInt}},
		[]byte("\x00\x02\x00\x04\x00\x00\x00\x01\x00\x04\x00\x00\x00\x02"),
		[]int{1, 2},
	},
	{
		&TypeInfo{Type: TypeSet, Elem: &TypeInfo{Type: TypeInt}},
		[]byte(nil),
		[]int(nil),
	},
	{
		&TypeInfo{Type: TypeMap,
			Key:  &TypeInfo{Type: TypeVarchar},
			Elem: &TypeInfo{Type: TypeInt},
		},
		[]byte("\x00\x01\x00\x03foo\x00\x04\x00\x00\x00\x01"),
		map[string]int{"foo": 1},
	},
	{
		&TypeInfo{Type: TypeMap,
			Key:  &TypeInfo{Type: TypeVarchar},
			Elem: &TypeInfo{Type: TypeInt},
		},
		[]byte(nil),
		map[string]int(nil),
	},
	{
		&TypeInfo{Type: TypeList, Elem: &TypeInfo{Type: TypeVarchar}},
		bytes.Join([][]byte{
			[]byte("\x00\x01\xFF\xFF"),
			bytes.Repeat([]byte("X"), 65535)}, []byte("")),
		[]string{strings.Repeat("X", 65535)},
	},
	{
		&TypeInfo{Type: TypeMap,
			Key:  &TypeInfo{Type: TypeVarchar},
			Elem: &TypeInfo{Type: TypeVarchar},
		},
		bytes.Join([][]byte{
			[]byte("\x00\x01\xFF\xFF"),
			bytes.Repeat([]byte("X"), 65535),
			[]byte("\xFF\xFF"),
			bytes.Repeat([]byte("Y"), 65535)}, []byte("")),
		map[string]string{
			strings.Repeat("X", 65535): strings.Repeat("Y", 65535),
		},
	},
	{
		&TypeInfo{Type: TypeVarint},
		[]byte("\x00"),
		0,
	},
	{
		&TypeInfo{Type: TypeVarint},
		[]byte("\x37\xE2\x3C\xEC"),
		int32(937573612),
	},
	{
		&TypeInfo{Type: TypeVarint},
		[]byte("\x37\xE2\x3C\xEC"),
		big.NewInt(937573612),
	},
	{
		&TypeInfo{Type: TypeVarint},
		[]byte("\xC0\x00\x00\x00\x00\x00\x00\x00"),
		big.NewInt(-4611686018427387904),
	},
	{
		&TypeInfo{Type: TypeVarint},
		[]byte("\x08\x00\x00\x00\x00\x00\x00\x00\x00\x00"),
		bigintize("37778931862957161709568"),
	},
}

func decimalize(s string) *inf.Dec {
	i, _ := new(inf.Dec).SetString(s)
	return i
}

func bigintize(s string) *big.Int {
	i, _ := new(big.Int).SetString(s, 10)
	return i
}

func TestMarshal(t *testing.T) {
	for i, test := range marshalTests {
		data, err := Marshal(test.Info, test.Value)
		if err != nil {
			t.Errorf("marshalTest[%d]: %v", i, err)
			continue
		}
		if !bytes.Equal(data, test.Data) {
			t.Errorf("marshalTest[%d]: expected %q, got %q.", i, test.Data, data)
		}
	}
}

func TestUnmarshal(t *testing.T) {
	for i, test := range marshalTests {
		v := reflect.New(reflect.TypeOf(test.Value))
		err := Unmarshal(test.Info, test.Data, v.Interface())
		if err != nil {
			t.Errorf("marshalTest[%d]: %v", i, err)
			continue
		}
		if !reflect.DeepEqual(v.Elem().Interface(), test.Value) {
			t.Errorf("marshalTest[%d]: expected %#v, got %#v.", i, test.Value, v.Elem().Interface())
		}
	}
}

func TestMarshalVarint(t *testing.T) {
	varintTests := []struct {
		Value       interface{}
		Marshaled   []byte
		Unmarshaled *big.Int
	}{
		{
			Value:       int8(0),
			Marshaled:   []byte("\x00"),
			Unmarshaled: big.NewInt(0),
		},
		{
			Value:       uint8(255),
			Marshaled:   []byte("\x00\xFF"),
			Unmarshaled: big.NewInt(255),
		},
		{
			Value:       int8(-1),
			Marshaled:   []byte("\xFF"),
			Unmarshaled: big.NewInt(-1),
		},
		{
			Value:       big.NewInt(math.MaxInt32),
			Marshaled:   []byte("\x7F\xFF\xFF\xFF"),
			Unmarshaled: big.NewInt(math.MaxInt32),
		},
		{
			Value:       big.NewInt(int64(math.MaxInt32) + 1),
			Marshaled:   []byte("\x00\x80\x00\x00\x00"),
			Unmarshaled: big.NewInt(int64(math.MaxInt32) + 1),
		},
		{
			Value:       big.NewInt(math.MinInt32),
			Marshaled:   []byte("\x80\x00\x00\x00"),
			Unmarshaled: big.NewInt(math.MinInt32),
		},
		{
			Value:       big.NewInt(int64(math.MinInt32) - 1),
			Marshaled:   []byte("\xFF\x7F\xFF\xFF\xFF"),
			Unmarshaled: big.NewInt(int64(math.MinInt32) - 1),
		},
		{
			Value:       math.MinInt64,
			Marshaled:   []byte("\x80\x00\x00\x00\x00\x00\x00\x00"),
			Unmarshaled: big.NewInt(math.MinInt64),
		},
		{
			Value:       uint64(math.MaxInt64) + 1,
			Marshaled:   []byte("\x00\x80\x00\x00\x00\x00\x00\x00\x00"),
			Unmarshaled: bigintize("9223372036854775808"),
		},
		{
			Value:       bigintize("2361183241434822606848"), // 2**71
			Marshaled:   []byte("\x00\x80\x00\x00\x00\x00\x00\x00\x00\x00"),
			Unmarshaled: bigintize("2361183241434822606848"),
		},
		{
			Value:       bigintize("-9223372036854775809"), // -2**63 - 1
			Marshaled:   []byte("\xFF\x7F\xFF\xFF\xFF\xFF\xFF\xFF\xFF"),
			Unmarshaled: bigintize("-9223372036854775809"),
		},
	}

	for i, test := range varintTests {
		data, err := Marshal(&TypeInfo{Type: TypeVarint}, test.Value)
		if err != nil {
			t.Errorf("error marshaling varint: %v (test #%d)", err, i)
		}

		if !bytes.Equal(test.Marshaled, data) {
			t.Errorf("marshaled varint mismatch: expected %v, got %v (test #%d)", test.Marshaled, data, i)
		}

		binder := new(big.Int)
		err = Unmarshal(&TypeInfo{Type: TypeVarint}, test.Marshaled, binder)
		if err != nil {
			t.Errorf("error unmarshaling varint: %v (test #%d)", err, i)
		}

		if test.Unmarshaled.Cmp(binder) != 0 {
			t.Errorf("unmarshaled varint mismatch: expected %v, got %v (test #%d)", test.Unmarshaled, binder, i)
		}
	}
}

type CustomString string

func (c CustomString) MarshalCQL(info *TypeInfo) ([]byte, error) {
	return []byte(strings.ToUpper(string(c))), nil
}
func (c *CustomString) UnmarshalCQL(info *TypeInfo, data []byte) error {
	*c = CustomString(strings.ToLower(string(data)))
	return nil
}

type MyString string

type MyInt int

var typeLookupTest = []struct {
	TypeName     string
	ExpectedType Type
}{
	{"AsciiType", TypeAscii},
	{"LongType", TypeBigInt},
	{"BytesType", TypeBlob},
	{"BooleanType", TypeBoolean},
	{"CounterColumnType", TypeCounter},
	{"DecimalType", TypeDecimal},
	{"DoubleType", TypeDouble},
	{"FloatType", TypeFloat},
	{"Int32Type", TypeInt},
	{"DateType", TypeTimestamp},
	{"UUIDType", TypeUUID},
	{"UTF8Type", TypeVarchar},
	{"IntegerType", TypeVarint},
	{"TimeUUIDType", TypeTimeUUID},
	{"InetAddressType", TypeInet},
	{"MapType", TypeMap},
	{"ListType", TypeInet},
	{"SetType", TypeInet},
	{"unknown", TypeCustom},
}

func testType(t *testing.T, cassType string, expectedType Type) {
	if computedType := getApacheCassandraType(apacheCassandraTypePrefix + cassType); computedType != expectedType {
		t.Errorf("Cassandra custom type lookup for %s failed. Expected %s, got %s.", cassType, expectedType.String(), computedType.String())
	}
}

func TestLookupCassType(t *testing.T) {
	for _, lookupTest := range typeLookupTest {
		testType(t, lookupTest.TypeName, lookupTest.ExpectedType)
	}
}
