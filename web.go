package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type getTrackResponse struct {
	TrackUrl string `json:"trackdata,omitempty"`
	Title    string `json:"filename,omitempty"`
	Author   string `json:"author,omitempty"`
	Error    string `json:"err,omitempty"`
}
type setTrackMetaRequest struct {
	TrackUrl string `json:"url,omitempty"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	AlbumArt string `json:"albumart"`
}
type setTrackMetaResponse struct {
	FileName string `json:"filename,omitempty"`
	Error    string `json:"err,omitempty"`
}
type trackConvertedResponse struct {
	TrackConverted bool   `json:"converted,omitempty"`
	TrackData      string `json:"trackdata,omitempty"`
	Error          string `json:"error,omitempty"`
}
type getMetaResponse struct {
	Results []metaResult `json:"results,omitempty"`
	Error   string       `json:"error,omitempty"`
}
type metaResult struct {
	Title    string `json:"title,omitempty"`
	Artist   string `json:"artist,omitempty"`
	Album    string `json:"album,omitempty"`
	AlbumArt string `json:"albumart,omitempty"`
}

var oldToken string

func generateToken() (string, error) {
	var secretKey = []byte(jwtSecret)

	claims := jwt.MapClaims{}
	now := time.Now()
	claims["exp"] = jwt.NewNumericDate(now.Add(300 * time.Second))
	claims["iat"] = jwt.NewNumericDate(now)
	claims["nbf"] = jwt.NewNumericDate(now)
	claims["authorized"] = true
	claims["user"] = "yt-dl-ui"
	nonce, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	claims["nonce"] = nonce.String()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	if tokenString == oldToken {
		return generateToken()
	}
	oldToken = tokenString
	return tokenString, nil
}
func sanitizeUrl(id string) (string, error) {
	if strings.Contains(id, "playlist") {
		return "", errors.New("playlist is not currently supported")
	}
	san := strings.Replace(id, "https://", "", 1)
	san = strings.Replace(san, "www.", "", 1)
	san = strings.Replace(san, "music.youtube.com/", "", 1)
	san = strings.Replace(san, "youtube.com/", "", 1)
	san = strings.Replace(san, "youtu.be/", "", 1)
	san = strings.Replace(san, "watch?v=", "", 1)
	san = strings.Replace(san, "&feature=share", "", 1)
	if strings.Contains(san, "&") || strings.Contains(san, "?") {
		sp := strings.Split(san, "&")
		if len(sp) != 2 {
			sp = strings.Split(san, "?")
		}
		san = sp[0]
	}
	if len(san) != 11 {
		return "", errors.New("invalid video id")
	}
	return san, nil
}
func getTrack(id string) (string, string, string, error) {
	san, err := sanitizeUrl(id)
	if err != nil {
		return "", "", "", err
	}
	req, err := http.NewRequest(http.MethodGet, "https://api.gagecottom.com/tracks/"+san, nil)
	if err != nil {
		return "", "", "", err
	}
	token, err := generateToken()
	if err != nil {
		return "", "", "", err
	}
	client := &http.Client{}
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return "", "", "", err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", "", err
	}
	var gtr getTrackResponse
	if err = json.Unmarshal(resBody, &gtr); err != nil {
		return "", "", "", err
	}
	if gtr.Error != "" {
		return "", "", "", errors.New(gtr.Error)
	}
	return gtr.TrackUrl, gtr.Title, gtr.Author, nil
}
func getConverted(trackUrl string) error {
	client := &http.Client{}
	s3id := strings.ReplaceAll(trackUrl, "https://yt-dl-ui-downloads.s3.us-east-2.amazonaws.com/", "")
	start := time.Now()
	for {
		if time.Now().After(start.Add(2 * time.Minute)) {
			return errors.New("conversion timed out")
		}
		req, err := http.NewRequest(http.MethodGet, "https://api.gagecottom.com/converted/"+s3id, nil)
		if err != nil {
			return err
		}
		token, err := generateToken()
		if err != nil {
			return err
		}
		req.Header.Add("Authorization", "Bearer "+token)
		res, err := client.Do(req)
		if err != nil {
			return err
		}
		rbody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}
		var tres trackConvertedResponse
		if err = json.Unmarshal(rbody, &tres); err != nil {
			return err
		}
		if tres.TrackConverted {
			break
		}
		time.Sleep(5 * time.Second)
	}
	return nil
}

func saveMeta(m Meta, url string) ([]byte, string, error) {
	reqBody, err := json.Marshal(&setTrackMetaRequest{url, m.song, m.artist, m.album, m.albumImage})
	if err != nil {
		return nil, "", err
	}
	jsonBody := bytes.NewReader(reqBody)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, "https://api.gagecottom.com/meta", jsonBody)
	if err != nil {
		return nil, "", err
	}
	token, err := generateToken()
	if err != nil {
		return nil, "", err
	}
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return nil, "", err
	}
	resBody, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, "", err
	}
	var stmr setTrackMetaResponse
	if err = json.Unmarshal(resBody, &stmr); err != nil {
		return nil, "", err
	}
	if stmr.Error != "" {
		return nil, "", errors.New(stmr.Error)
	}
	res2, err := http.Get(url)
	if err != nil {
		return nil, "", err
	}
	defer res2.Body.Close()
	data, err := io.ReadAll(res2.Body)
	if err != nil {
		return nil, "", err
	}
	return data, stmr.FileName, nil
}

func getMeta(m map[string][]string) ([]metaResult, error) {
	outResult := []metaResult{}
	for key, value := range m {
		for _, v1 := range value {

			req, err := http.NewRequest(http.MethodGet, "https://api.gagecottom.com/meta/"+url.PathEscape(key)+"/"+url.PathEscape(v1), nil)
			if err != nil {
				return nil, err
			}
			token, err := generateToken()
			if err != nil {
				return nil, err
			}
			client := &http.Client{}
			req.Header.Add("Authorization", "Bearer "+token)
			res, err := client.Do(req)
			if err != nil {
				return nil, err
			}
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			if err != nil {
				return nil, err
			}
			var gmr getMetaResponse
			if err = json.Unmarshal(resBody, &gmr); err != nil {
				return nil, err
			}
			if gmr.Error != "" {
				return nil, err
			}
			outResult = append(outResult, gmr.Results...)
		}
	}
	return outResult, nil

}
