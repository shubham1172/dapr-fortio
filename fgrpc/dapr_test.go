package fgrpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseDaprParameters4InvokeAppCallback(t *testing.T) {
	d := &DaprGRPCRunnerResults{}
	err := d.parseDaprParameters("capability=invoke,target=appcallback,method=load,ex1=1,ex2=2")
	assert.NoError(t, err, "parse failed")
	assert.NotNil(t, d.params)
	assert.Equal(t, d.params.capability, "invoke")
	assert.Equal(t, d.params.target, "appcallback")
	assert.Equal(t, d.params.method, "load")
	assert.Equal(t, d.params.appId, "")
	assert.Equal(t, d.params.extensions["ex1"], "1")
	assert.Equal(t, d.params.extensions["ex2"], "2")
}

func TestParseDaprParameters4InvokeDapr(t *testing.T) {
	d := &DaprGRPCRunnerResults{}
	err := d.parseDaprParameters("capability=invoke,target=dapr,method=load,appid=testapp,ex1=1,ex2=2")
	assert.NoError(t, err, "parse failed")
	assert.NotNil(t, d.params)
	assert.Equal(t, d.params.capability, "invoke")
	assert.Equal(t, d.params.target, "dapr")
	assert.Equal(t, d.params.method, "load")
	assert.Equal(t, d.params.appId, "testapp")
	assert.Equal(t, d.params.extensions["ex1"], "1")
	assert.Equal(t, d.params.extensions["ex2"], "2")
}

