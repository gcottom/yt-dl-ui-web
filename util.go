package main

import (
	"strings"
)

func dialogTextFormat(s string) string {
	if len(s)/lineLimit > 0 {
		split := strings.Split(s, " ")
		temp := ""
		s = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				if temp != "" {
					temp = temp + " " + word
				} else {
					temp = word
				}
				if i+1 == len(split) {
					s = s + "\n" + temp
				}
			} else {
				//temp + word length is over line limit
				if s == "" {
					if len(temp) > lineLimit {
						temp = temp[0:(lineLimit-1)] + "\n" + temp[lineLimit:]
					}
					s = temp
					if len(word) > lineLimit {
						word = word[0:(lineLimit-1)] + "\n" + word[lineLimit:]
					}
					temp = word
				} else {
					if len(temp) > lineLimit {
						temp = temp[0:(lineLimit-1)] + "\n" + temp[lineLimit:]
					}
					s = s + "\n" + temp
					if len(word) > lineLimit {
						word = word[0:(lineLimit-1)] + "\n" + word[lineLimit:]
					}
					temp = word
					if i+1 == len(split) {
						s = s + "\n" + temp
					}
				}
			}
		}
	}
	return s
}
