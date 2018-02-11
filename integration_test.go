package porthos

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func getAmqpUrl() string {
	return os.Getenv("AMQP_URL")
}

type testPorthosResponse struct {
	Original int `json:"original_value"`
	Sum      int `json:"value_plus_one"`
}

type testPorthosRequest struct {
	Value int `json:"value"`
}

func TestClientServerSync(t *testing.T) {
	serverName := "TestClientServer"
	methodName := "method"

	b, err := NewBroker(getAmqpUrl())

	if assert.Nil(t, err) {
		defer b.Close()

		server, err := NewServer(b, serverName, Options{AutoAck: false})
		if assert.Nil(t, err, "Failed to create server.") {
			defer server.Close()
			requestPayload := testPorthosRequest{Value: 1}

			server.Register(methodName, func(req Request, res Response) {
				payload := &testPorthosRequest{}

				assert.Nil(t, req.Bind(payload))

				res.JSON(StatusOK, testPorthosResponse{payload.Value, payload.Value + 1})
			})

			server.Register("dummyMethod", func(req Request, res Response) {
				payload := &testPorthosRequest{}

				assert.Nil(t, req.Bind(payload))

				res.JSON(StatusOK, testPorthosResponse{payload.Value, payload.Value + 2})
			})

			go server.ListenAndServe()

			client, err := NewClient(b, serverName, 1*time.Second)

			if assert.Nil(t, err, "Failed to create client.") {
				defer client.Close()

				ret, err := client.Call(methodName).WithStruct(requestPayload).Sync()

				if assert.Nil(t, err, "Failed to call remote method.") {
					responsePayload := &testPorthosResponse{}

					assert.Equal(t, StatusOK, ret.StatusCode)
					if assert.Nil(t, ret.UnmarshalJSONTo(responsePayload)) {
						assert.Equal(t, requestPayload.Value, responsePayload.Original)
						assert.Equal(t, requestPayload.Value+1, responsePayload.Sum)
					}
				}
			}
		}
	}
}
