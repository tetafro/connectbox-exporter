package main

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewConnectBox(t *testing.T) {
	t.Run("valid config", func(t *testing.T) {
		client, err := NewConnectBox("127.0.0.1:8080", "bob", "qwerty")
		require.NoError(t, err)
		require.Equal(t, "http://127.0.0.1:8080", client.addr)
		require.Equal(t, "bob", client.username)
		require.Equal(t,
			"65e84be33532fb784c48129675f9eff3a682b27168c0ea744b2cf58ee02337c5",
			client.password)
	})

	t.Run("invalid address", func(t *testing.T) {
		_, err := NewConnectBox("hello, world!", "bob", "qwerty")
		require.ErrorContains(t, err, "invalid address")
	})
}

func TestConnectBox_xmlRequest(t *testing.T) {
	t.Run("valid response", func(t *testing.T) {
		ctx := context.Background()

		connectbox := testConnectBox{
			status: http.StatusOK,
			resp:   "hello, world",
		}
		server := httptest.NewServer(&connectbox)
		defer server.Close()

		client, err := NewConnectBox(server.URL, "bob", "qwerty")
		require.NoError(t, err)
		client.token = "abc"

		args := xmlArgs{{"key", "value"}}
		resp, err := client.xmlRequest(ctx, "/test", "100", args)
		require.NoError(t, err)

		want := "token=abc&fun=100&key=value"
		require.Equal(t, want, connectbox.req)
		want = "hello, world"
		require.Equal(t, want, resp)
	})

	t.Run("invalid status code", func(t *testing.T) {
		ctx := context.Background()

		connectbox := testConnectBox{
			status: http.StatusInternalServerError,
		}
		server := httptest.NewServer(&connectbox)
		defer server.Close()

		client, err := NewConnectBox(server.URL, "bob", "qwerty")
		require.NoError(t, err)
		client.token = "abc"

		args := xmlArgs{{"key", "value"}}
		_, err = client.xmlRequest(ctx, "/test", "100", args)
		require.ErrorContains(t, err, "invalid response status")
	})
}

type testConnectBox struct {
	path   string
	req    string
	status int
	resp   string
}

func (t *testConnectBox) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	t.path = r.URL.Path
	body, err := io.ReadAll(r.Body)
	if err == nil {
		t.req = string(body)
	}
	w.WriteHeader(t.status)
	w.Write([]byte(t.resp))
}
