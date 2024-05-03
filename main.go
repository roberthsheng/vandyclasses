package main

import (
    "fmt"
    "html/template"
    "net/http"
	"strings"
	"encoding/json"
    "github.com/go-redis/redis"
	"os"
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

    matches := searchRedis(query)
	if len(matches) == 0 { // Check if no matches were found and handle accordingly
        jsonResponse, _ := json.Marshal([]map[string]string{})
        w.Header().Set("Content-Type", "application/json")
        w.Write(jsonResponse)
        return
    }
	
    jsonResponse, err := json.Marshal(matches)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    w.Write(jsonResponse)
}

func searchRedis(query string) []map[string]string {
    var matches []map[string]string

    // Properly format the query for wildcard searching without quotes around the wildcard part
    escapedQuery := strings.ReplaceAll(query, `"`, `\"`)
    formattedQuery := fmt.Sprintf("@name:%s* | @description:%s*", escapedQuery, escapedQuery)

    // Debugging: Print the formatted query to see what's being sent to Redis
    fmt.Printf("Formatted query: %s\n", formattedQuery)

    // Execute the search command using RediSearch
    results, err := redisClient.Do("FT.SEARCH", "idxCourses", formattedQuery, "LIMIT", 0, 10).Result()
    if err != nil {
        fmt.Printf("Search error: %v\n", err)
        return matches
    }

    // Debugging: Print raw results to understand what Redis is returning
    // fmt.Printf("Raw results: %#v\n", results)

    // Process the results
    resultSlice, ok := results.([]interface{})
    if !ok || len(resultSlice) < 2 { // The first element is the count of the results
        fmt.Println("No results found or parsing error.")
        return matches
    }

    // Iterate over the results, skipping the first element (the count)
    for i := 1; i < len(resultSlice); i += 2 {
        // courseKey, _ := resultSlice[i].(string) // The course key
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

        // Check if the necessary fields are present
        code, ok1 := courseMap["code"]
        name, ok2 := courseMap["name"]
        description, ok3 := courseMap["description"]
        if ok1 && ok2 && ok3 {
            key := fmt.Sprintf("course:%s", code)
            value := fmt.Sprintf("%s: %s", name, description)
            matches = append(matches, map[string]string{"key": key, "value": value})
        } else {
            fmt.Println("Error retrieving course data fields.")
        }
    }

    return matches
}