package main

import (
	"context"
	"log"
	"strings"

	"github.com/gcottom/musicbrainz"
	spotifyauth "github.com/zmb3/spotify/v2/auth"

	"golang.org/x/oauth2/clientcredentials"

	spotify "github.com/zmb3/spotify/v2"
)

type Meta struct {
	albumImage string
	album      string
	albumID    spotify.ID
	artist     string
	song       string
	trackID    spotify.ID
	genre      string
	year       string
	bpm        string
	infoSource string
}

var resultMeta []Meta
var SongTitle string

func getMetaFromSong(songName string) {
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		log.Fatalf("couldn't get token: %v", err)
	}

	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	searchTerm := "track:" + songName
	results, err := client.Search(ctx, searchTerm, spotify.SearchTypeTrack)
	if err == nil {
		processMeta(results)
	}

}
func getMetaFromSongAndArtist(songName string, artist string) error {
	resultMeta = []Meta{}
	ctx := context.Background()
	config := &clientcredentials.Config{
		ClientID:     spotifyClientID,
		ClientSecret: spotifySecret,
		TokenURL:     spotifyauth.TokenURL,
	}
	token, err := config.Token(ctx)
	if err != nil {
		return err
	}
	searchTerm := "track:" + songName
	if strings.Trim(artist, " ") != "" {
		searchTerm = "track:" + songName + " artist:" + artist
	}
	httpClient := spotifyauth.New().Client(ctx, token)
	client := spotify.New(httpClient)
	results, err := client.Search(ctx, searchTerm, spotify.SearchTypeTrack)
	if err == nil {
		processMeta(results)
	}
	return nil
}

/*
	func getMetaFromSongAndArtistLastFm(songName string, artist string) error {
		api := lastfm.New(lastFmApiKey, lastFmSecret) //lastfm creds
		lfmtoken, _ := api.GetToken()                 //discarding error
		//Send your user to "authUrl"
		//Once the user grant permission, then authorize the token.
		api.LoginWithToken(lfmtoken) //discarding error
		response, err := api.Track.GetInfo(lastfm.P{"track": songName, "artist": artist})
		if err != nil {
			return err
		} else {
			artist := response.Artist.Name
			song := response.Name
			album := response.Album.Title
			var albumImage = ""
			if len(response.Album.Images) > 0 {
				albumImage = response.Album.Images[0].Url
			}
			var trackGenre = ""
			for _, tag := range response.TopTags {
				log.Println("Tags: " + tag.Name)
				for _, genre := range genres {
					g := genre
					if strings.Compare(strings.ToLower(tag.Name), strings.ToLower(genre)) == 0 {
						log.Println("SET TAG:" + tag.Name + ", GENRE:" + genre)
						trackGenre = g
						break
					}
				}
				if trackGenre != "" {
					break
				}
				for _, genre := range genres {
					g := genre
					if strings.Contains(strings.ToLower(tag.Name), strings.ToLower(genre)) || strings.Contains(strings.ToLower(genre), strings.ToLower(tag.Name)) {
						trackGenre = g
						break
					}
				}
				if trackGenre != "" {
					break
				}
			}
			outMeta := Meta{albumImage, album, "", artist, song, "", trackGenre, "", "", "lastfm"}
			resultMeta = []Meta{}
			resultMeta = append(resultMeta, outMeta)
			return nil
		}

}
*/
func getMetaFromSongAndArtistMusicBrainz(song string, artist string) error {
	response, err := musicbrainz.SearchRecordingsByTitleAndArtist(song, artist)
	if err != nil {
		return err
	}
	resultMeta = []Meta{}
	for _, recording := range response {
		album := ""
		if len(recording.Releases) > 0 {
			album = recording.Releases[0].Title
		}
		artist := ""
		if len(recording.ArtistCredit) > 0 {
			artist = recording.ArtistCredit[0].Name
		}
		outMeta := Meta{"", album, "", artist, recording.Title, "", "", "", "", "musicbrainz"}
		resultMeta = append(resultMeta, outMeta)
	}
	return nil
}
func processMeta(results *spotify.SearchResult) {
	resultMeta = []Meta{}
	for _, track := range results.Tracks.Tracks {
		var albumImage = ""
		if len(track.Album.Images) > 0 {
			albumImage = track.Album.Images[0].URL
		}
		album := track.Album.Name
		albumID := track.Album.ID
		artist := ""
		for _, art := range track.Artists {
			artist += art.Name + ", "
		}
		artist = artist[:(strings.LastIndex(artist, ", "))] + strings.Replace(artist[(strings.LastIndex(artist, ", ")):], ", ", "", 1)
		song := track.Name
		trackID := track.ID

		outMeta := Meta{albumImage, album, albumID, artist, song, trackID, "", "", "", "spotify"}
		resultMeta = append(resultMeta, outMeta)
	}
}
