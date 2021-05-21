package gogen

import (
	"fmt"
	"reflect"
)

const (
	sepChar    = ','
	quoteChar  = '\''
	assignChar = '='
)

func extractTags(tag string) (map[string]string, error) {
	gogenTag := reflect.StructTag(tag).Get("gogen")
	tags, err := parseTagItems(gogenTag)
	if err != nil {
		return nil, err
	}
	tagMap := make(map[string]string)

	for k, v := range tags {
		if len(v) == 0 || len(v[0]) == 0 {
			tagMap[k] = "true"
			continue
		}

		tagMap[k] = v[0]
	}

	return tagMap, nil
}

func parseTagItems(tagString string) (map[string][]string, error) {
	d := map[string][]string{}
	key := []rune{}
	value := []rune{}
	quotes := false
	inKey := true

	add := func() {
		d[string(key)] = append(d[string(key)], string(value))
		key = []rune{}
		value = []rune{}
		inKey = true
	}

	runes := []rune(tagString)
	for idx := 0; idx < len(runes); idx++ {
		r := runes[idx]
		next := rune(0)
		eof := false
		if idx < len(runes)-1 {
			next = runes[idx+1]
		} else {
			eof = true
		}
		if !quotes && r == sepChar {
			add()
			continue
		}
		if r == assignChar && inKey {
			inKey = false
			continue
		}
		if r == '\\' {
			if next == quoteChar {
				idx++
				r = quoteChar
			}
		} else if r == quoteChar {
			if quotes {
				quotes = false
				if next == sepChar || eof {
					continue
				}

				return nil, fmt.Errorf("%v has an unexpected char at pos %v", tagString, idx)
			} else {
				quotes = true
				continue
			}
		}
		if inKey {
			key = append(key, r)
		} else {
			value = append(value, r)
		}
	}
	if quotes {
		return nil, fmt.Errorf("%v is not quoted properly", tagString)
	}

	add()

	return d, nil
}
