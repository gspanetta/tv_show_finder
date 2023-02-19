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
	"io"
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

type Context struct {
	input io.Reader
	api_key string
	show_list []TmdbQueryResponseResults
	num_of_seasons int
	strSelectedSeason string
	show_id int
	currentState int
	nextState int
}

const (
	stateUserQuery         int = 0
	stateUserSelectShow        = 1
	stateUserSelectSeason      = 2
	stateUserSelectEpisode     = 3
)


func getUserInputStr(ctx *Context, prompt string) string {
	reader := bufio.NewReader(ctx.input)
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

func getUserInputNum(ctx *Context, prompt string) (int, error) {
	query := getUserInputStr(ctx, prompt)

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

func userQuery(ctx *Context) {
	// wait for user input
	query := getUserInputStr(ctx, "Search for TV show (enter keywords or type 'q' to exit)")

	// restart when nothing is entered
	if strings.Compare("", query) == 0 {
		return
	}

	var queryResponse TmdbQueryResponse
	httpRequest("https://api.themoviedb.org/3/search/tv?api_key="+ctx.api_key+"&query="+url.QueryEscape(query)+"&include_adult=false", &queryResponse)
	ctx.show_list = queryResponse.Results

	if len(ctx.show_list) == 0 {
		fmt.Println("Didn't find anything. Please try searching with other keywords.")
		return
	}

	if queryResponse.Total_pages > 1 {
		fmt.Println(" -- more than 20 entries found, please add more keywords --")
	}

	ctx.nextState = stateUserSelectShow
}

func userSelectShow(ctx *Context) {

	// display all shows found
	for show := range ctx.show_list {
		fmt.Println(show+1, ": ", ctx.show_list[show].Name)
	}

	// let user select a show
	selectedShow, err := getUserInputNum(ctx, "Select TV show")
	if err != nil {
		fmt.Println("No valid input")
		return
	}

	if selectedShow > 20 || selectedShow < 1 {
		fmt.Println("No valid selection")
		return
	}

	ctx.show_id = ctx.show_list[selectedShow-1].Id

	// get all seasons
	var getDetailsResponse TmdbTvGetDetailsResponse
	httpRequest("https://api.themoviedb.org/3/tv/"+strconv.Itoa(ctx.show_id)+"?api_key="+ctx.api_key, &getDetailsResponse)

	ctx.num_of_seasons = getDetailsResponse.Number_of_seasons

	ctx.nextState = stateUserSelectSeason
}

func userSelectSeason(ctx *Context) {
	// let user select a season
	// read user input as number and convert to string afterwards, so we get an error when the user does not enter a number
	selectedSeason, err := getUserInputNum(ctx, "Select Season 1 to " + strconv.Itoa(ctx.num_of_seasons))

	if err != nil {
		fmt.Println("No valid input")
		return
	}

	if selectedSeason < 0 || selectedSeason > ctx.num_of_seasons {
		fmt.Println("No valid selection!")
		return
	}

	ctx.strSelectedSeason = strconv.Itoa(selectedSeason)

	ctx.nextState = stateUserSelectEpisode
}

func userSelectEpisode(ctx *Context) {
	// get episodes of the selected season
	var getSeasonDetailsResponse TmdbTvGetSeasonDetailsResponse
	httpRequest("https://api.themoviedb.org/3/tv/"+strconv.Itoa(ctx.show_id)+"/season/"+ctx.strSelectedSeason+"?api_key="+ctx.api_key, &getSeasonDetailsResponse)

	// display list of episodes
	fmt.Println("Episodes from Season ", ctx.strSelectedSeason, ":")

	for episode := range getSeasonDetailsResponse.Episodes {
		fmt.Println(getSeasonDetailsResponse.Episodes[episode].Episode_number, " - ", getSeasonDetailsResponse.Episodes[episode].Name)
	}

	// let user select episode
	selectedEpisode, err := getUserInputNum(ctx, "Select Episode 1 to " + strconv.Itoa(len(getSeasonDetailsResponse.Episodes)))

	if err != nil {
		fmt.Println("No valid selection")
	}

	if selectedEpisode < 1 || selectedEpisode > len(getSeasonDetailsResponse.Episodes) {
		fmt.Println("there's no such episode!")
		return
	}

	// display name and overview of the selected episode
	fmt.Println()
	fmt.Println("Overview of Season ", ctx.strSelectedSeason, " Episode ", strconv.Itoa(selectedEpisode), ": ", getSeasonDetailsResponse.Episodes[selectedEpisode-1].Name)
	fmt.Println("   ", getSeasonDetailsResponse.Episodes[selectedEpisode-1].Overview)

	ctx.nextState = stateUserQuery
}

func main() {
	var context Context

	// get TMDB API key from environment variable
	context.api_key = os.Getenv("TMDB_API_KEY")
	if context.api_key == "" {
		fmt.Println("Please set TMDB_API_KEY environment variable")
		return
	}
	context.input = os.Stdin
	context.currentState = stateUserQuery
	context.nextState = stateUserQuery

	for {
		switch context.currentState {
		case stateUserQuery:
			userQuery(&context)
		case stateUserSelectShow:
			userSelectShow(&context)
		case stateUserSelectSeason:
			userSelectSeason(&context)
		case stateUserSelectEpisode:
			userSelectEpisode(&context)
		}

		if context.nextState != context.currentState {
			context.currentState = context.nextState
		}
	}
}
