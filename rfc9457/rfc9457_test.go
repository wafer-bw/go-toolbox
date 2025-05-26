package rfc9457_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/wafer-bw/go-toolbox/rfc9457"
)

func TestProblem_Error(t *testing.T) {
	t.Parallel()

	t.Run("satisfies error.As", func(t *testing.T) {
		t.Parallel()

		err := func() error {
			return rfc9457.Problem{Title: "title", Detail: "detail"}
		}()
		require.True(t, errors.As(err, &rfc9457.Problem{}))
	})

	t.Run("returns error string", func(t *testing.T) {
		t.Parallel()

		pd := rfc9457.Problem{Title: "title", Detail: "detail"}
		require.Equal(t, "title: detail", pd.Error())
	})
}

func TestProblem_Extend(t *testing.T) {
	t.Parallel()

	t.Run("successfully extend problem", func(t *testing.T) {
		t.Parallel()

		pd := rfc9457.Problem{Extensions: map[string]any{"foo": "bar"}}
		expect := rfc9457.Problem{Extensions: map[string]any{"foo": "bar", "baz": "qux"}}
		pd.Extend("baz", "qux")
		require.Equal(t, expect.Extensions, pd.Extensions)
	})

	t.Run("successfully extend problem with nil extensions map", func(t *testing.T) {
		t.Parallel()

		pd := rfc9457.Problem{}
		expect := rfc9457.Problem{Extensions: map[string]any{"foo": "bar"}}
		pd.Extend("foo", "bar")
		require.Equal(t, expect.Extensions, pd.Extensions)
	})

	t.Run("does not extend reserved keys", func(t *testing.T) {
		t.Parallel()

		pd := rfc9457.Problem{}
		pd.Extend("title", "bar")
		require.Nil(t, pd.Extensions)
		pd.Extend("abc", 123)
		pd.Extend("type", "baz")
		require.Equal(t, map[string]any{"abc": 123}, pd.Extensions)
	})
}

func TestProblem_MarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("successfully marshal problem", func(t *testing.T) {
		t.Parallel()

		pd := rfc9457.Problem{
			Type:     "https://example.com/probs/out-of-credit",
			Title:    "You do not have enough credit.",
			Status:   403,
			Detail:   "Your current balance is 30, but that costs 50.",
			Instance: "/account/12345/msgs/abc",
			Extensions: map[string]any{
				"balance": 30,
				"accounts": []string{
					"/account/12345",
					"/account/67890",
				},
				"foo": map[string]int{
					"bar": 1,
				},
				"title": "ignore me",
			},
		}
		expected := `{"type":"https://example.com/probs/out-of-credit","title":"You do not have enough credit.","status":403,"detail":"Your current balance is 30, but that costs 50.","instance":"/account/12345/msgs/abc","balance":30,"accounts":["/account/12345","/account/67890"],"foo":{"bar":1}}`

		b, err := json.Marshal(pd)
		require.NoError(t, err)
		require.JSONEq(t, expected, string(b))
	})

	t.Run("successfully marshal empty problem as empty json object", func(t *testing.T) {
		t.Parallel()

		pd := rfc9457.Problem{}
		expected := `{}`

		b, err := json.Marshal(pd)
		require.NoError(t, err)
		require.JSONEq(t, expected, string(b))
	})
}

func TestProblem_UnmarshalJSON(t *testing.T) {
	t.Parallel()

	t.Run("successfully unmarshal problem", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{
			"type":"https://example.com/probs/out-of-credit",
			"title":"You do not have enough credit.",
			"status":403,
			"detail":"Your current balance is 30, but that costs 50.",
			"instance":"/account/12345/msgs/abc",
			"balance":30,
			"accounts":["/account/12345","/account/67890"],
			"foo":{"bar":1}
		}`)

		expected := rfc9457.Problem{
			Type:     "https://example.com/probs/out-of-credit",
			Title:    "You do not have enough credit.",
			Status:   403,
			Detail:   "Your current balance is 30, but that costs 50.",
			Instance: "/account/12345/msgs/abc",
			Extensions: map[string]any{
				"balance": float64(30),
				"accounts": []any{
					"/account/12345",
					"/account/67890",
				},
				"foo": map[string]any{
					"bar": float64(1),
				},
			},
		}

		pd := rfc9457.Problem{}
		err := json.Unmarshal(raw, &pd)
		require.NoError(t, err)
		require.Equal(t, expected, pd)
	})

	t.Run("successfully unmarshal problem with blank type", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"type":""}`)
		expected := rfc9457.Problem{}

		pd := rfc9457.Problem{}
		err := json.Unmarshal(raw, &pd)
		require.NoError(t, err)
		require.Equal(t, expected, pd)
	})

	t.Run("return error when unmarshalling unsupported JSON type", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`[]`)
		err := json.Unmarshal(raw, &rfc9457.Problem{})
		require.Error(t, err)
	})

	t.Run("return error when unable to unmarshal type string", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"type":-123}`)
		err := json.Unmarshal(raw, &rfc9457.Problem{})
		require.Error(t, err)
	})

	t.Run("return error when unable to unmarshal title", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"title":1}`)
		err := json.Unmarshal(raw, &rfc9457.Problem{})
		require.Error(t, err)
	})

	t.Run("return error when unable to unmarshal status", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"status":"words"}`)
		err := json.Unmarshal(raw, &rfc9457.Problem{})
		require.Error(t, err)
	})

	t.Run("return error when unable to unmarshal detail", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"detail":-123}`)
		err := json.Unmarshal(raw, &rfc9457.Problem{})
		require.Error(t, err)
	})

	t.Run("return error when unable to unmarshal instance string", func(t *testing.T) {
		t.Parallel()

		raw := []byte(`{"instance":-1}`)
		err := json.Unmarshal(raw, &rfc9457.Problem{})
		require.Error(t, err)
	})
}
