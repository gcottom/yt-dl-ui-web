package main

import (
	"fmt"
	"regexp"
	"strings"
)

func getArtistTitleCombos(filename, author string) map[string][]string {
	filename = strings.ReplaceAll(filename, ":", "-")
	t, c := pSanitize(filename)
	return artistTitleSplit(t, c, author)
}

func pSanitize(s string) (sanitizedTrack, coverArtist string) {
	inparReg := regexp.MustCompile(`\([^)]*\)`)
	inpar := inparReg.FindAllStringSubmatch(s, -1)
	san := inparReg.ReplaceAllString(s, "")
	for _, match := range inpar {
		if strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "albumversion") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "officialmusicvideo") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "liveversion") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "officialvideo") || strings.Contains(strings.ToLower(strings.ReplaceAll(match[0], " ", "")), "officiallyricvideo") {
			continue
		}
		if strings.Contains(strings.ToLower(match[0]), "cover by") || strings.Contains(strings.ToLower(match[0]), "by ") {
			match[0] = string([]byte(match[0])[0 : len(match[0])-1])
			fmt.Println("filename contains \"cover by\" or \"by\"")
			t := strings.Split(match[0], "by")
			return san, t[1]

		}
		if strings.Contains(strings.ToLower(match[0]), "cover") {
			match[0] = string([]byte(match[0])[1 : len(match[0])-1])
			fmt.Println("filename contains \"cover\"")
			fmt.Println("P String:", match[0])
			if strings.HasSuffix(strings.ToLower(match[0]), "cover") {
				return san, match[0]

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
		//cover artist overrides original artist
		if c != "" && len(sp) == 2 {
			m[strings.Trim(sanitizeAuthor(c), " ")] = []string{strings.Trim(sp[0], " "), strings.Trim(sp[1], " ")}
			return m
		}
		//artist - title case
		if c == "" && len(sp) == 2 {
			if strings.EqualFold(sanitizeAuthor(strings.Trim(a, " ")), strings.Trim(sp[0], " ")) {
				m[sanitizeAuthor(strings.Trim(sp[0], " "))] = []string{strings.Trim(sp[1], " "), strings.Trim(sp[0]+"-"+sp[1], " ")}
			} else {
				m[sanitizeAuthor(strings.Trim(sp[0], " "))] = []string{strings.Trim(sp[1], " ")}
				m[sanitizeAuthor(strings.Trim(a, " "))] = []string{strings.Trim(sp[0]+"-"+sp[1], " ")}
			}
			m[sanitizeAuthor(strings.Trim(sp[1], " "))] = []string{strings.Trim(sp[0], " ")}
			return m
		}
		//artist - title-title case or
		//artist-artist - title case
		if c == "" && len(sp) == 3 {
			if strings.EqualFold(sanitizeAuthor(strings.Trim(a, " ")), strings.Trim(sp[0], " ")) {
				m[sanitizeAuthor(strings.Trim(sp[0], " "))] = []string{strings.Trim(sp[1]+"-"+sp[2], " "), strings.Trim(sp[0]+"-"+sp[1]+"-"+sp[2], " ")}
			} else {
				m[sanitizeAuthor(strings.Trim(sp[0], " "))] = []string{strings.Trim(sp[1]+"-"+sp[2], " ")}
				m[sanitizeAuthor(strings.Trim(a, " "))] = []string{strings.Trim(sp[0]+"-"+sp[1]+"-"+sp[2], " ")}
			}
			m[sanitizeAuthor(strings.Trim(sp[0]+"-"+sp[1], " "))] = []string{strings.Trim(sp[2], " ")}
			m[sanitizeAuthor(strings.Trim(sp[1]+"-"+sp[2], " "))] = []string{strings.Trim(sp[0], " ")}
			m[sanitizeAuthor(strings.Trim(sp[2], " "))] = []string{strings.Trim(sp[0]+"-"+sp[1], " ")}
			return m
		}
		//artist - title-title-title
		//artist-artist - title-title
		//artits-artits-artist - title
		if c == "" && len(sp) == 4 {
			m[sanitizeAuthor(strings.Trim(sp[0], " "))] = []string{strings.Trim(sp[1]+"-"+sp[2]+"-"+sp[3], " ")}
			m[sanitizeAuthor(strings.Trim(sp[0]+"-"+sp[1], " "))] = []string{strings.Trim(sp[2]+"-"+sp[3], " ")}
			m[sanitizeAuthor(strings.Trim(sp[2]+"-"+sp[3], " "))] = []string{strings.Trim(sp[0]+"-"+sp[1], " ")}
			m[sanitizeAuthor(strings.Trim(sp[0]+"-"+sp[1]+"-"+sp[2], " "))] = []string{strings.Trim(sp[3], " ")}
			m[sanitizeAuthor(strings.Trim(sp[1]+"-"+sp[2]+"-"+sp[3], " "))] = []string{strings.Trim(sp[0], " ")}
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
