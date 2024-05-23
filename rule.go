package main

import (
	"log"
	"strconv"
	"strings"
)

// a GOL rule
type Rule struct {
	Birth []int
	Death []int
}

// parse one part of a GOL rule into rule slice
func NumbersToList(numbers string) []int {
	list := []int{}

	items := strings.Split(numbers, "")
	for _, item := range items {
		num, err := strconv.Atoi(item)
		if err != nil {
			log.Fatalf("failed to parse game rule part <%s>: %s", numbers, err)
		}

		list = append(list, num)
	}

	return list
}

// parse GOL rule, used in CheckRule()
func ParseGameRule(rule string) *Rule {
	parts := strings.Split(rule, "/")

	if len(parts) < 2 {
		log.Fatalf("Invalid game rule <%s>", rule)
	}

	golrule := &Rule{}

	for _, part := range parts {
		if part[0] == 'B' {
			golrule.Birth = NumbersToList(part[1:])
		} else {
			golrule.Death = NumbersToList(part[1:])
		}
	}

	return golrule
}
