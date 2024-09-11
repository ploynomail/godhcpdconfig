package godhcpdconfig

import (
	"fmt"
	"io"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/ploynomail/godhcpdconfig/isccfg"
)

type CustomOptions map[int][]byte

type Options struct {
	DefaultLeaseTime  int           `json:",omitempty"`
	MaxLeaseTime      int           `json:",omitempty"`
	Authoritative     bool          `json:",omitempty"`
	LogFacility       string        `json:",omitempty"`
	DomainName        string        `json:",omitempty"`
	DomainNameServers NSs           `json:",omitempty"`
	Range             Range         `json:",omitempty"`
	Routers           []string      `json:",omitempty"`
	BroadcastAddress  string        `json:",omitempty"`
	NextServer        string        `json:",omitempty"`
	Filename          string        `json:",omitempty"`
	RootPath          string        `json:",omitempty"`
	MTU               int           `json:",omitempty"`
	Custom            CustomOptions `json:",omitempty"`
}

func (options Options) configWrite(out io.Writer, root Root, indent string) (err error) {
	if options.DefaultLeaseTime != 0 {
		_, err = fmt.Fprintf(out, "%vdefault-lease-time %v;\n", indent, options.DefaultLeaseTime)
		if err != nil {
			return err
		}
	}
	if options.MaxLeaseTime != 0 {
		fmt.Fprintf(out, "%vmax-lease-time %v;\n", indent, options.MaxLeaseTime)
	}
	if options.Authoritative {
		fmt.Fprintf(out, "%vauthoritative;\n", indent)
	}
	if options.LogFacility != "" {
		fmt.Fprintf(out, "%vlog-facility %v;\n", indent, options.LogFacility)
	}
	if options.DomainName != "" {
		fmt.Fprintf(out, "%voption domain-name \"%v\";\n", indent, options.DomainName)
	}
	if len(options.DomainNameServers) > 0 {
		fmt.Fprintf(out, "%voption domain-name-servers %v;\n", indent, strings.Join(options.DomainNameServers.ToStrings(), ", "))
	}
	if options.Range.Start != nil {
		fmt.Fprintf(out, "%vrange %v %v;\n", indent, options.Range.Start, options.Range.End)
	}
	if len(options.Routers) > 0 {
		fmt.Fprintf(out, "%voption routers %v;\n", indent, strings.Join(options.Routers, ", "))
	}
	if options.BroadcastAddress != "" {
		fmt.Fprintf(out, "%voption broadcast-address %v;\n", indent, options.BroadcastAddress)
	}
	if options.NextServer != "" {
		fmt.Fprintf(out, "%vnext-server %v;\n", indent, options.NextServer)
	}
	if options.Filename != "" {
		fmt.Fprintf(out, "%vfilename \"%v\";\n", indent, options.Filename)
	}
	if options.RootPath != "" {
		fmt.Fprintf(out, "%voption root-path \"%v\";\n", indent, options.RootPath)
	}
	if options.MTU != 0 {
		fmt.Fprintf(out, "%voption interface-mtu %v;\n", indent, options.MTU)
	}

	var keys []int
	for k := range options.Custom {
		keys = append(keys, k)
	}
	sort.IntSlice(keys).Sort()

	customOptionNameMap := map[int]string{}
	for k, f := range root.UserDefinedOptionFields {
		customOptionNameMap[f.Code] = k
	}
	customOptionValueTypeMap := map[int]ValueType{}
	for _, f := range root.UserDefinedOptionFields {
		customOptionValueTypeMap[f.Code] = f.ValueType
	}

	for _, k := range keys {
		if customOptionNameMap[k] == "" {
			customOptionNameMap[k] = fmt.Sprintf("option%v", k)
		}
		if customOptionValueTypeMap[k] == 0 {
			if IsAsciiString(string(options.Custom[k])) {
				customOptionValueTypeMap[k] = ASCIISTRING
			} else {
				customOptionValueTypeMap[k] = BYTEARRAY
			}

			if indent != "" {
				panic("This case is not implemented, yet")
			}
			err := UserDefinedOptionFields{customOptionNameMap[k]: &UserDefinedOptionField{Code: k, ValueType: customOptionValueTypeMap[k]}}.ConfigWrite(out)
			if err != nil {
				panic(err)
			}
		}
	}

	for _, k := range keys {
		option := options.Custom[k]

		var valueString string
		switch customOptionValueTypeMap[k] {
		case BYTEARRAY:
			var result []string
			for _, byteValue := range option {
				result = append(result, strconv.Itoa(int(byteValue)))
			}
			valueString = strings.Join(result, ", ")
		case ASCIISTRING:
			valueString = `"` + string(option) + `"`
		default:
			panic(fmt.Errorf("this shouldn't happened: %v: %v: %v, %v: %v", k, customOptionValueTypeMap[k], customOptionValueTypeMap, root.UserDefinedOptionFields, string(option)))
		}

		fmt.Fprintf(out, "%voption %v %v;\n", indent, customOptionNameMap[k], valueString)
	}

	return nil
}

