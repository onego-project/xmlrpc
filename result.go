package xmlrpc

import (
	"encoding/base64"
	"strconv"
	"time"

	"github.com/beevik/etree"
	"github.com/pkg/errors"
)

// Kind represents a data types available in XML-RPC standard
type Kind uint

// Constants representing XML-RPC data types
const (
	KindInvalid Kind = iota
	KindArray
	KindBase64
	KindBool
	KindDateTime
	KindDouble
	KindInt
	KindString
	KindStruct
)

const methodResponseValuePath = "methodResponse/params/param/value"
const methodResponseFaultPath = "methodResponse/fault"
const arrayValuePath = "data/value"
const faultMembersPath = "value/struct/member"
const faultCodeName = "faultCode"
const faultStringName = "faultString"
const faultMemberNameTag = "name"
const faultMemberValueIntPath = "value/int"
const faultMemberValueI4Path = "value/i4"
const faultMemberValueStringPath = "value/string"
const structMemberPath = "member"
const structMemberNameTag = "name"
const structMemberValueTag = "value"

// Result represents a return value from XML-RPC method call
type Result struct {
	resString   string
	resInt      int64
	resBoolean  bool
	resDouble   float64
	resDateTime time.Time
	resBase64   []byte
	resStruct   map[string]*Result
	resArray    []*Result
	kind        Kind
}

func parseResult(data []byte) (*Result, error) {
	doc, err := constructXML(data)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse XML RPC response")
	}

	result, err := parseResponse(doc)
	if err != nil {
		return nil, errors.Wrap(err, "cannot parse XML RPC response")
	}

	return result, nil
}

func constructXML(data []byte) (*etree.Document, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromBytes(data); err != nil {
		return nil, errors.Wrap(err, "failed to reconstruct XML DOM")
	}

	return doc, nil
}

func parseResponse(doc *etree.Document) (*Result, error) {
	valueTag := doc.FindElement(methodResponseValuePath)
	faultTag := doc.FindElement(methodResponseFaultPath)
	if (valueTag == nil && faultTag == nil) || (valueTag != nil && faultTag != nil) {
		return nil, errors.Errorf("failed to recognize XML RPC response")
	}

	if faultTag != nil {
		return parseFault(faultTag)
	}

	return parseValue(valueTag)
}

func parseFault(e *etree.Element) (*Result, error) {
	members := e.FindElements(faultMembersPath)
	if len(members) != 2 {
		return nil, errors.Errorf("failed to recognize XML RPC fault")
	}

	var errCode *etree.Element
	var errMsg *etree.Element

	for _, member := range members {
		name := member.FindElement(faultMemberNameTag)
		if name == nil {
			break
		}

		if name.Text() == faultCodeName {
			errCode = member.FindElement(faultMemberValueIntPath)
			if errCode == nil {
				errCode = member.FindElement(faultMemberValueI4Path)
			}
		}

		if name.Text() == faultStringName {
			errMsg = member.FindElement(faultMemberValueStringPath)
		}
	}

	if errCode == nil || errMsg == nil {
		return nil, errors.Errorf("failed to recognize XML RPC fault")
	}

	return nil, errors.Errorf("XML RPC error: %s: %s", errCode.Text(), errMsg.Text())
}

func parseValue(e *etree.Element) (*Result, error) {
	childElements := e.ChildElements()
	if len(childElements) != 1 {
		return nil, errors.Errorf("'value' tag doesn't contain exactly one child tag")
	}

	return parseElement(childElements[0])
}

