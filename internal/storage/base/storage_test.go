package base

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStorage_AppendWithHolding(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "storage test#1: append not existed with holding",
			value: "http://notexists.com",
			want:  "5277b20",
		},
		{
			name:  "storage test#2: append existed with holding",
			value: "http://notexists.com",
			want:  "5277b20",
		},
	}

	st := New(WithFile("test.json"))

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shorty, err := st.Push(test.value, "test")
			if assert.NoError(t, err, fmt.Sprintf("%s: error handled", test.name)) {
				assert.Equal(t, test.want, shorty, fmt.Sprintf("%s: unexpected result", test.name))
				assert.Equal(t, test.value, st.uris[shorty], fmt.Sprintf("%s: unexpected result", test.name))
			}
		})
	}

	assert.NoError(t, st.File.Close(), "unexpected error")
	assert.NoError(t, os.Remove("test.json"), "unexpected error")
}

func TestStorage_AppendWithoutHolding(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{
			name:  "storage test#1: append not existed without holding",
			value: "http://notexists.com",
			want:  "5277b20",
		},
		{
			name:  "storage test#2: append existed without holding",
			value: "http://notexists.com",
			want:  "5277b20",
		},
	}

	st := New()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			shorty, err := st.Push(test.value, "test")
			if assert.NoError(t, err, fmt.Sprintf("%s: error handled", test.name)) {
				assert.Equal(t, test.want, shorty, fmt.Sprintf("%s: unexpected result", test.name))
				assert.Equal(t, test.value, st.uris[shorty], fmt.Sprintf("%s: unexpected result", test.name))
			}
		})
	}
}
