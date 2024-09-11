package godhcpdconfig

import (
	"fmt"
	"github.com/ploynomail/godhcpdconfig/isccfg"
	"io"
	"strconv"
	"strings"
)

type Root struct {
	UserDefinedOptionFields UserDefinedOptionFields
	Options                 Options
	Subnets                 Subnets
}

func NewRoot() *Root {
	return &Root{
		Subnets:                 Subnets{},
		UserDefinedOptionFields: UserDefinedOptionFields{},
		Options: Options{
			Custom: CustomOptions{},
		},
	}
}

func (root Root) ConfigWrite(out io.Writer) (err error) {
	err = root.UserDefinedOptionFields.ConfigWrite(out)
	if err != nil {
		return err
	}
	err = root.Options.ConfigWrite(out, root)
	if err != nil {
		return err
	}
	err = root.Subnets.ConfigWrite(out, root)
	if err != nil {
		return err
	}

	return nil
}

func (root *Root) addUserDefinedOptionField(k string, c *isccfg.Config) (err error) {
	words := c.Unroll()
	name := k
	if len(words) < 3 {
		return fmt.Errorf(`too short: %v`, words)
	}
	if words[1] != "=" {
		return fmt.Errorf(`"=" is expected, got: %v`, words[1])
	}
	code, err := strconv.Atoi(words[0])
	if err != nil {
		return err
	}

	var valueType ValueType
	valueTypeStr := strings.Join(words[2:], " ")
	switch valueTypeStr {
	case "array of integer 8":
		valueType = BYTEARRAY
	case "text":
		valueType = ASCIISTRING
	default:
		panic(fmt.Errorf(`this case is not implemented, yet: %v`, valueTypeStr))
	}

	field := UserDefinedOptionField{
		Code:      code,
		ValueType: valueType,
	}

	root.UserDefinedOptionFields[name] = &field

	return nil
}
