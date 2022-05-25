package main

import (
	"fmt"
	"github.com/rueian/rueidis"
	"strings"
)

type FtSearchResult struct {
	Key     string
	Results map[string]string
}

// parseFtSearch is a really cursed way of parsing the []RedisMessage returned by a full-text search without the use of rueidis' ObjectMapping library
func parseFtSearch(ms []rueidis.RedisMessage) (int64, []FtSearchResult, error) {
	var err error
	count, _ := ms[0].ToInt64()

	if count == 0 {
		return 0, nil, NotFoundErr
	}

	if count > 10 { // FIXME(george): Discovery should be able to handle over ten results via paging (although this should never be required)
		count = 10
	}

	r := make([]FtSearchResult, count)

	cur := 0
	for i := 1; i < len(ms); {
		r[cur].Key, err = ms[i].ToString()
		if err != nil {
			fmt.Println(err)
			return 0, nil, err
		}

		r[cur].Results, err = ms[i+1].AsStrMap()

		cur++
		i += 2
	}

	return count, r, nil
}

func escapeId(s string) string {
	return strings.ReplaceAll(s, "-", "\\-")
}
