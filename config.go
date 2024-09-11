package godhcpdconfig

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"sort"

	"godhcpdconfig/isccfg"

	"github.com/timtadh/lexmachine"
)

type UserDefinedOptionField struct {
	Code      int
	ValueType ValueType
}

type UserDefinedOptionFields map[string]*UserDefinedOptionField

func (fields *UserDefinedOptionFields) Set(keyName string, codeId int, valueType interface{ ToValueType() ValueType }) {
	(*fields)[keyName] = &UserDefinedOptionField{codeId, valueType.ToValueType()}
}

func (fields UserDefinedOptionFields) ConfigWrite(out io.Writer) (err error) {
	var keys []string
	for k := range fields {
		keys = append(keys, k)
	}
	sort.StringSlice(keys).Sort()

	for _, k := range keys {
		field := fields[k]
		_, err = fmt.Fprintf(out, "option %v code %v = %v;\n", k, field.Code, field.ValueType.ConfigString())
		if err != nil {
			return err
		}
	}

	return err
}

type Config struct {
	Root
	lexer *lexmachine.Lexer
}

func NewConfig() *Config {
	return &Config{
		Root:  *NewRoot(),
		lexer: isccfg.NewLexer(),
	}
}

func (cfg Config) ConfigWriteTo(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}

	cfgWriter := bufio.NewWriter(file)

	defer func() {
		cfgWriter.Write([]byte("\n"))
		cfgWriter.Flush()
		file.Close()
	}()

	cfgWriter.Write([]byte("# auto-generated config\n\n"))
	return cfg.ConfigWrite(cfgWriter)
}

func (cfg *Config) LoadFrom(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}

	cfgReader := bufio.NewReader(file)

	cfgRaw, err := isccfg.Parse(cfgReader)
	if err != nil {
		return err
	}

	cfg.Root = *NewRoot()
	for k, v := range cfgRaw {
		if k == "subnet" {
			continue
		}
		err := cfg.Root.Options.parse(&cfg.Root, k, v)
		if err != nil {
			return err
		}
	}
	for k, v := range cfgRaw {
		if k != "subnet" {
			continue
		}
		for net, netDetails := range *(v.(*isccfg.Config)) {
			newSubnet := *NewSubnet()
			err := newSubnet.parse(&cfg.Root, net, netDetails.(*isccfg.Config))
			if err != nil {
				return err
			}
			cfg.Root.Subnets[net] = newSubnet
		}
	}

	return nil
}

func (cfg Config) ConfigWrite(out io.Writer) error {
	return cfg.Root.ConfigWrite(out)
}
