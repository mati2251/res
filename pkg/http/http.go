package http

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"res/pkg/config"
	"res/pkg/db"
	"res/pkg/virtual"
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

type Service struct {
	Queries *db.Queries
	Config  *config.Config
	js      virtual.JobService
}

func NewService(queries *db.Queries, config *config.Config) *Service {
	return &Service{
		Queries: queries,
		Config:  config,
		js:      virtual.JobService{Config: config, Queries: queries},
	}
}

func (s Service) Job(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getJob(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

// TODO add checking if the sources is available
func (s Service) PostJob(w http.ResponseWriter, r *http.Request) {
	if !negotiateContentType(r) {
		w.WriteHeader(http.StatusNotAcceptable)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	ctx := context.Background()
	vm, err := s.Queries.InsertVm(ctx, db.InsertVmParams{
		Name:   "test",
		Memory: 1024,
		Cpu:    1,
		Disk:   1024,
		Image:  "debian",
		Port:   8080,
	})
	if err != nil {
		log.Printf("Error during inserting VM: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	job, err := s.Queries.InsertJob(ctx, db.InsertJobParams{
		VmID:     vm.ID,
		Status:   db.VmStatusRunning,
		BasePath: s.Config.BaseDir,
	})
	if err != nil {
		log.Printf("Error during inserting job: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = os.MkdirAll(job.BasePath, 0755)
	if err != nil {
		log.Printf("Error during creating job directory %s: %v", job.BasePath, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = virtual.CreateSpecFile(&job)
	if err != nil {
		log.Printf("Error during creating spec file for job %d: %v", job.ID, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	go func() {
		cmd, err := virtual.Spawn(&vm)
		if err != nil {
			log.Printf("Error during spawning job %d: %v", job.ID, err)
			return
		}
		log.Printf("Job %d spawned", job.ID)
		defer func() {
			log.Printf("Job %d killed %d", job.ID, cmd.Process.Pid)
			if err := cmd.Process.Signal(os.Kill); err != nil {
				log.Printf("Error during killing job %d: %v", job.ID, err)
			}
		}()
		if err := virtual.ExecScript(&job, &vm); err != nil {
			log.Printf("Error during executing script on job %d: %v", job.ID, err)
			return
		}
		log.Printf("Job %d finished", job.ID)
	}()

	jobJson, err := json.Marshal(job)
	if err != nil {
		log.Printf("Error during marshalling job %d: %v", job.ID, err)
	}
	w.WriteHeader(http.StatusCreated)
	w.Write(jobJson)
}

func (s Service) getJob(w http.ResponseWriter, r *http.Request) {
	// if !negotiateContentType(r) {
	// 	w.WriteHeader(http.StatusNotAcceptable)
	// 	return
	// }
	//
	// jobIdRaw := r.PathValue("id")
	// if jobIdRaw == "" {
	// 	w.WriteHeader(http.StatusBadRequest)
	// 	return
	// }
	//
	// jobId, err := strconv.ParseInt(jobIdRaw, 10, 64)
	//
	// job, err := s.js.JobFromSpecFile(int(jobId))
	// if err != nil {
	// 	log.Printf("Error during getting job %s: %v", jobIdRaw, err)
	// 	w.WriteHeader(http.StatusNotFound)
	// 	return
	// }
	//
	// err = json.NewEncoder(w).Encode(job)
	// if err != nil {
	// 	log.Printf("Error during encoding job %s: %v", jobIdRaw, err)
	// }
}
