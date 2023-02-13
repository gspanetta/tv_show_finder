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
	Id                int      `json:"id"`
	Overview          string   `json:"overview"`
	Name              string   `json:"name"`
}

type TmdbQueryResponse struct {
	Page          int                        `json:"page"`
	Results       []TmdbQueryResponseResults `json:"results"`
	Total_results int                        `json:"total_results"`
	Total_pages   int                        `json:"total_pages"`
}

type TmdbTvGetDetailsResponse struct {
	Number_of_seasons int                               `json:"number_of_seasons"`
}

type TmdbTvGetSeasonDetailsResponseEpisodes struct {
	Episode_number int    `json:"episode_number"`
	Name           string `json:"name"`
	Overview       string `json:"overview"`
}

type TmdbTvGetSeasonDetailsResponse struct {
	Episodes []TmdbTvGetSeasonDetailsResponseEpisodes `json:"episodes"`
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
		fmt.Println()
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
			continue
		} else if num > 20 || num < 0 {
			fmt.Println("No valid selection")
			continue
		}

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

		// let user select a season
		fmt.Print("Select Season 1 to ", strconv.Itoa(num_of_seasons), ": ")
		selected_season_str, _ := reader.ReadString('\n')
		selected_season_str = strings.Replace(selected_season_str, "\r\n", "", -1)
		selected_season_str = strings.Replace(selected_season_str, "\n", "", -1)

		// get episodes of the selected season
		tmdb_search_url = "https://api.themoviedb.org/3/tv/" + strconv.Itoa(show_id) + "/season/" + selected_season_str + "?api_key=" + api_key
		fmt.Println("DEBUG: Get ", tmdb_search_url)
		resp, err = http.Get(tmdb_search_url)
		if err != nil {
			log.Fatalln("http request failed: ", err)
		}

		defer resp.Body.Close()

		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var getSeasonDetailsResponse TmdbTvGetSeasonDetailsResponse
		err = json.Unmarshal(body, &getSeasonDetailsResponse)
		if err != nil {
			log.Fatalln(err)
		}

		// display list of episodes
		fmt.Println("Episodes from Season ", selected_season_str, ":")

		for episode := range getSeasonDetailsResponse.Episodes {
			fmt.Println(getSeasonDetailsResponse.Episodes[episode].Episode_number, " - ", getSeasonDetailsResponse.Episodes[episode].Name)
		}

		// let user select episode
		fmt.Print("Select Episode 1 to ", strconv.Itoa(len(getSeasonDetailsResponse.Episodes)), ": ")
		selected_episode_str, _ := reader.ReadString('\n')
		selected_episode_str = strings.Replace(selected_episode_str, "\r\n", "", -1)
		selected_episode_str = strings.Replace(selected_episode_str, "\n", "", -1)

		selected_episode, err := strconv.Atoi(selected_episode_str)
		if err != nil {
			log.Fatalln(err)
		}

		if selected_episode < 1 || selected_episode > len(getSeasonDetailsResponse.Episodes) {
			fmt.Println("there's no such episode!")
			continue
		}

		// display name and overview of the selected episode
		fmt.Println()
		fmt.Println("Overview of Season ", selected_season_str, " Episode ", selected_episode_str, ": ", getSeasonDetailsResponse.Episodes[selected_episode-1].Name)
		fmt.Println("   ", getSeasonDetailsResponse.Episodes[selected_episode-1].Overview)

	}
}
