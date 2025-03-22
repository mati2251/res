package http

import (
	"net/http"
	"res/pkg/vm"
)

// add checking if the sources is available
func PostJob(w http.ResponseWriter, r *http.Request) {
	virtualMachine := vm.New()
	go func() {
		err := vm.Spawn(virtualMachine)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := vm.ExecScript(virtualMachine, "script.sh"); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if err := vm.Kill(virtualMachine); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}()
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
