package main

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

func getArtistTitleCombos(filename, author string) map[string][]string {
	t, c := pSanitize(getBestPatternMatch(identifyPattern(filename)), filename)
	return artistTitleSplit(t, c, author)
}
func identifyPattern(s string) []string {
	out := make([]string, 0)
	regs := make(map[string]*regexp.Regexp)
	regs["artist-h-title-gen"] = regexp.MustCompile(`^([\w\s]+)-([\w\s]+)$`)
	regs["artist-h-title-specials-no-hp"] = regexp.MustCompile(`^([\w\s.,_&]+)-([\w\s.,_&]+)$`)
	regs["artist-h-title-specials-no-h-wp-l"] = regexp.MustCompile(`^([\w\s.,_&\(\)]+)-([\w\s.,_&]+)$`)
	regs["artist-h-title-specials-no-h-wp-r"] = regexp.MustCompile(`^([\w\s.,_&]+)-([\w\s.,_&\(\)]+)$`)
	regs["artist-h-title-specials-no-h-wp-b"] = regexp.MustCompile(`^([\w\s.,_&\(\)]+)-([\w\s.,_&\(\)]+)$`)
	regs["artist-semi-title-gen"] = regexp.MustCompile(`^([\w\s]+):([\w\s]+)$`)
	regs["artist-semi-title-specials-no-hp"] = regexp.MustCompile(`^([\w\s.,_&]+):([\w\s.,_&]+)$`)
	regs["artist-semi-title-specials-no-h-wp-l"] = regexp.MustCompile(`^([\w\s.,_&\(\)]+):([\w\s.,_&]+)$`)
	regs["artist-semi-title-specials-no-h-wp-r"] = regexp.MustCompile(`^([\w\s.,_&]+):([\w\s.,_&\(\)]+)$`)
	regs["artist-semi-title-specials-no-h-wp-b"] = regexp.MustCompile(`^([\w\s.,_&\(\)]+):([\w\s.,_&\(\)]+)$`)
	for k, v := range regs {
		if v.MatchString(s) {
			out = append(out, k)
		}
	}
	return out
}
func getBestPatternMatch(s []string) string {
	if len(s) < 1 {
		//no patterns matched
		return ""
	}
	sort.Strings(s)
	if len(s) > 1 {
		//contains more than 1 pattern match
		//take the most specific pattern match
		for _, r := range s {
			if regexp.MustCompile(`^.+gen$`).MatchString(r) {
				return r
			}
			if regexp.MustCompile(`^.+hp`).MatchString(r) {
				return r
			}
		}
		//is no-h-wp type
		//the length will be 2, h-wp-b will be at index 0
		//index 1 will have the more specific type
		return s[1]
	}
	//there is only 1 match so return it
	return s[0]
}
func pSanitize(pattern, s string) (sanitizedTrack, coverArtist string) {
	fmt.Println("Most specific match pattern:", pattern)
	rppat := regexp.MustCompile(`^([\w-]+)wp-([rb])$`)
	if !rppat.MatchString(pattern) {
		return s, ""
	}
	inparReg := regexp.MustCompile(`\([^)]*\)`)
	inpar := inparReg.FindAllStringSubmatch(s, -1)
	san := inparReg.ReplaceAllString(s, "")
	for _, match := range inpar {
		if strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "albumversion") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "officialmusicvideo") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "liveversion") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "officialvideo") {
			continue
		}
		if strings.Contains(strings.ToLower(match[0]), "cover by") || strings.Contains(strings.ToLower(match[0]), "by ") {
			match[0] = string([]byte(match[0])[0 : len(match[0])-1])
			fmt.Println("filename contains \"cover by\" or \"by\"")
			t := strings.Split(match[0], "by")
			return strings.ReplaceAll(san, ":", "-"), t[1]

		}
		if strings.Contains(strings.ToLower(match[0]), "cover") {
			match[0] = string([]byte(match[0])[1 : len(match[0])-1])
			fmt.Println("filename contains \"cover\"")
			fmt.Println("P String:", match[0])
			if strings.HasSuffix(strings.ToLower(match[0]), "cover") {
				return strings.ReplaceAll(san, ":", "-"), match[0]

			}
		}

	}
	return san, ""
}
func artistTitleSplit(s, c, a string) map[string][]string {
	m := make(map[string][]string)
	s = strings.ReplaceAll(strings.ReplaceAll(strings.Trim(s, " "), ",", ""), "  ", "")
	c = strings.ReplaceAll(strings.ReplaceAll(strings.Trim(c, " "), ",", ""), "  ", "")
	a = strings.ReplaceAll(strings.ReplaceAll(strings.Trim(a, " "), ",", ""), "  ", "")

	if strings.Contains(s, "-") {
		sp := strings.Split(s, "-")
		for i, p := range sp {
			sp[i] = strings.Trim(p, " ")
		}
		if c == "" {
			m[sp[0]] = []string{sp[1]}
			m[sp[1]] = []string{sp[0]}
		} else {
			m[sanitizeAuthor(c)] = []string{sp[0], sp[1]}
		}
		return m
	}
	m[sanitizeAuthor(a)] = []string{s}
	if c != "" {
		m[c] = []string{s}
	}

	return m

}
func sanitizeAuthor(a string) string {
	a = strings.ToLower(a)
	a = strings.ReplaceAll(a, " - official", "")
	a = strings.ReplaceAll(a, "-official", "")
	a = strings.ReplaceAll(a, "official", "")
	a = strings.ReplaceAll(a, " - vevo", "")
	a = strings.ReplaceAll(a, "-vevo", "")
	a = strings.ReplaceAll(a, "vevo", "")
	a = strings.ReplaceAll(a, "@", "")
	a = strings.ReplaceAll(a, " - topic", "")
	a = strings.ReplaceAll(a, "-topic", "")
	a = strings.ReplaceAll(a, "topic", "")
	a = strings.Trim(a, " ")
	return a
}
