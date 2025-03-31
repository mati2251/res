package http

import (
	"encoding/json"
	"log"
	"net/http"
	"res/pkg/vm"
	"strconv"
	"strings"
	"time"
)

var acceptedContentType = "application/json"

func negotiateContentType(r *http.Request) bool {
	acceptedRaw := r.Header.Get("Accept")
	accepted := strings.SplitSeq(acceptedRaw, ",")
	for accept := range accepted {
		withPriority := strings.Split(accept, ";")
		if len(withPriority) < 1 {
			continue
		}
		acceptedTuple := strings.Split(withPriority[0], "/")
		if len(acceptedTuple) < 2 {
			continue
		}
		if acceptedTuple[0] == "*" {
			return true
		}
		if acceptedTuple[0] == "application" && (acceptedTuple[1] == "*" || acceptedTuple[1] == "json") {
			return true
		}
	}
	return false
}

type ResponseWriterMiddleware struct {
	http.ResponseWriter
	statusCode    int
	contentLength int
}

func (rwm *ResponseWriterMiddleware) WriteHeader(statusCode int) {
	rwm.statusCode = statusCode
	rwm.ResponseWriter.WriteHeader(statusCode)
}

func (rwm *ResponseWriterMiddleware) Write(b []byte) (int, error) {
	if rwm.statusCode == 0 {
		rwm.statusCode = http.StatusOK
	}
	if rwm.contentLength == 0 {
		rwm.contentLength += len(b)
	}
	return rwm.ResponseWriter.Write(b)
}

func CommonLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rwm := &ResponseWriterMiddleware{w, 0, 0}
		next.ServeHTTP(rwm, r)
		log.Printf("%s - - [%s] \"%s %s %s\" %d %d",
			r.RemoteAddr,
			start.Format(time.DateTime),
			r.Method,
			r.URL.Path,
			r.Proto,
			rwm.statusCode,
			rwm.contentLength,
		)
	})
}

func Job(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		GetJob(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TODO add checking if the sources is available
func PostJob(w http.ResponseWriter, r *http.Request) {
	if !negotiateContentType(r) {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	job := vm.NewJob("script.sh")

	err := job.CreateSpecFile()
	if err != nil {
		log.Printf("Error during creating spec file for job %d: %v", job.Id, err)
		w.WriteHeader(http.StatusInternalServerError)
	}

	go func() {
		err := job.Vm.Spawn()
		if err != nil {
			log.Printf("Error during spawning job %d: %v", job.Id, err)
			return
		}
		log.Printf("Job %d spawned", job.Id)
		defer func() {
			if err := job.Vm.Kill(); err != nil {
				log.Printf("Error during killing job %d: %v", job.Id, err)
				return
			}
		}()
		if err := job.ExecScript(); err != nil {
			log.Printf("Error during executing script on job %d: %v", job.Id, err)
			return
		}
		log.Printf("Job %d finished", job.Id)
	}()

	jobJson, err := json.Marshal(job)
	if err != nil {
		log.Printf("Error during marshalling job %d: %v", job.Id, err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(jobJson)
}

func GetJob(w http.ResponseWriter, r *http.Request) {
	if !negotiateContentType(r) {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	jobIdRaw := r.PathValue("id")
	if jobIdRaw == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	jobId, err := strconv.ParseInt(jobIdRaw, 10, 64)

	job, err := vm.JobFromSpecFile(int(jobId))
	if err != nil {
		log.Printf("Error during getting job %s: %v", jobIdRaw, err)
		w.WriteHeader(http.StatusNotFound)
		return
	}

	err = json.NewEncoder(w).Encode(job)
	if err != nil {
		log.Printf("Error during encoding job %s: %v", jobIdRaw, err)
	}
}
