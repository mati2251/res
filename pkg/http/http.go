package http

import (
	"log"
	"net/http"
	"res/pkg/vm"
	"time"
)

type ResponseWriterMiddleware struct {
  http.ResponseWriter
  statusCode int
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

	job := vm.NewJob("script.sh")
	go func() {
		err := vm.Spawn(job.Vm)
		if err != nil {
      log.Printf("Error during spawning job %d: %v", job.Id, err)
			return
		}
		if err := vm.ExecScript(job.Vm, "script.sh"); err != nil {
      log.Printf("Error during executing script on job %d: %v", job.Id, err)
			return
		}
		if err := vm.Kill(job.Vm); err != nil {
      log.Printf("Error during killing job %d: %v", job.Id, err)
			return
		}
	}()
	w.WriteHeader(http.StatusCreated)
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
