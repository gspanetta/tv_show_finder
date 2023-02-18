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
	"reflect"
	"strconv"
	"strings"
)

type TmdbQueryResponseResults struct {
	Id       int    `json:"id"`
	Overview string `json:"overview"`
	Name     string `json:"name"`
}

type TmdbQueryResponse struct {
	Page          int                        `json:"page"`
	Results       []TmdbQueryResponseResults `json:"results"`
	Total_results int                        `json:"total_results"`
	Total_pages   int                        `json:"total_pages"`
}

type TmdbTvGetDetailsResponse struct {
	Number_of_seasons int `json:"number_of_seasons"`
}

type TmdbTvGetSeasonDetailsResponseEpisodes struct {
	Episode_number int    `json:"episode_number"`
	Name           string `json:"name"`
	Overview       string `json:"overview"`
}

type TmdbTvGetSeasonDetailsResponse struct {
	Episodes []TmdbTvGetSeasonDetailsResponseEpisodes `json:"episodes"`
}

func getUserInputStr(prompt string) string {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Println(prompt)
	fmt.Print("> ")
	query, _ := reader.ReadString('\n')
	query = strings.Replace(query, "\r\n", "", -1)
	query = strings.Replace(query, "\n", "", -1)

	return query
}

func getUserInputNum(prompt string) (int, error) {
	reader := bufio.NewReader(os.Stdin)
	fmt.Println()
	fmt.Println(prompt)
	fmt.Print("> ")
	query, _ := reader.ReadString('\n')
	query = strings.Replace(query, "\r\n", "", -1)
	query = strings.Replace(query, "\n", "", -1)

	num, err := strconv.Atoi(query)
	if err != nil {
		log.Fatalln(err)
		return 0, err
	}

	return num, nil
}

func httpRequest(uri string, sink interface{}) error {
	v := reflect.ValueOf(sink)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("sink is not a struct pointer") // TODO proper error handling
	}

	resp, err := http.Get(uri)
	if err != nil {
		log.Fatalln("http request failed: ", err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	//var queryResponse TmdbQueryResponse
	err = json.Unmarshal(body, sink)
	if err != nil {
		log.Fatalln(err)
	}

	return nil
}

func main() {
	// get TMDB API key from environment variable
	api_key := os.Getenv("TMDB_API_KEY")
	if api_key == "" {
		fmt.Println("Please set TMDB_API_KEY environment variable")
		return
	}

	for {
		// wait for user input
		query := getUserInputStr("Search for TV show (enter keywords or type 'q' to exit)")

		// quit program when "q" is entered, restart when nothing is entered
		if strings.Compare("q", query) == 0 {
			fmt.Println("Goodbye!")
			os.Exit(0)
		} else if strings.Compare("", query) == 0 {
			continue
		}

		var queryResponse TmdbQueryResponse
		httpRequest("https://api.themoviedb.org/3/search/tv?api_key="+api_key+"&query="+url.QueryEscape(query)+"&include_adult=false", &queryResponse)

		show_list := queryResponse.Results

		// display all shows found
		for show := range show_list {
			fmt.Println(show+1, ": ", show_list[show].Name)
		}

		if queryResponse.Total_pages > 1 {
			fmt.Println(" -- more than 20 entries found, please add more keywords --")
		}

		// let user select a show
		selectedShow, _ := getUserInputNum("Select TV show")

		if selectedShow == 0 {
			continue
		} else if selectedShow > 20 || selectedShow < 0 {
			fmt.Println("No valid selection")
			continue
		}

		show_id := show_list[selectedShow-1].Id

		// get all seasons
		var getDetailsResponse TmdbTvGetDetailsResponse
		httpRequest("https://api.themoviedb.org/3/tv/"+strconv.Itoa(show_id)+"?api_key="+api_key, &getDetailsResponse)

		num_of_seasons := getDetailsResponse.Number_of_seasons

		// let user select a season
		// read user input as number and convert to string afterwards, so we get an error when the user does not enter a number
		selectedSeason, _ := getUserInputNum("Select Season 1 to " + strconv.Itoa(num_of_seasons))
		if selectedSeason < 0 && selectedSeason > num_of_seasons {
			fmt.Println("No valid selection!")
			continue
			// TODO: go back to prompt
		}

		strSelectedSeason := strconv.Itoa(selectedSeason)

		// get episodes of the selected season
		var getSeasonDetailsResponse TmdbTvGetSeasonDetailsResponse
		httpRequest("https://api.themoviedb.org/3/tv/"+strconv.Itoa(show_id)+"/season/"+strSelectedSeason+"?api_key="+api_key, &getSeasonDetailsResponse)

		// display list of episodes
		fmt.Println("Episodes from Season ", strSelectedSeason, ":")

		for episode := range getSeasonDetailsResponse.Episodes {
			fmt.Println(getSeasonDetailsResponse.Episodes[episode].Episode_number, " - ", getSeasonDetailsResponse.Episodes[episode].Name)
		}

		// let user select episode
		selectedEpisode, _ := getUserInputNum("Select Episode 1 to " + strconv.Itoa(len(getSeasonDetailsResponse.Episodes)))

		if selectedEpisode < 1 || selectedEpisode > len(getSeasonDetailsResponse.Episodes) {
			fmt.Println("there's no such episode!")
			continue
		}

		// display name and overview of the selected episode
		fmt.Println()
		fmt.Println("Overview of Season ", strSelectedSeason, " Episode ", strconv.Itoa(selectedEpisode), ": ", getSeasonDetailsResponse.Episodes[selectedEpisode-1].Name)
		fmt.Println("   ", getSeasonDetailsResponse.Episodes[selectedEpisode-1].Overview)

	}
}