func TestPrepareRequest4PubSub(t *testing.T) {
	t.Run("params sanity test", func(t *testing.T) {
		testcases := []struct {
			name    string
			params  string
			isError bool
			errSub  string
		}{
			{
				name:    "method is missing",
				params:  "capability=pubsub,target=dapr,store=memstore,topic=mytopic",
				isError: true,
				errSub:  "method is required",
			},
			{
				name:    "store is missing",
				params:  "capability=pubsub,target=dapr,method=publish,topic=mytopic",
				isError: true,
				errSub:  "store(pubsub name) is required",
			},
			{
				name:    "topic is missing",
				params:  "capability=pubsub,target=dapr,method=publish,store=memstore",
				isError: true,
				errSub:  "topic is required",
			},
			{
				name:    "method is invalid",
				params:  "capability=pubsub,target=dapr,method=invalid,store=memstore,topic=mytopic",
				isError: true,
				errSub:  "unsupported method",
			},
			{
				name:    "when method is publish-multi, numevents is missing",
				params:  "capability=pubsub,target=dapr,method=publish-multi,store=memstore,topic=mytopic",
				isError: true,
				errSub:  "numevents is required",
			},
			{
				name:    "when method is publish-multi, numevents is invalid",
				params:  "capability=pubsub,target=dapr,method=publish-multi,store=memstore,topic=mytopic,numevents=invalid",
				isError: true,
				errSub:  "numevents must be integer",
			},
			{
				name:    "when method is bulkpublish, numevents is missing",
				params:  "capability=pubsub,target=dapr,method=bulkpublish,store=memstore,topic=mytopic",
				isError: true,
				errSub:  "numevents is required",
			},
			{
				name:    "when method is bulkpublish, numevents is invalid",
				params:  "capability=pubsub,target=dapr,method=bulkpublish,store=memstore,topic=mytopic,numevents=invalid",
				isError: true,
				errSub:  "numevents must be integer",
			},
			{
				name:    "valid with method publish",
				params:  "capability=pubsub,target=dapr,method=publish,store=memstore,topic=mytopic",
				isError: false,
			},
			{
				name:    "valid with method publish-multi",
				params:  "capability=pubsub,target=dapr,method=publish-multi,store=memstore,topic=mytopic,numevents=100",
				isError: false,
			},
			{
				name:    "valid with method bulkpublish",
				params:  "capability=pubsub,target=dapr,method=bulkpublish,store=memstore,topic=mytopic,numevents=100",
				isError: false,
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				o := &GRPCRunnerOptions{}
				o.UseDapr = tc.params

				d := &DaprGRPCRunnerResults{}
				err := d.parseDaprParameters(o.UseDapr)
				assert.NoError(t, err, "parse failed")

				err = d.prepareRequest4PubSub(o)
				if tc.isError {
					assert.Error(t, err, "should fail")
				} else {
					assert.NoError(t, err, "should succeed")
				}
			})
		}
	})

	t.Run("publish request", func(t *testing.T) {
		o := &GRPCRunnerOptions{}
		o.UseDapr = "capability=pubsub,target=dapr,method=publish,store=memstore,topic=mytopic,contenttype=text/plain"
		o.Payload = "hello world"

		d := &DaprGRPCRunnerResults{}
		err := d.parseDaprParameters(o.UseDapr)
		assert.NoError(t, err, "parse failed")

		err = d.prepareRequest4PubSub(o)
		assert.NoError(t, err, "prepare failed")

		assert.Equal(t, "memstore", d.publishEventRequest.PubsubName)
		assert.Equal(t, "mytopic", d.publishEventRequest.Topic)
		assert.Equal(t, "text/plain", d.publishEventRequest.DataContentType)
		assert.Equal(t, "hello world", string(d.publishEventRequest.Data))

		assert.Len(t, d.publishEventRequests, 0)
	})

	t.Run("publish multi request", func(t *testing.T) {
		o := &GRPCRunnerOptions{}
		o.UseDapr = "capability=pubsub,target=dapr,method=publish-multi,store=memstore,topic=mytopic,contenttype=text/plain,numevents=100"
		o.Payload = "hello world"

		d := &DaprGRPCRunnerResults{}
		err := d.parseDaprParameters(o.UseDapr)
		assert.NoError(t, err, "parse failed")

		err = d.prepareRequest4PubSub(o)
		assert.NoError(t, err, "prepare failed")

		assert.Len(t, d.publishEventRequests, 100)
		for _, req := range d.publishEventRequests {
			assert.Equal(t, "memstore", req.PubsubName)
			assert.Equal(t, "mytopic", req.Topic)
			assert.Equal(t, "text/plain", req.DataContentType)
			assert.Equal(t, "hello world", string(req.Data))
		}

		assert.Nil(t, d.publishEventRequest)
	})

	t.Run("bulk publish request", func(t *testing.T) {
		o := &GRPCRunnerOptions{}
		o.UseDapr = "capability=pubsub,target=dapr,method=bulkpublish,store=memstore,topic=mytopic,contenttype=text/plain,numevents=100"
		o.Payload = "hello world"

		d := &DaprGRPCRunnerResults{}
		err := d.parseDaprParameters(o.UseDapr)
		assert.NoError(t, err, "parse failed")

		err = d.prepareRequest4PubSub(o)
		assert.NoError(t, err, "prepare failed")

		assert.Equal(t, "memstore", d.bulkPublishRequest.PubsubName)
		assert.Equal(t, "mytopic", d.bulkPublishRequest.Topic)
		for _, entry := range d.bulkPublishRequest.Entries {
			assert.Equal(t, "text/plain", entry.ContentType)
			assert.Equal(t, "hello world", string(entry.Event))
			assert.NotEmpty(t, entry.EntryId)
		}
	})
}

func TestUtils(t *testing.T) {
	t.Run("stringToMap", func(t *testing.T) {
		testcases := []struct {
			name   string
			input  string
			output map[string]string
		}{
			{
				name:   "empty",
				input:  "",
				output: map[string]string{},
			},
			{
				name:   "invalid format",
				input:  "invalid",
				output: map[string]string{},
			},
			{
				name:   "invalid pair",
				input:  "key1=value1,key2=value2,key3",
				output: map[string]string{"key1": "value1", "key2": "value2"},
			},
			{
				name:   "invalid key",
				input:  "key1=value1,key2=value2,=value3",
				output: map[string]string{"key1": "value1", "key2": "value2"},
			},
			{
				name:   "invalid value",
				input:  "key1=value1,key2=value2,key3=",
				output: map[string]string{"key1": "value1", "key2": "value2"},
			},
		}

		for _, tc := range testcases {
			t.Run(tc.name, func(t *testing.T) {
				output := stringToMap(tc.input)
				assert.Equal(t, tc.output, output)
			})
		}
	})
}