func (options Options) ConfigWrite(out io.Writer, root Root) error {
	return options.configWrite(out, root, "")
}

func (options *Options) parse(root *Root, k string, value isccfg.Value) (err error) {
	cfgRaw, _ := value.(*isccfg.Config)
	if k == "_value" {
		k = value.([]string)[0]
	}

	switch k {
	case "authoritative":
		options.Authoritative = true
	case "default-lease-time":
		options.DefaultLeaseTime, err = strconv.Atoi(cfgRaw.Values()[0])
	case "max-lease-time":
		options.MaxLeaseTime, err = strconv.Atoi(cfgRaw.Values()[0])
	case "log-facility":
		options.LogFacility = cfgRaw.Values()[0]
	case "ddns-update-style":
		// TODO: implement this
	case "filename":
		options.Filename = cfgRaw.Values()[0]
	case "next-server":
		options.NextServer = cfgRaw.Values()[0]

	case "range":
		var startStr, endStr string
		cfgRaw, startStr = cfgRaw.Unwrap()
		for startStr == "dynamic-bootp" {
			cfgRaw, startStr = cfgRaw.Unwrap()
		}
		endStr = cfgRaw.Values()[0]
		options.Range.Start = net.ParseIP(startStr)
		options.Range.End = net.ParseIP(endStr)

	case "option":
		for k, v := range *cfgRaw {
			c, ok := v.(*isccfg.Config)
			if !ok {
				panic(fmt.Errorf("\"v\" (%T) is not *isccfg.Config; k == %v", v, k))
			}
			switch k {
			case "domain-name":
				options.DomainName = c.Values()[0]
			case "domain-name-servers":
				options.DomainNameServers.Set(c.Values())
			case "broadcast-address":
				options.BroadcastAddress = c.Values()[0]
			case "routers":
				options.Routers = c.Values()
			case "root-path":
				options.RootPath = c.Values()[0]
			case "interface-mtu":
				options.MTU, err = strconv.Atoi(c.Values()[0])
			case "static-routes":

			default:
				field := root.UserDefinedOptionFields[k]
				if field == nil {
					c, codeWord := c.Unwrap()
					if codeWord != "code" {
						fmt.Fprintf(os.Stderr, "Not recognized option: %v\n", k)
						break
					}
					err := root.addUserDefinedOptionField(k, c)
					if err != nil {
						return err
					}
					field = root.UserDefinedOptionFields[k]
				}

				bytesStr := c.Values()
				switch field.ValueType {
				case BYTEARRAY:
					var result []byte
					for _, str := range bytesStr {
						oneByte, err := strconv.Atoi(str)
						if err != nil {
							return err
						}
						result = append(result, byte(oneByte))
					}
					options.Custom[field.Code] = result
				case ASCIISTRING:
					options.Custom[field.Code] = []byte(strings.Join(bytesStr, ", "))
				default:
					panic("This shouldn't happened")
				}

			}
		}
	default:
		fmt.Fprintf(os.Stderr, "Not recognized: %v\n", k)
	}

	return
}
