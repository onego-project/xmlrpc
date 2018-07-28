package xmlrpc

import (
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"github.com/beevik/etree"
)

const xmlInstructionName = "xml"
const xmlInstruction = `version="1.0" encoding="UTF-8"`

const enMethodCall = "methodCall"
const enMethodName = "methodName"
const enParams = "params"
const enParam = "param"
const enValue = "value"
const enInt = "int"
const enBoolean = "boolean"
const enString = "string"
const enDouble = "double"
const enDateTime = "dateTime.iso8601"
const enBase64 = "base64"
const enMember = "member"
const enName = "name"
const enStruct = "struct"
const enArray = "array"
const enData = "data"

const timeFormat  = "2006-01-02T15:04:05-0700"

type payload struct {
	*etree.Document
}

type value struct {
	*etree.Element
}

type param struct {
	*etree.Element
}

type member struct {
	*etree.Element
}

type structure struct {
	*etree.Element
}

type array struct {
	*etree.Element
}

type scalar struct {
	*etree.Element
}

type valueizable interface {
	toValue() *value
	etree.Token
}

func newPayload(methodName string) *payload {
	p := &payload{etree.NewDocument()}
	p.CreateProcInst(xmlInstructionName, xmlInstruction)
	elMethodCall := p.CreateElement(enMethodCall)
	elMethodName := elMethodCall.CreateElement(enMethodName)
	elMethodName.SetText(methodName)
	elMethodCall.CreateElement(enParams)

	return p
}

func newScalar(typeName string, data string) *scalar {
	elScalar := &scalar{etree.NewElement(typeName)}
	elScalar.SetText(data)

	return elScalar
}

func newInt(data int64) *scalar {
	return newScalar(enInt, strconv.FormatInt(data, 10))
}

func newBoolean(data bool) *scalar {
	if data {
		return newScalar(enBoolean, "1")
	}
	return newScalar(enBoolean, "0")
}

func newString(data string) *scalar {
	return newScalar(enString, data)
}

func newDouble(data float64) *scalar {
	return newScalar(enDouble, strconv.FormatFloat(data, 'f', -1, 64))
}

func newDateTime(data time.Time) *scalar {
	return newScalar(enDateTime, data.UTC().Format(timeFormat))
}

func newBase64(data []byte) *scalar {
	return newScalar(enBase64, base64.StdEncoding.EncodeToString(data))
}

func newStruct() *structure {
	return &structure{etree.NewElement(enStruct)}
}

func newArray() *array {
	elArray := &array{etree.NewElement(enArray)}
	elArray.CreateElement(enData)

	return elArray
}

func (p *payload) addParam(v valueizable) {
	elParams := p.FindElement(fmt.Sprintf("%s/%s", enMethodCall, enParams))
	if elParams == nil {
		panic("'params' element not found")
	}
	elParams.AddChild(v.toValue().toParam())
}

func wrapToValue(v valueizable) *value {
	elValue := &value{etree.NewElement(enValue)}
	elValue.AddChild(v)

	return elValue
}

func (s *scalar) toValue() *value {
	return wrapToValue(s)
}

func newMember(name string, v valueizable) *member {
	elMember := &member{etree.NewElement(enMember)}
	elName := elMember.CreateElement(enName)
	elName.SetText(name)
	elMember.AddChild(v.toValue())

	return elMember
}

func (s *structure) addMember(name string, v valueizable) {
	elMember := newMember(name, v)
	s.AddChild(elMember)
}

func (s *structure) toValue() *value {
	return wrapToValue(s)
}

func (a *array) addValue(v valueizable) {
	elData := a.FindElement(enData)
	if elData == nil {
		panic("'data' element not found")
	}

	elData.AddChild(v.toValue())
}

func (a *array) toValue() *value {
	return wrapToValue(a)
}

func (v *value) toParam() *param {
	elParam := &param{etree.NewElement(enParam)}
	elParam.AddChild(v)

	return elParam
}
