package main

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"

	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

var (
	a                    fyne.App
	w                    fyne.Window
	msb                  *widget.Entry
	tsb                  *widget.Entry
	searchMetaWithArtist func(s string)
	searchMeta           func()
	trackSavedNotif      func()
	downloadTrack        func(d string)
	title                string
	OutFile              string
	DLDir                string
	viewContent          *fyne.Container
	lastSearch           time.Time
	searchWaitingChannel chan bool
)

const (
	lineLimit = 50
)

func formatSearchResults(artist string, song string, album string) string {
	if len(artist)/lineLimit > 0 {
		split := strings.Split(artist, " ")
		temp := ""
		artist = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					artist = artist + "\n" + temp
				}
			} else {
				if artist == "" {
					artist = temp
					temp = word
				} else {
					artist = artist + "\n" + temp
					temp = word
					if i+1 == len(split) {
						artist = artist + "\n" + temp
					}
				}
			}
		}
	}
	if len(song)/lineLimit > 0 {
		split := strings.Split(song, " ")
		temp := ""
		song = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					song = song + "\n" + temp
				}
			} else {
				if song == "" {
					song = temp
					temp = word
				} else {
					song = song + "\n" + temp
					temp = word
					if i+1 == len(split) {
						song = song + "\n" + temp
					}
				}
			}
		}
	}
	if len(album)/lineLimit > 0 {
		split := strings.Split(artist, " ")
		temp := ""
		album = ""
		for i, word := range split {
			if (len(temp) + len(word)) < lineLimit {
				temp = temp + " " + word
				if i+1 == len(split) {
					album = album + "\n" + temp
				}
			} else {
				if album == "" {
					album = temp
					temp = word
				} else {
					album = album + "\n" + temp
					temp = word
					if i+1 == len(split) {
						album = album + "\n" + temp
					}
				}
			}
		}
	}
	return "Artist: " + artist + "\n\nTrack: " + song + "\n\nAlbum: " + album
}
func delayedMetaWithArtistSearch(s string) {
	for {
		v := <-searchWaitingChannel
		if v {
			close(searchWaitingChannel)
		}
		if time.Now().After(lastSearch.Add(3 * time.Second)) {
			go searchMetaWithArtist(s)
		}
	}
}
func init() {
	go delayedMetaWithArtistSearch("")
	searchWaitingChannel = make(chan bool)
	searchMetaWithArtist = func(s string) {
		if time.Now().Before(lastSearch.Add(3 * time.Second)) {
			select {
			case searchWaitingChannel <- false:
				return
			default:
				return
			}

		}
		lastSearch = time.Now()
		err := getMetaFromSongAndArtist(tsb.Text, msb.Text)
		if err != nil {
			fmt.Println(err)
		}
		filteredElements := []*fyne.Container{}
		for _, result := range resultMeta {
			if strings.Contains(strings.ToLower(result.artist), strings.ToLower(s)) {
				meta := result
				var img fyne.Resource = nil
				button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), img, func() {
					pb := widget.NewProgressBarInfinite()
					tbSpacer := layout.NewSpacer()
					tbSpacer.Resize(fyne.NewSize(0, 200))
					w.SetContent(container.NewCenter(pb))
					tdata, fname, err := saveMeta(meta, OutFile)
					if err != nil {
						handleError(err)
						showMainScreen()
					}
					jsDownload(tdata, fname)
				})
				hbox := container.New(layout.NewHBoxLayout(), button)
				filteredElements = append(filteredElements, hbox)
			}
		}
		if len(filteredElements) == 0 {
			resultMeta = []Meta{}
			err := getMetaFromSongAndArtistMusicBrainz(title, s)
			if err != nil {
				fmt.Println(err)
			}
			for _, result := range resultMeta {
				if strings.Contains(strings.ToLower(result.artist), strings.ToLower(s)) {
					meta := result
					button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), theme.ErrorIcon(), func() {
						pb := widget.NewProgressBarInfinite()
						tbSpacer := layout.NewSpacer()
						tbSpacer.Resize(fyne.NewSize(0, 200))
						w.SetContent(container.NewCenter(pb))
						tdata, fname, err := saveMeta(meta, OutFile)
						if err != nil {
							handleError(err)
							showMainScreen()
						}
						jsDownload(tdata, fname)
					})
					hbox := container.New(layout.NewHBoxLayout(), button)
					filteredElements = append(filteredElements, hbox)
				}
			}
		}
		addTitleTitle := widget.NewLabel("Track:")
		addArtistTitle := widget.NewLabel("Artist:")
		selectTagsLabel := widget.NewLabel("Select Tags For The Downloaded Track")
		searchTitleBar := tsb
		searchBar := msb
		done := widget.NewButton("Done", func() {
			pb := widget.NewProgressBarInfinite()
			tbSpacer := layout.NewSpacer()
			tbSpacer.Resize(fyne.NewSize(0, 200))
			w.SetContent(container.NewCenter(pb))
			tdata, fname, err := saveMeta(Meta{song: title}, OutFile)
			if err != nil {
				handleError(err)
				showMainScreen()
			}
			jsDownload(tdata, fname)
		})
		cbox := container.New(layout.NewVBoxLayout(), selectTagsLabel, container.NewBorder(nil, nil, addTitleTitle, nil, searchTitleBar), container.NewBorder(nil, nil, addArtistTitle, done, searchBar))
		var vbox *fyne.Container
		if len(filteredElements) == 0 {
			noResults := widget.NewLabel("No Matching Results Found!")
			vbox = container.New(layout.NewVBoxLayout(), cbox, noResults)
		} else {
			vbox = container.New(layout.NewVBoxLayout(), cbox)

			for _, element := range filteredElements {
				vbox.Add(element)
			}
		}
		viewContent = vbox
		w.SetContent(container.NewVScroll(viewContent))
	}
	searchMeta = func() {
		searchWaitingChannel = make(chan bool)
		lastSearch = time.Now().Add(2 * time.Second)
		getMetaFromSong(title)
		var elements []*fyne.Container
		for _, result := range resultMeta {
			meta := result
			var img fyne.Resource = nil
			button := NewCustomButton(formatSearchResults(meta.artist, meta.song, meta.album), img, func() {
				pb := widget.NewProgressBarInfinite()
				tbSpacer := layout.NewSpacer()
				tbSpacer.Resize(fyne.NewSize(0, 200))
				w.SetContent(container.NewCenter(pb))
				tdata, fname, err := saveMeta(meta, OutFile)
				if err != nil {
					handleError(err)
					showMainScreen()
				}
				jsDownload(tdata, fname)
			})
			hbox := container.New(layout.NewHBoxLayout(), button)
			elements = append(elements, hbox)
		}
		addTitleTitle := widget.NewLabel("Track:")
		addArtistTitle := widget.NewLabel("Artist:")
		selectTagsLabel := widget.NewLabel("Select Tags For The Downloaded Track")
		searchTitleBar := tsb
		searchBar := msb
		done := widget.NewButton("Done", func() {
			tdata, fname, err := saveMeta(Meta{song: title}, OutFile)
			if err != nil {
				handleError(err)
				showMainScreen()
			}
			jsDownload(tdata, fname)
		})
		vbox := container.New(layout.NewVBoxLayout(), selectTagsLabel, container.NewBorder(nil, nil, addTitleTitle, nil, searchTitleBar), container.NewBorder(nil, nil, addArtistTitle, done, searchBar))
		for _, element := range elements {
			vbox.Add(element)
		}
		viewContent = vbox
		searchTitleBar.Refresh()
		searchBar.Refresh()
		w.SetContent(container.NewVScroll(viewContent))
	}
	trackSavedNotif = func() {
		dialog.ShowInformation("Complete", "Track Downloaded and MetaData Saved Successfully!", w)
	}
	downloadTrack = func(d string) {
		if strings.TrimSpace(d) != "" {
			showLabeledLoadingScreen("Fetching Track...")
			tempFile, title, author, err := getTrack(d)
			if err != nil {
				fmt.Print("Download Error")
				handleError(err)
				showMainScreen()
				return
			}
			showLabeledLoadingScreen("Converting Track...")
			err = getConverted(tempFile)
			if err != nil {
				fmt.Print("Download Error")
				handleError(err)
				showMainScreen()
				return
			}
			showLabeledLoadingScreen("Fetching Metadata...")
			m := getArtistTitleCombos(title, author)
			results, err := getMeta(m)
			if err != nil {
				fmt.Print("Metadata Search Error")
				handleError(err)
				showMainScreen()
				return
			}
			absolute_match := Meta{}
			absolute_match_found := false
		r_loop:
			for _, r := range results {
				for k, v := range m {
					if strings.EqualFold(k, r.Artist) {
						for _, v1 := range v {
							if strings.EqualFold(v1, r.Title) {
								absolute_match_found = true
								absolute_match.artist = r.Artist
								absolute_match.album = r.Album
								absolute_match.albumImage = r.AlbumArt
								absolute_match.song = r.Title
								fmt.Println("Absolute Match Found")
								break r_loop
							}
						}
					}
				}
			}
			if absolute_match_found {
				showLabeledLoadingScreen("Processing Metadata...")
				tdata, fname, err := saveMeta(Meta{song: absolute_match.song, albumImage: absolute_match.albumImage, album: absolute_match.album, artist: absolute_match.artist}, tempFile)
				if err != nil {
					handleError(err)
					showMainScreen()
				}
				jsDownload(tdata, fname)
			} else {
				tsb.Text = title
				OutFile = tempFile
				searchMeta()
			}

		}
	}

}
func main() {

	appIcon, err := fyne.LoadResourceFromPath("appIcon.png")
	if err != nil {
		fmt.Print(err)
	}
	a = app.NewWithID("yt-dl-ui-web")
	a.SetIcon(appIcon)
	w = a.NewWindow("yt-dl-ui - Youtube Downloader")
	w.SetIcon(appIcon)
	w.Resize(fyne.NewSize(600, 600))
	w.SetFixedSize(true)
	showMainScreen()
	w.ShowAndRun()

}

