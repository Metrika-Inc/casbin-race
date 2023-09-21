package main

import (
	"regexp"
	"strings"
	"sync"
)

// URLMatcher is an improved version of casbin matcher.
// It lazily initializes regex to check the path URLs against
// and reuses them, thus reducing compute time.
type URLMatcher struct {
	PathletRegex *regexp.Regexp
	regexMap     *sync.Map
}

// NewURLMatcher creates a new instance of URLMatcher.
func NewURLMatcher() *URLMatcher {
	u := &URLMatcher{
		PathletRegex: regexp.MustCompile(`:[^/]+`),
		regexMap:     &sync.Map{},
	}

	return u
}

// Match is a modified implementation of casbin/util.KeyMatch2
// It determines whether key1 matches the pattern of key2 (similar to RESTful path),
// key2 can contain a * and tokens in a format of ":name".
// For example, "/foo/bar" matches "/foo/*", "/resource1" matches "/:resource".
func (u *URLMatcher) Match(key1 string, key2 string) bool {
	regex, ok := u.regexMap.Load(key2)
	if !ok {
		pattern := strings.Replace(key2, "/*", "/.*", -1)
		pattern = u.PathletRegex.ReplaceAllString(pattern, "$1[^/]+$2")
		regex = regexp.MustCompile("^" + pattern + "$")
		u.regexMap.Store(key2, regex)
	}
	r := regex.(*regexp.Regexp)

	return r.MatchString(key1)
}
