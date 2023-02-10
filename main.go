package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type TmdbQueryResponseResults struct {
	Poster_path       string   `json:"poster_path"`
	Populatity        float64  `json:"popularity"`
	Id                int      `json:"id"`
	Backdrop_path     string   `json:"backdrop_path"`
	Vote_average      float64  `json:"vote_average"`
	Overview          string   `json:"overview"`
	First_air_date    string   `json:"first_air_date"`
	Origin_country    []string `json:"origin_countr"`
	Genre_ids         []int    `json:"genre_ids"`
	Original_language string   `json:"original_language"`
	Vote_count        int      `json:"vote_count"`
	Name              string   `json:"name"`
	Original_name     string   `json:"original_name"`
}

type TmdbQueryResponse struct {
	Page          int                        `json:"page"`
	Results       []TmdbQueryResponseResults `json:"results"`
	Total_results int                        `json:"total_results"`
	Total_pages   int                        `json:"total_pages"`
}

// TODO: kann raus
type TmdbTvGetDetailsResponseSeasons struct {
	Episode_count int    `json:"episode_count"`
	Id            int    `json:"id"`
	Name          string `json:"name"`
	Overview      string `json:"overview"`
	Season_number int    `json:"season_number"`
}

type TmdbTvGetDetailsResponse struct {
	Number_of_seasons int                               `json:"number_of_seasons"`
	Seasons           []TmdbTvGetDetailsResponseSeasons `json:"seasons"`
}

type TmdbTvGetSeasonDetailsResponseEpisodes struct {
	Episode_number int    `json:"episode_number"`
	Name           string `json:"name"`
	Overview       string `json:"overview"`
}

type TmdbTvGetSeasonDetailsResponse struct {
	Name          string                                   `json:"name"`
	Overview      string                                   `json:"overview"`
	Season_number int                                      `json:"season_number"`
	Episodes      []TmdbTvGetSeasonDetailsResponseEpisodes `json:"episodes"`
}

func main() {
	// get TMDB API key from environment variable
	api_key := os.Getenv("TMDB_API_KEY")
	if api_key == "" {
		fmt.Println("Please set TMDB_API_KEY environment variable")
		return
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		// wait for user input
		fmt.Println("Search for TV show (enter keywords or type 'q' to exit)")
		fmt.Print("> ")
		query, _ := reader.ReadString('\n')
		query = strings.Replace(query, "\r\n", "", -1)
		query = strings.Replace(query, "\n", "", -1)

		// quit program when "q" is entered, restart when nothing is entered
		if strings.Compare("q", query) == 0 {
			fmt.Println("Goodbye!")
			os.Exit(0)
		} else if strings.Compare("", query) == 0 {
			continue
		}

		// search for tv show at tmdb -> "adult=false" because customer wants children to watch
		tmdb_search_url := "https://api.themoviedb.org/3/search/tv?api_key=" + api_key + "&query=" + url.QueryEscape(query) + "&include_adult=false"
		resp, err := http.Get(tmdb_search_url)
		if err != nil {
			log.Fatalln("http request failed: ", err)
		}

		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var queryResponse TmdbQueryResponse
		err = json.Unmarshal(body, &queryResponse)
		if err != nil {
			log.Fatalln(err)
		}

		show_list := queryResponse.Results

		// display all shows found
		for show := range show_list {
			fmt.Println(show+1, ": ", show_list[show].Name)
		}

		if queryResponse.Total_pages > 1 {
			fmt.Println(" -- more than 20 entries found, please add more keywords --")
		}

		// let user select a show
		// TODO: if no valid selection or wrong input: user should re-input a number
		fmt.Println("Which show do you want to watch? Select number (or choose 0 to go back to the search mask)")
		fmt.Print("> ")
		numstr, _ := reader.ReadString('\n')
		numstr = strings.Replace(numstr, "\r\n", "", -1)
		numstr = strings.Replace(numstr, "\n", "", -1)

		num, err := strconv.Atoi(numstr)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if num == 0 {
			fmt.Println()
			continue
		} else if num > 20 || num < 0 {
			fmt.Println("No valid selection")
			fmt.Println()
			continue
		}

		// display overview of selected show or go back to search mask
		fmt.Println("\nOverview:")
		fmt.Println(show_list[num-1].Overview)
		fmt.Println()

		show_id := show_list[num-1].Id

		// get all seasons
		tmdb_search_url = "https://api.themoviedb.org/3/tv/" + strconv.Itoa(show_id) + "?api_key=" + api_key
		resp, err = http.Get(tmdb_search_url)
		if err != nil {
			log.Fatalln("http request failed: ", err)
		}

		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var getDetailsResponse TmdbTvGetDetailsResponse
		err = json.Unmarshal(body, &getDetailsResponse)
		if err != nil {
			log.Fatalln(err)
		}

		num_of_seasons := getDetailsResponse.Number_of_seasons

		for i := 1; i <= num_of_seasons; i++ {
			fmt.Println("SEASON ", i)
			tmdb_search_url = "https://api.themoviedb.org/3/tv/" + strconv.Itoa(show_id) + "/season/" + strconv.Itoa(i) + "?api_key=" + api_key
			resp, err = http.Get(tmdb_search_url)
			if err != nil {
				log.Fatalln("http request failed: ", err)
			}

			defer resp.Body.Close()

			// TODO: does not work yet
			var getSeasonDetailsResponse TmdbTvGetSeasonDetailsResponse
			err = json.Unmarshal(body, &getSeasonDetailsResponse)
			if err != nil {
				log.Fatalln(err)
			}

			fmt.Println(getSeasonDetailsResponse.Episodes)

		}

	}
}
