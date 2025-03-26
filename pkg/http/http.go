package http

import (
	"encoding/json"
	"log"
	"net/http"
	"res/pkg/vm"
	"time"
)

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
			start.Format("02/Jan/2006:15:04:05 -0700"),
			r.Method,
			r.URL.Path,
			r.Proto,
			rwm.statusCode,
			rwm.contentLength,
		)
	})
}

// TODO add checking if the sources is available
func PostJob(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	job := vm.NewJob("script.sh")
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
}

func CancleJob(w http.ResponseWriter, r *http.Request) {
}

func GetJobLogs(w http.ResponseWriter, r *http.Request) {
}

func GetJobs(w http.ResponseWriter, r *http.Request) {
}

func GetArtifact(w http.ResponseWriter, r *http.Request) {
}

func PostArtifact(w http.ResponseWriter, r *http.Request) {
}

func DeleteArtifact(w http.ResponseWriter, r *http.Request) {
}

func PutArtifact(w http.ResponseWriter, r *http.Request) {
}

func GetArtifacts(w http.ResponseWriter, r *http.Request) {
}

func PostVM(w http.ResponseWriter, r *http.Request) {
}

func GetVM(w http.ResponseWriter, r *http.Request) {
}

func DeleteVM(w http.ResponseWriter, r *http.Request) {
}

func GetImage(w http.ResponseWriter, r *http.Request) {
}

func PostImage(w http.ResponseWriter, r *http.Request) {
}

func DeleteImage(w http.ResponseWriter, r *http.Request) {
}

func GetImages(w http.ResponseWriter, r *http.Request) {
}
