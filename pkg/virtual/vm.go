package virtual

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"res/pkg/config"
	"res/pkg/db"
	"time"

	"golang.org/x/crypto/ssh"
)

type JobFile = string

type JobService struct {
	Config  *config.Config
	Queries *db.Queries
}

const (
	Log  JobFile = "log.txt"
	Spec JobFile = "spec.json"
	Pid JobFile = "qemu.pid"
)

func filePath(job *db.Job, fileType JobFile) string {
	return fmt.Sprintf("%s/%s", job.BasePath, fileType)
}

func CreateSpecFile(job *db.Job) error {
	specFile, err := os.Create(filePath(job, Spec))
	if err != nil {
		return fmt.Errorf("failed to create spec file: %v", err)
	}
	defer close(specFile)
	err = json.NewEncoder(specFile).Encode(job)
	if err != nil {
		return fmt.Errorf("failed to encode spec file: %v", err)
	}
	return nil
}

func close(closer io.Closer) {
	err := closer.Close()
	if err != nil {
		log.Printf("failed to close: %v", err)
	}
}

func closeIgnore(closer io.Closer) {
	_ = closer.Close()
}

// TODO: Implement custom dir for scripts per vm
// TODO: Port per open port
func Spawn(vm *db.Vm) (*exec.Cmd, error) {
	qemu_args := []string{
		"-m", fmt.Sprint(vm.Memory),
		"-hda", vm.Image,
		"-smp", fmt.Sprint(vm.Cpu),
		"-device", "qemu-xhci",
		"-no-reboot",
		"-nic", fmt.Sprintf("user,hostfwd=tcp::%d-:22", 2222),
		"--enable-kvm",
		"-display", "none",
		"-fsdev", "local,id=fs1,path=/home/mateusz/Test/,security_model=mapped",
		"-device", "virtio-9p-pci,fsdev=fs1,mount_tag=hostshare",
	}

	cmd := exec.Command("qemu-system-x86_64", qemu_args...)
	err := cmd.Start()
	if err != nil {
		return nil, fmt.Errorf("failed to start vm: %v", err)
	}
	return cmd, nil
}

// TODO: implement custom timeout
func ExecScript(j *db.Job, vm *db.Vm) error {
	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var client *ssh.Client
	var err error
	for range 10 {
		client, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%d", vm.Port), config)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer close(client)

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer closeIgnore(session)
	err = session.Run("mount -t 9p -o trans=virtio hostshare /mnt/share")
	if err != nil {
		return fmt.Errorf("failed to mount: %v", err)
	}

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	err = os.MkdirAll(j.BasePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create base dir: %v", err)
	}
	logFile, err := os.Create(filePath(j, Log))
	if err != nil {
		return fmt.Errorf("failed to create log file: %v", err)
	}
	defer close(logFile)
	session.Stdout = logFile
	session.Stderr = logFile
	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdin: %v", err)
	}
	err = session.Run(fmt.Sprintf("/mnt/share/%s", "script.sh"))
	if err != nil {
		return fmt.Errorf("failed to run script: %v", err)
	}
	_ = stdin.Close()
	return nil
}
