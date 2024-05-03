package main

import (
    "fmt"
    "html/template"
    "net/http"
	"strings"
	"encoding/json"
	"os"
	"time"
	"sort"
    "github.com/go-redis/redis"
	"github.com/joho/godotenv"
)

var redisClient *redis.Client

func main() {
    err := godotenv.Load()
    if err != nil {
        panic(err)
    }

    redisAddr := os.Getenv("REDIS_ADDR")
    redisPassword := os.Getenv("REDIS_PASSWORD")

    redisClient = redis.NewClient(&redis.Options{
        Addr:     redisAddr,
        Password: redisPassword,
    })

	// test Redis connection
    _, err = redisClient.Ping().Result()
    if err != nil {
        panic(err)
    }

    http.HandleFunc("/", searchHandler)
    fmt.Println("Server is running on http://localhost:8080")
    http.ListenAndServe(":8080", nil)
}

func searchHandler(w http.ResponseWriter, r *http.Request) {
    query := r.URL.Query().Get("query")
    if query == "" {
        tmpl := template.Must(template.ParseFiles("search.html"))
        tmpl.Execute(w, nil)
        return
    }

	startTime := time.Now()
    matches := searchRedis(query)
    elapsedTime := time.Since(startTime).Seconds()

    response := map[string]interface{}{
        "count":   len(matches),
        "time":    fmt.Sprintf("%.2f", elapsedTime),
        "matches": matches,
    }

    jsonResponse, err := json.Marshal(response)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}

func searchRedis(query string) []map[string]string {
    var matches []map[string]string

    escapedQuery := strings.ReplaceAll(query, `"`, `\"`)

    exactQuery := fmt.Sprintf("(@code:%s* => {$weight: 10.0})", escapedQuery)
    partialQuery := fmt.Sprintf("(@code:*%s* => {$weight: 5.0}) | (@description:*%s* => {$weight: 2.0}) | (@name:*%s* => {$weight: 1.0})", escapedQuery, escapedQuery, escapedQuery)

    exactResults, err := redisClient.Do("FT.SEARCH", "idxCourses", exactQuery, "LIMIT", 0, 3000).Result()
    if err != nil {
        fmt.Printf("Exact match search error: %v\n", err)
    } else {
        matches = processResults(exactResults)
    }

    partialResults, err := redisClient.Do("FT.SEARCH", "idxCourses", partialQuery, "LIMIT", 0, 3000).Result()
    if err != nil {
        fmt.Printf("Partial match search error: %v\n", err)
    } else {
        partialMatches := processResults(partialResults)
        matches = append(matches, partialMatches...)
    }

    return matches
}

func processResults(results interface{}) []map[string]string {
    var matches []map[string]string

    resultSlice, ok := results.([]interface{})
    if !ok || len(resultSlice) < 2 {
        return matches
    }

    for i := 1; i < len(resultSlice); i += 2 {
        courseFields, ok := resultSlice[i+1].([]interface{})
        if !ok {
            fmt.Println("Error casting course fields to interface.")
            continue
        }

        courseMap := make(map[string]string)
        for j := 0; j < len(courseFields); j += 2 {
            fieldKey, okKey := courseFields[j].(string)
            fieldValue, okValue := courseFields[j+1].(string)
            if okKey && okValue {
                courseMap[fieldKey] = fieldValue
            }
        }

        code, ok1 := courseMap["code"]
        name, ok2 := courseMap["name"]
        description, ok3 := courseMap["description"]
        if ok1 && ok2 && ok3 {
            key := fmt.Sprintf("%s - %s", code, name)
            value := fmt.Sprintf("%s: ", description)
            matches = append(matches, map[string]string{"key": key, "value": value})
        } else {
            fmt.Println("Error retrieving course data fields.")
        }
    }

	// sort matches alphabetically by "key"
    sort.Slice(matches, func(i, j int) bool {
        return matches[i]["key"] < matches[j]["key"]
    })

    return matches
}