package main

import (
    "encoding/json"
    "log"
    "os"
    "time"
    "net/http"
    "github.com/patrickmn/go-cache"
    // Shortening the import reference name seems to make it a bit easier
    owm "github.com/briandowns/openweathermap"
    "github.com/gin-gonic/gin"
)

var apiKey = os.Getenv("OWM_API_KEY") 
// Create a cache with a default expiration time of 15 minutes
var c = cache.New(15*time.Minute, 30*time.Minute)

func main() {}

// This function's name is a must. App Engine uses it to drive the requests properly.
func init() {
    // Starts a new Gin instance with no middle-ware
    r := gin.New()

    // Define your handlers
    r.GET("/", func(c *gin.Context) {
        c.String(http.StatusOK, "Hello World!")
    })
    r.GET("/ping", func(c *gin.Context) {
        c.String(http.StatusOK, "pong")
    })
    r.POST("/ping", func(c *gin.Context) {
        c.JSON(200, gin.H{
            "message": "pong",
        })
    })
    // Parameters in path
    // This handler will match /user/john but will not match /user/ or /user
    r.GET("/user/:name", func(c *gin.Context) {
        name := c.Param("name")
        c.String(http.StatusOK, "Hello %s", name)
    })
    // Parameters in querystring
    // Query string parameters are parsed using the existing underlying request object.
    // The request responds to a url matching:  /welcome?firstname=Jane&lastname=Doe
    r.GET("/weather", func(c *gin.Context) {
        lang := c.DefaultQuery("lang", "EN")
        city := c.Query("city") // shortcut for c.Request.URL.Query().Get("lastname")
        weather, err := weather(city, lang)
        if err != nil {
            log.Println(err.Error())
            c.String(http.StatusOK, err.Error())
        }
        // c.JSON(200, weather) // weather here should be struct not JSON 
        // (possible but requires changes in weather() code )
        c.String(http.StatusOK, weather) // This works but isn't kosher
    })
    r.POST("/weather", func(c *gin.Context) {
        // post form not querystring 
        lang := c.PostForm("lang")
        city := c.DefaultPostForm("city", "Cortona")
        weather, err := weather(city, lang)
        if err != nil {
            log.Println(err.Error())
            c.String(http.StatusOK, err.Error())
        }
        // c.JSON(200, weather)
        c.String(http.StatusOK, weather) // This works but isn't kosher
    })
    // By default it serves on :8080 unless a
    // PORT environment variable was defined.
    // r.Run(":3080")
    r.Run() // listen and serve on 0.0.0.0:8080
    // router.Run(":3000") for a hard coded port
    // Handle all requests using net/http
    http.Handle("/", r)
}

func weather(city string, lang string) (string, error) {
    if len(lang) < 1  {
        lang = "EN"
    }
    w, owmErr := owm.NewCurrent("C", lang, apiKey)
    if len(city) < 1  {
        city = "Cortona"
    }
    b, foundBody := c.Get(city)
    if foundBody {
        //TODO - add option to force refresh
        log.Println("Found weather", city)
    } else {
        log.Println("Not found weather", city)
        w.CurrentByName(city)
        if owmErr != nil {
            log.Println(owmErr)
        }
        out, jsonErr := json.MarshalIndent(w, "  ", "    ")
        if jsonErr != nil {
            log.Println(jsonErr)
        }
        // With some odd errors it is important to cache only when errors are nil
        if owmErr == nil && jsonErr ==nil {
            body := string(out)
            b = &body
            c.Set(city, &body, cache.DefaultExpiration)
            log.Println("Cached weather", city)
        }
    }
    body := *b.(*string) // It's little fucked up that Golang shit
    return string(body), nil
}


/* openWeather
https://mholt.github.io/json-to-go/ */
type openWeather struct {
    Coord struct {
        Lon float64 `json:"lon"`
        Lat float64 `json:"lat"`
    } `json:"coord"`
    Sys struct {
        Type    int     `json:"type"`
        ID      int     `json:"id"`
        Message float64 `json:"message"`
        Country string  `json:"country"`
        Sunrise int     `json:"sunrise"`
        Sunset  int     `json:"sunset"`
    } `json:"sys"`
    Base    string `json:"base"`
    Weather []struct {
        ID          int    `json:"id"`
        Main        string `json:"main"`
        Description string `json:"description"`
        Icon        string `json:"icon"`
    } `json:"weather"`
    Main struct {
        Temp      int `json:"temp"`
        TempMin   int `json:"temp_min"`
        TempMax   int `json:"temp_max"`
        Pressure  int `json:"pressure"`
        SeaLevel  int `json:"sea_level"`
        GrndLevel int `json:"grnd_level"`
        Humidity  int `json:"humidity"`
    } `json:"main"`
    Wind struct {
        Speed float64 `json:"speed"`
        Deg   int     `json:"deg"`
    } `json:"wind"`
    Clouds struct {
        All int `json:"all"`
    } `json:"clouds"`
    Rain struct {
        ThreeH int `json:"3h"`
    } `json:"rain"`
    Snow struct {
        ThreeH int `json:"3h"`
    } `json:"snow"`
    Dt   int    `json:"dt"`
    ID   int    `json:"id"`
    Name string `json:"name"`
    Cod  int    `json:"cod"`
    Unit string `json:"Unit"`
    Lang string `json:"Lang"`
    Key  string `json:"Key"`
}