type customWidget struct {
	widget.Entry
}

func newCustomWidget() *customWidget {
	wid := &customWidget{}
	wid.ExtendBaseWidget(wid)
	return wid
}
func (cw *customWidget) Tapped(ev *fyne.PointEvent) {
	w.Canvas().Focus(cw)
	showKeyboard()
}

func showMainScreen() {
	tsb = widget.NewEntry()
	msb = widget.NewEntry()
	msb.OnChanged = searchMetaWithArtist
	tsb.OnChanged = searchMetaWithArtist
	titleLabel := widget.NewLabel("YT/YT Music URL:")
	//ytId := binding.NewString()
	urlBox := newCustomWidget()
	urlBox.OnSubmitted = downloadTrack
	downloadButton := widget.NewButton("DOWNLOAD", func() {
		id := urlBox.Text
		downloadTrack(id)
	})
	topbox := container.New(layout.NewHBoxLayout(), widget.NewLabel("Download A Track"), layout.NewSpacer())
	hContent := container.New(layout.NewVBoxLayout(), container.New(layout.NewFormLayout(), titleLabel, urlBox), downloadButton)
	vBox := container.New(layout.NewVBoxLayout(), topbox, hContent)
	viewContent = vBox
	w.SetContent(container.NewVScroll(viewContent))
}
func showLabeledLoadingScreen(s string) {
	pb := widget.NewProgressBarInfinite()
	tbSpacer := layout.NewSpacer()
	tbSpacer.Resize(fyne.NewSize(0, 200))
	vb := container.NewVBox(tbSpacer, pb, widget.NewLabel(s))
	w.SetContent(container.NewCenter(vb))
}

func handleError(err error) {
	s := dialogTextFormat(fmt.Sprintf("WELL FUCK!\n%v", err))
	dialog.ShowError(errors.New(s), w)
}
