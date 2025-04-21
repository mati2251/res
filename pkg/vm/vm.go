package vm

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"res/pkg/config"
	"strconv"
	"syscall"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	Log    JobFile = "log.txt"
	Spec   JobFile = "spec.json"
	Pid    JobFile = "qemu.pid"
	Script JobFile = "script"
)

type JobFile = string

type JobService struct {
	Config *config.Config
}

type JobQuery struct {
	Script   string `json:"script"`
	Memory   int    `json:"memory"`
	Cpu      int    `json:"cpu"`
	Image    string `json:"image"`
	BasePath string `json:"base_path"`
}

type Job struct {
	Query JobQuery `json:"query"`
	cmd   *exec.Cmd
	port  int
}

func NewJob(query JobQuery) (Job, error) {
	port, err := getFreePort()
	if err != nil {
		return Job{}, fmt.Errorf("failed to get free port: %v", err)
	}
	err = os.MkdirAll(query.BasePath, 0755)
	if err != nil {
		return Job{}, fmt.Errorf("failed to create base dir: %v", err)
	}
	query.BasePath, err = filepath.Abs(query.BasePath)
	if err != nil {
		return Job{}, fmt.Errorf("failed to get absolute path: %v", err)
	}
	query.Image, err = filepath.Abs(query.Image)
	if err != nil {
		return Job{}, fmt.Errorf("failed to get absolute path: %v", err)
	}
	query.Script, err = filepath.Abs(query.Script)
	if err != nil {
		return Job{}, fmt.Errorf("failed to get absolute path: %v", err)
	}
	job := Job{
		Query: query,
		cmd:   nil,
		port:  port,
	}
	return job, nil
}

func getFreePort() (int, error) {
	var addr *net.TCPAddr
	addr, err := net.ResolveTCPAddr("tcp", "")
	if err != nil {
		return 0, fmt.Errorf("failed to resolve address: %v", err)
	}
	ln, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, fmt.Errorf("failed to listen on address: %v", err)
	}
	defer close(ln)
	return ln.Addr().(*net.TCPAddr).Port, nil
}

func (j Job) filePath(fileType JobFile) string {
	return fmt.Sprintf("%s/%s", j.Query.BasePath, fileType)
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
func (j Job) Spawn() error {
	qemu_args := []string{
		"-m", fmt.Sprint(j.Query.Memory),
		"-hda", j.Query.Image,
		"-smp", fmt.Sprint(j.Query.Cpu),
		"-device", "qemu-xhci",
		"-no-reboot",
		"-nic", fmt.Sprintf("user,hostfwd=tcp::%d-:22", j.port),
		"--enable-kvm",
		"-display", "none",
		"-fsdev", fmt.Sprintf("local,id=fs1,path=%s,security_model=mapped", j.Query.BasePath),
		"-device", "virtio-9p-pci,fsdev=fs1,mount_tag=hostshare",
		"-pidfile", j.filePath(Pid),
		"-daemonize",
	}

	cmd := exec.Command("qemu-system-x86_64", qemu_args...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start vm: %s %v", out, err)
	}
	return nil
}

func (j Job) Kill() error {
	path := j.filePath(Pid)
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("failed to open pid file: %v", err)
	}
	defer close(file)
	pidRaw, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("failed to read pid file: %v", err)
	}
	pid, err := strconv.Atoi(string(pidRaw[:len(pidRaw)-1]))
	if err != nil {
		return fmt.Errorf("failed to convert pid to int: %v", err)
	}
	err = syscall.Kill(pid, syscall.SIGKILL)
	if err != nil {
		return fmt.Errorf("failed to kill process: %v", err)
	}
	return nil
}

func (j Job) copyScript() (JobFile, error) {
	filename := filepath.Base(j.Query.Script)
	dst, err := os.Create(j.filePath(filename))
	if err != nil {
		return "", fmt.Errorf("failed to create script file: %v", err)
	}
	defer close(dst)
	src, err := os.Open(j.Query.Script)
	if err != nil {
		return "", fmt.Errorf("failed to open script file: %v", err)
	}
	defer close(src)
	_, err = io.Copy(dst, src)
	if err != nil {
		return "", fmt.Errorf("failed to copy script file: %v", err)
	}
	err = os.Chmod(j.filePath(filename), 0755)
	if err != nil {
		return "", fmt.Errorf("failed to chmod script file: %v", err)
	}
	return filename, nil
}

// TODO: implement custom timeout
func (j Job) ExecScript() error {
	filename, err := j.copyScript()
	if err != nil {
		return fmt.Errorf("failed to copy script: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.Password("root"),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	var client *ssh.Client
	for range 10 {
		client, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%d", j.port), config)
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
	err = os.MkdirAll(j.Query.BasePath, 0755)
	if err != nil {
		return fmt.Errorf("failed to create base dir: %v", err)
	}
	logFile, err := os.Create(j.filePath(Log))
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
	err = session.Run(fmt.Sprintf("/mnt/share/%s", filename))
	if err != nil {
		return fmt.Errorf("failed to run script: %v", err)
	}
	_ = stdin.Close()
	return nil
}
