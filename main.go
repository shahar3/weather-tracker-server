package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/xuri/excelize/v2"
	"strconv"
	"strings"
)

func main() {
	locationsMap := loadData()

	//build autocomplete array in the format of "city, country"
	options := buildAutoCompleteOptionsList(locationsMap)

	fmt.Println("Running server")
	r := gin.Default()

	r.GET("/health", func(context *gin.Context) {
		context.JSON(200, map[string]string{"message": "ok"})
	})

	apiGroup := r.Group("api")
	locationsApi := apiGroup.Group("locations")
	usersApi := apiGroup.Group("users")

	locationsApi.GET("/auto", func(context *gin.Context) {
		term := context.Query("term")

		//filter options
		filteredOptions := filterOptions(options, term)

		//return a list of all the filtered options location including the latitude and longitude
		filteredLocations := buildFilteredLocationsList(filteredOptions, locationsMap)
		context.JSON(200, map[string]interface{}{
			"options": filteredLocations,
		})
	})

	usersApi.GET("/fetch-user", func(context *gin.Context) {
		context.String(200, "fetching user")
	})
	_ = r.Run(":3001")
}

func buildFilteredLocationsList(options []string, locationsMap map[string]Location) []Location {
	var filteredLocations []Location
	for _, option := range options {
		cityName := option[:strings.Index(option, ",")]
		for key, location := range locationsMap {
			if strings.ToLower(key) == strings.ToLower(cityName) {
				filteredLocations = append(filteredLocations, location)
			}
		}
	}

	return filteredLocations
}

func filterOptions(options []string, term string) []string {
	var filteredOptions []string

	term = strings.ToLower(term)
	for _, option := range options {
		option = strings.ToLower(option)
		if strings.HasPrefix(option, term) {
			filteredOptions = append(filteredOptions, option)
		}
	}

	return filteredOptions
}

func buildAutoCompleteOptionsList(locationsMap map[string]Location) []string {
	var options []string
	for _, location := range locationsMap {
		autoCompleteOption := fmt.Sprintf("%s, %s", location.City, location.Country)
		options = append(options, autoCompleteOption)
	}

	return options
}

type Location struct {
	City      string
	Country   string
	Longitude float64
	Latitude  float64
}

func loadData() map[string]Location {
	fmt.Println("Reading data file")
	locationsMap := make(map[string]Location, 0)

	file, err := excelize.OpenFile("data/worldcities.xlsx")
	if err != nil {
		fmt.Println(err)
	}

	defer func() {
		if err = file.Close(); err != nil {
			fmt.Println(err)
		}
	}()

	rows, err := file.Rows("Sheet1")
	if err != nil {
		fmt.Println(err)
	}

	//skip the first line
	rows.Next()
	for rows.Next() {
		row, err := rows.Columns()
		if err != nil {
			fmt.Println(err)
		}

		city := row[1]
		lat, err := strconv.ParseFloat(row[2], 64)
		if err != nil {
			fmt.Println(err)
		}
		lng, err := strconv.ParseFloat(row[3], 64)
		if err != nil {
			fmt.Println(err)
		}
		country := row[4]

		locationsMap[city] = Location{
			City:      city,
			Country:   country,
			Longitude: lng,
			Latitude:  lat,
		}
	}

	return locationsMap
}
