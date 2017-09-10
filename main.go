package main

import (
    "fmt"
    "net/http"
    "log"
    "encoding/base64"
    "strings"
    "strconv"
    "time"
    "encoding/json"
)

type hashStats struct {
    Total   int     `json:"total"`
    Average float64 `json:"average"`
    Unit    string  `json:"unit"`
}

var passwordStorage = make(map[int]string)
var hashTimes = make(map[int]float64)

func main() {

    var jobCounter = 0

    http.HandleFunc("/hash/", func(w http.ResponseWriter, r *http.Request) {

        switch r.Method {

        case "GET":
            urlParts := strings.Split(r.URL.Path, "/")
            jobId, err := strconv.Atoi(urlParts[len(urlParts) - 1])
            if err != nil {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            w.WriteHeader(http.StatusOK)
            w.Write([]byte(passwordStorage[jobId]))

        default:
            w.WriteHeader(http.StatusNotFound)
        }
    })

    http.HandleFunc("/hash", func(w http.ResponseWriter, r *http.Request) {

        switch r.Method {
        case "POST":
            startTime := time.Now()
            r.ParseForm()

            password := r.Form.Get("password")
            if password == "" {
                w.WriteHeader(http.StatusBadRequest)
                return
            }

            jobCounter += 1
            go storePassword(jobCounter, password)

            w.WriteHeader(http.StatusAccepted)
            w.Write([]byte(strconv.Itoa(jobCounter)))

            elapsed := time.Since(startTime)
            hashTimes[jobCounter] += elapsed.Seconds()

        default:
            w.WriteHeader(http.StatusNotFound)
        }
    })

    http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
        switch r.Method {
        case "GET":
            stats := getLatestStats()

            b, err := json.Marshal(&stats)
            if err != nil {
                w.WriteHeader(http.StatusInternalServerError)
            }

            w.WriteHeader(http.StatusOK)
            w.Write(b)

        default:
            w.WriteHeader(http.StatusNotFound)
        }
    })

    log.Fatal(http.ListenAndServe(":8080", nil))
}

func getLatestStats() hashStats {
    count := 0
    sum := 0.0

    stats := hashStats{
        Total:   0,
        Average: 0.0,
        Unit: "seconds",
    }

    for _, v := range hashTimes {
        count += 1
        sum += v
    }

    if count != 0 {
        stats.Total = count
        stats.Average = sum / float64(count)
    }

    return stats
}

func storePassword(jobId int, password string) {
    time.Sleep(5 * time.Second)

    startTime := time.Now()
    encoded := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(password)))
    passwordStorage[jobId] = encoded

    elapsed := time.Since(startTime)
    hashTimes[jobId] += elapsed.Seconds()
}
