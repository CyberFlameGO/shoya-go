package api

import "strings"

func boolConvert(s string) bool {
	s = strings.ToLower(s)
	return s == "true"
}

func sliceContains[T comparable](elems []T, v T) bool {
	for _, s := range elems {
		if v == s {
			return true
		}
	}
	return false
}

func dedupeTags(oldTags, newTags []string) []string {
	i := 0
	c := make(map[string]struct{})
	at := append(oldTags, newTags...)

	for _, v := range at {
		c[v] = struct{}{}
	}

	fin := make([]string, len(c))
	for tag := range c {
		fin[i] = tag
		i++
	}
	return fin
}
