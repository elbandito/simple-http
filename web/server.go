package web

import (
    "net/http"
    "context"
    "time"
    "strings"
    "strconv"
    "encoding/json"
    "encoding/base64"
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
            w.Write([]byte(s.passwordStorage[jobId]))

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

            s.hashTimes[s.jobCounter] += time.Since(startTime).Seconds()

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

func (s *Server) Stop() {
    ctx, cancel := context.WithTimeout(context.Background(), 10 * time.Second)
    defer cancel()

    s.httpServer.Shutdown(ctx)
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
        Unit: "seconds",
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
    time.Sleep(5 * time.Second)

    startTime := time.Now()
    encoded := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(password)))
    s.passwordStorage[jobId] = encoded

    s.hashTimes[jobId] += time.Since(startTime).Seconds()
}