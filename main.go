package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
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

	if strings.Compare("q", query) == 0 {
		fmt.Println("Goodbye!")
		os.Exit(0)
	}

	return query
}

func getUserInputNum(prompt string) (int, error) {
	query := getUserInputStr(prompt)

	num, err := strconv.Atoi(query)
	if err != nil {
		fmt.Println("No valid input")
		return 0, err
	}

	return num, nil
}

func httpRequest(uri string, sink interface{}) error {
	v := reflect.ValueOf(sink)
	if v.Kind() != reflect.Ptr || v.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("sink is not a struct pointer")
	}

	resp, err := http.Get(uri)
	if err != nil {
		fmt.Println("HTTP request failed: ", err)
		return err
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Could not read body: ", err)
		return err
	}

	err = json.Unmarshal(body, sink)
	if err != nil {
		fmt.Println("Could not parse body: ", err)
		return err
	}

	return nil
}

var api_key string
var show_list []TmdbQueryResponseResults
var num_of_seasons int
var strSelectedSeason string
var show_id int

const (
	stateUserQuery         int = 0
	stateUserSelectShow        = 1
	stateUserSelectSeason      = 2
	stateUserSelectEpisode     = 3
)

var currentState int
var nextState int

func userQuery() {
	// wait for user input
	query := getUserInputStr("Search for TV show (enter keywords or type 'q' to exit)")

	// restart when nothing is entered
	if strings.Compare("", query) == 0 {
		return
	}

	var queryResponse TmdbQueryResponse
	httpRequest("https://api.themoviedb.org/3/search/tv?api_key="+api_key+"&query="+url.QueryEscape(query)+"&include_adult=false", &queryResponse)
	show_list = queryResponse.Results

	if len(show_list) == 0 {
		fmt.Println("Didn't find anything. Please try searching with other keywords.")
		return
	}

	if queryResponse.Total_pages > 1 {
		fmt.Println(" -- more than 20 entries found, please add more keywords --")
	}

	nextState = stateUserSelectShow
}

func userSelectShow() {

	// display all shows found
	for show := range show_list {
		fmt.Println(show+1, ": ", show_list[show].Name)
	}

	// let user select a show
	selectedShow, err := getUserInputNum("Select TV show")
	if err != nil {
		fmt.Println("No valid input")
		return
	}

	if selectedShow > 20 || selectedShow < 1 {
		fmt.Println("No valid selection")
		return
	}

	show_id = show_list[selectedShow-1].Id

	// get all seasons
	var getDetailsResponse TmdbTvGetDetailsResponse
	httpRequest("https://api.themoviedb.org/3/tv/"+strconv.Itoa(show_id)+"?api_key="+api_key, &getDetailsResponse)

	num_of_seasons = getDetailsResponse.Number_of_seasons

	nextState = stateUserSelectSeason
}

func userSelectSeason() {
	// let user select a season
	// read user input as number and convert to string afterwards, so we get an error when the user does not enter a number
	selectedSeason, err := getUserInputNum("Select Season 1 to " + strconv.Itoa(num_of_seasons))

	if err != nil {
		fmt.Println("No valid input")
		return
	}

	if selectedSeason < 0 || selectedSeason > num_of_seasons {
		fmt.Println("No valid selection!")
		return
	}

	strSelectedSeason = strconv.Itoa(selectedSeason)

	nextState = stateUserSelectEpisode
}

func userSelectEpisode() {
	// get episodes of the selected season
	var getSeasonDetailsResponse TmdbTvGetSeasonDetailsResponse
	httpRequest("https://api.themoviedb.org/3/tv/"+strconv.Itoa(show_id)+"/season/"+strSelectedSeason+"?api_key="+api_key, &getSeasonDetailsResponse)

	// display list of episodes
	fmt.Println("Episodes from Season ", strSelectedSeason, ":")

	for episode := range getSeasonDetailsResponse.Episodes {
		fmt.Println(getSeasonDetailsResponse.Episodes[episode].Episode_number, " - ", getSeasonDetailsResponse.Episodes[episode].Name)
	}

	// let user select episode
	selectedEpisode, err := getUserInputNum("Select Episode 1 to " + strconv.Itoa(len(getSeasonDetailsResponse.Episodes)))

	if err != nil {
		fmt.Println("No valid selection")
	}

	if selectedEpisode < 1 || selectedEpisode > len(getSeasonDetailsResponse.Episodes) {
		fmt.Println("there's no such episode!")
		return
	}

	// display name and overview of the selected episode
	fmt.Println()
	fmt.Println("Overview of Season ", strSelectedSeason, " Episode ", strconv.Itoa(selectedEpisode), ": ", getSeasonDetailsResponse.Episodes[selectedEpisode-1].Name)
	fmt.Println("   ", getSeasonDetailsResponse.Episodes[selectedEpisode-1].Overview)

	nextState = stateUserQuery
}

func main() {
	// get TMDB API key from environment variable
	api_key = os.Getenv("TMDB_API_KEY")
	if api_key == "" {
		fmt.Println("Please set TMDB_API_KEY environment variable")
		return
	}

	currentState = stateUserQuery
	nextState = stateUserQuery

	for {
		switch currentState {
		case stateUserQuery:
			userQuery()
		case stateUserSelectShow:
			userSelectShow()
		case stateUserSelectSeason:
			userSelectSeason()
		case stateUserSelectEpisode:
			userSelectEpisode()
		}

		if nextState != currentState {
			currentState = nextState
		}
	}
}
