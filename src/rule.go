package main

import (
	"log"
	"strconv"
	"strings"
)

// a GOL rule
type Rule struct {
	Definition string
	Birth      []uint8
	Death      []uint8
}

// parse one part of a GOL rule into rule slice
func NumbersToList(numbers string) []uint8 {
	list := []uint8{}

	items := strings.Split(numbers, "")
	for _, item := range items {
		num, err := strconv.ParseInt(item, 10, 8)
		if err != nil {
			log.Fatalf("failed to parse game rule part <%s>: %s", numbers, err)
		}

		list = append(list, uint8(num))
	}

	return list
}

// parse GOL rule, used in CheckRule()
func ParseGameRule(rule string) *Rule {
	parts := strings.Split(rule, "/")

	if len(parts) < 2 {
		log.Fatalf("Invalid game rule <%s>", rule)
	}

	golrule := &Rule{Definition: rule}

	for _, part := range parts {
		if part[0] == 'B' {
			golrule.Birth = NumbersToList(part[1:])
		} else {
			golrule.Death = NumbersToList(part[1:])
		}
	}

	return golrule
}
