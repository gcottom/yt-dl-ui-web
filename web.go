package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

type getTrackRequest struct {
	URL string `json:"url"`
}
type getTrackResponse struct {
	TrackUrl string `json:"url,omitempty"`
	Title    string `json:"title,omitempty"`
	Error    string `json:"err,omitempty"`
}
type convertTrackRequest struct {
	TrackUrl string `json:"url"`
	Title    string `json:"title"`
}
type convertTrackResponse struct {
	TrackUrl string `json:"url,omitempty"`
	Error    string `json:"err,omitempty"`
}
type setTrackMetaRequest struct {
	TrackUrl string `json:"url"`
	Title    string `json:"title"`
	Artist   string `json:"artist"`
	Album    string `json:"album"`
	AlbumArt string `json:"albumart"`
}
type setTrackMetaResponse struct {
	TrackData []byte `json:"trackdata,omitempty"`
	FileName  string `json:"filename,omitempty"`
	Error     string `json:"err,omitempty"`
}

func getTrack(url string) (string, string, error) {
	reqBody, err := json.Marshal(&getTrackRequest{url})
	if err != nil {
		return "", "", err
	}
	jsonBody := bytes.NewReader(reqBody)
	res, err := http.Post("http://localhost:8080/track", "application/json", jsonBody)
	if err != nil {
		return "", "", err
	}
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
	return gtr.TrackUrl, gtr.Title, nil
}
func convertTrack(url, title string) (string, error) {
	reqBody, err := json.Marshal(&convertTrackRequest{url, title})
	if err != nil {
		return "", err
	}
	jsonBody := bytes.NewReader(reqBody)
	res, err := http.Post("http://localhost:8080/convert", "application/json", jsonBody)
	if err != nil {
		return "", err
	}
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	var ctr convertTrackResponse
	if err = json.Unmarshal(resBody, &ctr); err != nil {
		return "", err
	}
	if ctr.Error != "" {
		return "", errors.New(ctr.Error)
	}
	return ctr.TrackUrl, nil
}
func saveMeta(m Meta, filepath string) ([]byte, string, error) {
	reqBody, err := json.Marshal(&setTrackMetaRequest{filepath, m.song, m.artist, m.album, m.albumImage})
	if err != nil {
		return nil, "", err
	}
	jsonBody := bytes.NewReader(reqBody)
	res, err := http.Post("http://localhost:8080/meta", "application/json", jsonBody)
	if err != nil {
		return nil, "", err
	}
	resBody, err := io.ReadAll(res.Body)
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
	return stmr.TrackData, stmr.FileName, nil
}
