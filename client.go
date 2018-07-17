package xmlrpc

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/pkg/errors"
)

// Client is an XML-RPC client
type Client struct {
	client   *http.Client
	endpoint string
}

// NewClient is an XML-RPC client constructor
func NewClient(endpoint string) *Client {
	return &Client{&http.Client{}, endpoint}
}

func toValue(arg interface{}) (valueizable, error) {
	v := reflect.ValueOf(arg)
	switch v.Kind() {
	case reflect.Bool:
		return newBoolean(v.Bool()), nil
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		fallthrough
	case reflect.Int32:
		return newInt(v.Int()), nil
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		return newDouble(v.Float()), nil
	case reflect.String:
		return newString(v.String()), nil
	case reflect.Struct:
		if v.Type().PkgPath() != "time" || v.Type().Name() != "Time" {
			return nil, errors.Errorf("invalid type %s", v.Kind().String())
		}

		return newDateTime(arg.(time.Time)), nil
	case reflect.Array:
		fallthrough
	case reflect.Slice:
		if v.Type().Elem().Kind() == reflect.Uint8 {
			return newBase64(arg.([]byte)), nil
		}

		return constructArray(v)
	case reflect.Map:
		return constructStruct(v)
	default:
		return nil, errors.Errorf("invalid type %s", v.Kind().String())
	}
}

func constructArray(v reflect.Value) (*array, error) {
	array := newArray()
	for i := 0; i < v.Len(); i++ {
		value, err := toValue(v.Index(i).Interface())
		if err != nil {
			return nil, err
		}
		array.addValue(value)
	}

	return array, nil
}

func constructStruct(v reflect.Value) (*structure, error) {
	s := newStruct()
	for _, k := range v.MapKeys() {
		if k.Kind() != reflect.String {
			return nil, errors.Errorf("invalid type %s", v.Kind().String())
		}

		key := k.String()
		value, err := toValue(v.MapIndex(reflect.ValueOf(key)).Interface())
		if err != nil {
			return nil, err
		}
		s.addMember(key, value)
	}

	return s, nil
}

func (c *Client) preparePayload(methodName string, args ...interface{}) (*bytes.Buffer, error) {
	payload := newPayload(methodName)
	for _, arg := range args {
		value, err := toValue(arg)
		if err != nil {
			return nil, errors.Wrap(err, "method arguments parsing failed")
		}
		payload.addParam(value)
	}

	buffer := new(bytes.Buffer)
	if _, err := payload.WriteTo(buffer); err != nil {
		return nil, errors.Wrap(err, "write to buffer failed")
	}

	return buffer, nil
}

func (c *Client) makeRequest(ctx context.Context, content io.Reader) ([]byte, error) {
	req, err := http.NewRequest("POST", c.endpoint, content)
	if err != nil {
		return nil, errors.Wrap(err, "request preparation failed")
	}

	req.Header.Set("Content-Type", "text/xml")
	res, err := c.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.Wrap(err, "connection error")
	}

	defer func() {
		if err = res.Body.Close(); err != nil {
			logError(errors.Wrap(err, "response body closing failed").Error())
		}
	}()

	if res.StatusCode/100 != 2 {
		return nil, errors.Errorf("response error: code %d", res.StatusCode)
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, "response body read failed")
	}

	return body, nil
}

// Call represents an XML-RPC method call
func (c *Client) Call(ctx context.Context, methodName string, args ...interface{}) (*Result, error) {
	content, err := c.preparePayload(methodName, args...)
	if err != nil {
		return nil, errors.Wrap(err, "payload preparation failed")
	}

	res, err := c.makeRequest(ctx, content)
	if err != nil {
		return nil, errors.Wrap(err, "request failed")
	}

	return parseResult(res)
}
