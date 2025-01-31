package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"text/template"
)

type Artist struct {
	ID           int    `json:"id"`
	Image        string `json:"image"`
	Name         string `json:"name"`
	Members      []string `json:"members"`
	CreationDate int    `json:"creationDate"`
	FirstAlbum   string `json:"firstAlbum"`
}

type Artists []Artist

type Relation struct {
	ID           int    `json:"id"`
	DatesLocations map[string][]string `json:"datesLocations"`
}

func getArtists() ([]Artist, error) {
	url := "https://groupietrackers.herokuapp.com/api/artists"
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, err
	}

	var artists []Artist
	err = json.NewDecoder(response.Body).Decode(&artists)
	return artists, err
}

func getRelationDetails(artistID string) ([]Relation, error) {
	url := "https://groupietrackers.herokuapp.com/api/relation?artist_id=" + artistID
	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	var relations map[string][]Relation
	err = json.NewDecoder(response.Body).Decode(&relations)
	if err != nil {
		return nil, err
	}

	var result []Relation
	for _, relations := range relations {
		for _, relation := range relations {
			id, err := strconv.Atoi(artistID)
			if err != nil {
				return nil, err
			}
			if relation.ID == id {
				result = append(result, Relation{
					ID: relation.ID,
					DatesLocations: relation.DatesLocations,
				})
			}
		}
	}

	return result, nil
}

func handler(w http.ResponseWriter, r *http.Request) {
	artists, err := getArtists()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles("ArtistList.html")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	data := struct {
		Artists []Artist
	}{
		Artists: artists,
	}

	err = tmpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func artistDetailsHandler(w http.ResponseWriter, r *http.Request) error {
	artistID := r.URL.Path[len("/artist/"):]
	artist, err := getArtistDetails(artistID)
	if err != nil {
		return err
	}

	relations, err := getRelationDetails(artistID)
	if err != nil {
		return err
	}

	var filteredRelations []Relation
	for _, relation := range relations {
		if relation.ID == artist.ID {
			filteredRelations = append(filteredRelations, relation)
		}
	}

	tmpl, err := template.ParseFiles("artistDetails.html")
	if err != nil {
		return err
	}

	data := struct {
		Artist   Artist
		Relations []Relation
	}{
		Artist:   artist,
		Relations: filteredRelations,
	}

	return tmpl.Execute(w, data)
}

func getArtistDetails(artistID string) (Artist, error) {
	url := "https://groupietrackers.herokuapp.com/api/artists/" + artistID
	response, err := http.Get(url)
	if err != nil {
		return Artist{}, err
	}
	defer response.Body.Close()

	var artist Artist
	err = json.NewDecoder(response.Body).Decode(&artist)
	return artist, err
}

func main() {
	http.HandleFunc("/artist/", func(w http.ResponseWriter, r *http.Request) {
		err := artistDetailsHandler(w, r)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
	http.HandleFunc("/", handler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}