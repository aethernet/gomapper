package main

import (
	"fmt"
	"os"

	"github.com/antonholmquist/jason"
)
type Fixture struct {
		start []int64
    end []int64
    shape string
    pixelCount int64
    pixelAddressStartsAt int64
    universe int64
}

func decodeFixtureFromJson(file string) ([]Fixture, error) {
	var fixtures []Fixture

		bytes, err := os.ReadFile(file)

    if err != nil {
        fmt.Println("Unable to load config file!")
        return fixtures, err
    }
		
		data, err := jason.NewObjectFromBytes(bytes)

    if err != nil {
        fmt.Println("JSON decode error!")
        return fixtures, err
    }

		decodedfixtures, _ := data.GetObjectArray("fixtures")
		for _, fixture := range decodedfixtures {
			shape, _ := fixture.GetString("shape")
			start, _ := fixture.GetInt64Array("start")
			end, _ := fixture.GetInt64Array("end")
			pixelCount, _ := fixture.GetInt64("pixelCount")
			pixelAddressStartsAt, _ := fixture.GetInt64("pixelAddressStartsAt")
			universe, _ := fixture.GetInt64("universe")

			fix := Fixture{start, end, shape, pixelCount, pixelAddressStartsAt, universe}

			fixtures = append(fixtures, fix)
		}

		return fixtures, nil
}