func parseElement(e *etree.Element) (*Result, error) {
	switch e.Tag {
	case "string":
		return &Result{resString: e.Text(), kind: KindString}, nil
	case "int":
		fallthrough
	case "i4":
		number, err := strconv.Atoi(e.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot convert '%s' to integer", e.Text())
		}
		return &Result{resInt: int64(number), kind: KindInt}, nil
	case "boolean":
		boolean, err := strconv.ParseBool(e.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot convert '%s' to boolean", e.Text())
		}
		return &Result{resBoolean: boolean, kind: KindBool}, nil
	case "double":
		double, err := strconv.ParseFloat(e.Text(), 64)
		if err != nil {
			return nil, errors.Wrapf(err, "cannot convert '%s' to floating point number", e.Text())
		}
		return &Result{resDouble: double, kind: KindDouble}, nil
	case "dateTime.iso8601":
		time, err := time.Parse(time.RFC3339, e.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot convert '%s' to a date", e.Text())
		}
		return &Result{resDateTime: time, kind: KindDateTime}, nil
	case "base64":
		base64, err := base64.StdEncoding.DecodeString(e.Text())
		if err != nil {
			return nil, errors.Wrapf(err, "cannot decode '%s' as base64", e.Text())
		}
		return &Result{resBase64: base64, kind: KindBase64}, nil
	case "array":
		results, err := parseArray(e)
		if err != nil {
			return nil, err
		}
		return &Result{resArray: results, kind: KindArray}, nil
	case "struct":
		results, err := parseStruct(e)
		if err != nil {
			return nil, err
		}
		return &Result{resStruct: results, kind: KindStruct}, nil
	default:
		return nil, errors.Errorf("cannot recognize tag '%s'", e.Tag)
	}
}

func parseArray(e *etree.Element) ([]*Result, error) {
	results := make([]*Result, 0)
	for _, element := range e.FindElements(arrayValuePath) {
		childElements := element.ChildElements()
		if len(childElements) != 1 {
			return nil, errors.Errorf("'value' tag doesn't contain exactly one child tag")
		}
		value, err := parseElement(childElements[0])
		if err != nil {
			return nil, err
		}
		results = append(results, value)
	}

	if len(results) == 0 {
		return nil, errors.Errorf("no values found in array")
	}

	return results, nil
}

func parseStruct(e *etree.Element) (map[string]*Result, error) {
	results := make(map[string]*Result)
	for _, member := range e.FindElements(structMemberPath) {
		name := member.FindElement(structMemberNameTag)
		value := member.FindElement(structMemberValueTag)
		if name == nil {
			return nil, errors.Errorf("no 'name' tag found for struct member")
		}
		if value == nil {
			return nil, errors.Errorf("no 'value' tag found for struct member")
		}
		if results[name.Text()] != nil {
			return nil, errors.Errorf("struct member '%s' found multiple times", name.Text())
		}

		childElements := value.ChildElements()
		if len(childElements) != 1 {
			return nil, errors.Errorf("'value' tag doesn't contain exactly one child tag")
		}
		ret, err := parseElement(childElements[0])
		if err != nil {
			return nil, err
		}
		results[name.Text()] = ret
	}

	if len(results) == 0 {
		return nil, errors.Errorf("no members found in struct")
	}

	return results, nil
}

// ResultString returns a return value from XML-RPC method call of string type
func (r *Result) ResultString() string {
	return r.resString
}

// ResultInt returns a return value from XML-RPC method call of integer type
func (r *Result) ResultInt() int64 {
	return r.resInt
}

// ResultBoolean returns a return value from XML-RPC method call of boolean type
func (r *Result) ResultBoolean() bool {
	return r.resBoolean
}

// ResultDouble returns a return value from XML-RPC method call of float type
func (r *Result) ResultDouble() float64 {
	return r.resDouble
}

// ResultDateTime returns a return value from XML-RPC method call of type representing date and time
func (r *Result) ResultDateTime() time.Time {
	return r.resDateTime
}

// ResultBase64 returns a return value from XML-RPC method call of base64 type
func (r *Result) ResultBase64() []byte {
	return r.resBase64
}

// ResultStruct returns a return value from XML-RPC method call of struct type
func (r *Result) ResultStruct() map[string]*Result {
	return r.resStruct
}

// ResultArray returns a return value from XML-RPC method call of array type
func (r *Result) ResultArray() []*Result {
	return r.resArray
}

// Kind returns the data type of the result
func (r *Result) Kind() Kind {
	return r.kind
}
