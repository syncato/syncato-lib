package storage

import (
	. "gopkg.in/check.v1"
	"testing"
)

func Test(t *testing.T) { TestingT(t) }

type TestSuite struct{}

var _ = Suite(&TestSuite{})

var GetURIFromPathTests = []struct {
	path     string
	expected string
}{
	{"http:/example.org", "http"},
	{"local://abc:abc@some/data", "local"},
	{"swift://abc:abc@example.org/photos", "swift:"},
}

func (s *TestSuite) TestGetURIFromPath(c *C) {
	for _, t := range GetURIFromPathTests {
		uri, err := GetURIFromPath(t.path)
		if err != nil {
			c.Error(err)
		}
		c.Assert(uri.Scheme, Equals, t.expected)
	}
}
