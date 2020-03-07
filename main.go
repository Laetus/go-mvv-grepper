package main

import (
    "bytes"
    "encoding/json"
    "fmt"
    "io/ioutil"
    "net/http"
    "strings"
    "sync"
    "time"

    "github.com/bradfitz/slice"
)

var wg sync.WaitGroup

type Product string

const (
    BUS   Product = "BUS"
    TRAM  Product = "TRAM"
    SBAHN Product = "SBAHN"
    UBAHN Product = "UBAHN"
)

type Lines struct {
    tram       [100]string
    nachttram  [100]string
    sbahn      [100]string
    ubahn      [100]string
    bus        [100]string
    nachtbus   [100]string
    otherlines [100]string
}

type Location struct {
    Typ         string    `json:"type"`
    Latitude    float64   `json:"latitude"`
    Longitude   float64   `json:"longitude"`
    Id          int32     `json:"id"`
    Place       string    `json:"place"`
    Name        string    `json:"name"`
    HasLiveData bool      `json:"hasLiveData"`
    HasZoomData bool      `json:"hasZoomData"`
    Products    []Product `json:"products"`
    Distance    int32     `json:"distance"`
    Lines       Lines     `json:"lines"`
}

type Departure struct {
    DepartureTime       int64   `json:"departureTime"`
    Product             Product `json:"product"`
    Label               string  `json:"label"`
    Destination         string  `json:"destination"`
    Live                bool    `json:"live"`
    LineBackgroundColor string  `json:"lineBackgroundColor"`
    DepartureId         int64   `json:"departureId"`
    Sev                 bool    `json:"sev"`
}

type Line struct {
    Destination string  `json:"destination"`
    Sev         bool    `json:"sev"`
    PartialNet  string  `json:"partialNet"`
    Product     Product `json:"product"`
    LineNumber  string  `json:lineNumber`
    DiveId      string  `json:divaId`
}

type LocationResponse struct {
    Locations []Location `json:"locations"`
}

type DepartureResponse struct {
    ServingLines []Line      `json:servingLines`
    Departures   []Departure `json:departures`
}

type GeoLocation struct {
    latitude  float64
    longitude float64
}

func get_locations(geo_location *GeoLocation) []Location {
    var url_buffer bytes.Buffer
    var url string
    url_buffer.WriteString("https://www.mvg.de/fahrinfo/api/location/nearby?latitude=")
    url_buffer.WriteString(fmt.Sprintf("%f", (*geo_location).latitude))
    url_buffer.WriteString("&longitude=")
    url_buffer.WriteString(fmt.Sprintf("%f", (*geo_location).longitude))

    url = url_buffer.String()
    req, _ := http.NewRequest("GET", url, nil)

    req.Header.Add("Cache-Control", "no-cache")
    req.Header.Add("x-mvg-authorization-key", "5af1beca494712ed38d313714d4caff6")

    res, _ := http.DefaultClient.Do(req)

    defer res.Body.Close()

    body, _ := ioutil.ReadAll(res.Body)
    var foo LocationResponse
    err2 := json.Unmarshal([]byte(body), &foo)
    if err2 != nil {
        fmt.Println("ERROR2: \n", err2)
    }

    return foo.Locations
}

func get_departures(c chan map[*Location][]Departure, location *Location) {
    defer wg.Done()
    var url_buffer bytes.Buffer
    var url string

    url_buffer.WriteString("https://www.mvg.de/fahrinfo/api/departure/")
    url_buffer.WriteString(fmt.Sprintf("%d", (*location).Id))

    url = url_buffer.String()
    req, _ := http.NewRequest("GET", url, nil)

    req.Header.Add("Cache-Control", "no-cache")
    req.Header.Add("x-mvg-authorization-key", "5af1beca494712ed38d313714d4caff6")
    res, _ := http.DefaultClient.Do(req)

    defer res.Body.Close()

    body, _ := ioutil.ReadAll(res.Body)
    var tmp DepartureResponse
    err := json.Unmarshal([]byte(body), &tmp)
    if err != nil {
        fmt.Println("ERROR: \n", err)
    }

    fmt.Println("Loaded  Location: ", (*location).Name)
    ret_map := make(map[*Location][]Departure)
    ret_map[location] = tmp.Departures
    c <- ret_map
    return
}

func get_all_departures(locations *([]Location)) map[*Location][]Departure {
    channel := make(chan map[*Location][]Departure, len(*locations))
    for _, location := range *locations {
        wg.Add(1)
        fmt.Println("Loading Location: ", location.Name)
        loc := location
        go get_departures(channel, &loc)
    }
    fmt.Println("Waiting last time")
    wg.Wait()
    fmt.Println("Done last time")

    close(channel)
    fmt.Println("Channel closed")

    ret_map := make(map[*Location][]Departure)
    for elem := range channel {
        for k, v := range elem {
            ret_map[k] = v
        }
    }
    return ret_map
}

func pretty_print(loc2deps *map[*Location][]Departure) {
    now := int64(time.Now().Unix()) * 1000

    keys := make([]*Location, 0, len(*loc2deps))
    for key := range *loc2deps {
        keys = append(keys, key)
    }

    // Sort locations
    slice.Sort(keys[:], func(i, j int) bool {
        return (*keys[i]).Distance < (*keys[j]).Distance
    })

    for _, location := range keys {
        deps := (*loc2deps)[location]
        fmt.Println("\n\nHaltestelle: ",
            location.Name, "\tDistance: ",
            fmt.Sprintf("%d", location.Distance))
        fmt.Println("--------------------------------------------------------")
        for index, elem := range deps {
            if index >= 5 {
                break
            }
            destination := elem.Destination
            max_length := 15
            destination = strings.Replace(destination, "ß ", "ß", -1)
            if len(destination) < max_length {
                destination += strings.Repeat(" ", max_length-len(destination))
            }
            if len(destination) > max_length {
                destination = destination[:max_length-3] + "..."
            }
            fmt.Println("Richtung: ", destination, "  Abfahrt: ",
                fmt.Sprintf("%d", (elem.DepartureTime-now)/1000/60), "\t",
                elem.Product, "\t", elem.Label)
        }
    }
}

func main() {
    var geo_loc GeoLocation
    // get locations nearby
    geo_loc = GeoLocation{latitude: 48.157221, longitude: 11.511238}
    locations := get_locations(&geo_loc)

    res := get_all_departures(&locations)
    pretty_print(&res)
    //unixTime := time.Unix(departure.DepartureTime/1000, 0)
    //fmt.Println(unixTime.Format(time.RFC3339))

}
