package web

import (
    "net/http"
    "context"
    "time"
    "strings"
    "strconv"
    "encoding/json"
    "encoding/base64"
    "fmt"
    "log"
    "crypto/sha512"
)

type hashStats struct {
    Total   int     `json:"total"`
    Average float64 `json:"average"`
    Unit    string  `json:"unit"`
}

type Server struct {
    httpServer *http.Server
    passwordStorage map[int]string
    hashTimes map[int]float64
    jobCounter int
}

func NewServer() *Server {
    return &Server{
        httpServer: &http.Server{Addr: ":8080", Handler: nil},
        passwordStorage: make(map[int]string),
        hashTimes: make(map[int]float64),
        jobCounter: 0,
    }
}

func (s *Server) Start() {

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
            base64Encoded := base64.StdEncoding.EncodeToString([]byte(s.passwordStorage[jobId]))

            w.Write([]byte(base64Encoded))

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

            s.jobCounter += 1
            go s.storePassword(s.jobCounter, password)

            w.WriteHeader(http.StatusAccepted)
            w.Write([]byte(strconv.Itoa(s.jobCounter)))

            s.hashTimes[s.jobCounter] += time.Since(startTime).Seconds() * 1000

        default:
            w.WriteHeader(http.StatusNotFound)
        }
    })

    http.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {

        switch r.Method {
        case "GET":
            stats := s.getLatestStats()

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

    s.httpServer.ListenAndServe()
}

func (s *Server) Stop(ctx context.Context) {
    if err := s.httpServer.Shutdown(ctx); err != nil {
        log.Println(err)
    }

    // Wait for outstanding requests to finish.  Had to add this separate timer
    // because we have potential background jobs that could still be running.
    // These background jobs are NOT tied to any open requests thus the server
    // has no knowledge of them and will not wait for them during Shutdown().
    time.Sleep(10 * time.Second)
}

/************
/* Private **
************/

func (s *Server) getLatestStats() hashStats {
    count := 0
    sum := 0.0

    stats := hashStats{
        Total:   0,
        Average: 0.0,
        Unit: "milliseconds",
    }

    for _, v := range s.hashTimes {
        count += 1
        sum += v
    }

    if count != 0 {
        stats.Total = count
        stats.Average = sum / float64(count)
    }

    return stats
}

func (s *Server) storePassword(jobId int, password string) {
    fmt.Printf("start processing job %d...\n", jobId)
    time.Sleep(5 * time.Second)

    // Include only the processing time excluding the 5 second wait period.
    startTime := time.Now()
    sha512Encoded := sha512.Sum512([]byte(password))
    s.passwordStorage[jobId] = string(sha512Encoded[:])

    s.hashTimes[jobId] += time.Since(startTime).Seconds() * 1000
    fmt.Printf("finished processing job %d\n", jobId)
}