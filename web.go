package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
)

type getTrackResponse struct {
	TrackUrl string `json:"trackdata,omitempty"`
	Title    string `json:"filename,omitempty"`
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

func generateToken() (string, error) {
	var secretKey = []byte("A Dirty Dummy Secret")
	token := jwt.New(jwt.SigningMethodHS512)
	claims := token.Claims.(jwt.MapClaims)
	claims["exp"] = jwt.NewNumericDate(time.Now().Add(10 * time.Minute))
	claims["authorized"] = true
	claims["user"] = "yt-dl-ui"
	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func getTrack(id string) (string, string, error) {
	id = strings.Replace(id, "&feature=share", "", 1)
	id = strings.Replace(id, "https://music.youtube.com/watch?v=", "", 1)
	id = strings.Replace(id, "https://www.music.youtube.com/watch?v=", "", 1)
	id = strings.Replace(id, "https://www.youtube.com/watch?v", "", 1)
	id = strings.Replace(id, "https://youtube.com/watch?v", "", 1)
	req, err := http.NewRequest(http.MethodGet, "https://api.gagecottom.com/gettrack/"+id, nil)
	if err != nil {
		return "", "", err
	}
	token, err := generateToken()
	if err != nil {
		return "", "", err
	}
	client := &http.Client{}
	req.Header.Add("Authorization", "Bearer "+token)
	res, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", "", err
	}
	var gtr getTrackResponse
	if err = json.Unmarshal(resBody, &gtr); err != nil {
		return "", "", err
	}
	if gtr.Error != "" {
		return "", "", errors.New(gtr.Error)
	}
	s3id := strings.ReplaceAll(gtr.TrackUrl, "https://yt-dl-ui-downloads.s3.us-east-2.amazonaws.com/", "")
	start := time.Now()
	for {
		if time.Now().After(start.Add(2 * time.Minute)) {
			return "", "", errors.New("conversion timed out")
		}
		req, err = http.NewRequest(http.MethodGet, "https://api.gagecottom.com/gettrackconverted/"+s3id, nil)
		if err != nil {
			return "", "", err
		}
		token, err := generateToken()
		if err != nil {
			return "", "", err
		}
		req.Header.Add("Authorization", "Bearer "+token)
		res, err := client.Do(req)
		if err != nil {
			return "", "", err
		}
		rbody, err := io.ReadAll(res.Body)
		if err != nil {
			return "", "", err
		}
		var tres trackConvertedResponse
		if err = json.Unmarshal(rbody, &tres); err != nil {
			return "", "", err
		}
		if tres.TrackConverted {
			break
		}
		time.Sleep(5 * time.Second)

	}
	return gtr.TrackUrl, gtr.Title, nil
}

func saveMeta(m Meta, url string) ([]byte, string, error) {
	reqBody, err := json.Marshal(&setTrackMetaRequest{url, m.song, m.artist, m.album, m.albumImage})
	if err != nil {
		return nil, "", err
	}
	jsonBody := bytes.NewReader(reqBody)
	client := &http.Client{}
	req, err := http.NewRequest(http.MethodPut, "https://api.gagecottom.com/setmeta", jsonBody)
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
