package main

import (
    "fmt"
    "html/template"
    "net/http"
    "strings"

    "github.com/go-redis/redis"
)

var redisClient *redis.Client

func main() {
    redisClient = redis.NewClient(&redis.Options{
        Addr: "redis-16488.c284.us-east1-2.gce.redns.redis-cloud.com:16488",
		Password: "XRrxCA9ZeH9bBsnCKYDgnqRHY4yQQifQ",
    })

	// test Redis connection
    _, err := redisClient.Ping().Result()
    if err != nil {
        panic(err)
    }

    http.HandleFunc("/", searchHandler)
    fmt.Println("Server is running on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method == "GET" {
        // render the search form
        tmpl := template.Must(template.ParseFiles("search.html"))
        tmpl.Execute(w, nil)
    } else if r.Method == "POST" {
        query := r.FormValue("query")

		// search and render
        matches := searchRedis(query)
        tmpl := template.Must(template.ParseFiles("results.html"))
        tmpl.Execute(w, matches)
    }
}

func searchRedis(query string) []map[string]string {
    var matches []map[string]string

    keys, err := redisClient.Keys("*").Result()
    if err != nil {
        return matches
    }

	// pipeline to get all keys and vals
    pipe := redisClient.Pipeline()
    for _, key := range keys {
        pipe.Get(key)
    }
    values, err := pipe.Exec()
    if err != nil {
        return matches
    }

    for i, key := range keys {
        value, ok := values[i].(*redis.StringCmd).Val(), true
        if !ok {
            continue
        }

        // check if key or value contains the query
        if strings.Contains(strings.ToLower(key), strings.ToLower(query)) ||
            strings.Contains(strings.ToLower(value), strings.ToLower(query)) {
            match := map[string]string{
                "key":   key,
                "value": value,
            }
            matches = append(matches, match)
        }
    }

    return matches
}