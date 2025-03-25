package vm

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

var jobCounter, vmCounter int64 = 0, 0

type JobFile = string

const (
	Log JobFile = "log.txt"
)

type Job struct {
	Id      int64           `json:"id"`
	Script  string          `json:"script"`
	Vm      *VirtualMachine `json:"vm"`
	baseDir string
}

func NewJob(script string) *Job {
	jobCounter++
	return &Job{
		Id:      jobCounter,
		Script:  script,
		baseDir: "/tmp",
		Vm:      NewVM(),
	}
}

func (job *Job) BaseDir() string {
  return fmt.Sprintf("%s/%d", job.baseDir, job.Id)
}

func (job *Job) FilePath(file JobFile) string {
	return fmt.Sprintf("%s/%d/%s", job.baseDir, job.Id, file)
}

type VirtualMachine struct {
	Id     int64  `json:"id"`
	Image  string `json:"image"`
	Memory int    `json:"memory"`
	Cpus   int    `json:"cpus"`
	Port   int    `json:"port"`
	cmd    *exec.Cmd
}

func NewVM() *VirtualMachine {
	vmCounter++
	return &VirtualMachine{
		Id:     vmCounter,
		Image:  "/home/mateusz/Images/qemu/debian/clean-ssh.qcow2",
		Memory: 2048,
		Cpus:   2,
		Port:   2222,
		cmd:    nil,
	}
}

// TODO: Implement custom dir for scripts per vm
// TODO: Port per open port
func (vm *VirtualMachine) Spawn() error {
	qemu_args := []string{
		"-m", fmt.Sprint(vm.Memory),
		"-hda", vm.Image,
		"-smp", fmt.Sprint(vm.Cpus),
		"-device", "qemu-xhci",
		"-no-reboot",
		"-nic", fmt.Sprintf("user,hostfwd=tcp::%d-:22", vm.Port),
		"--enable-kvm",
		"-display", "none",
		"-fsdev", "local,id=fs1,path=/home/mateusz/Test/,security_model=mapped",
		"-device", "virtio-9p-pci,fsdev=fs1,mount_tag=hostshare",
	}

	vm.cmd = exec.Command("qemu-system-x86_64", qemu_args...)
	err := vm.cmd.Start()
	if err != nil {
		return fmt.Errorf("failed to start vm: %v", err)
	}
	return nil
}

// TODO: implement custom timeout
func (j *Job) ExecScript() error {
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
		client, err = ssh.Dial("tcp", fmt.Sprintf("localhost:%d", j.Vm.Port), config)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		return fmt.Errorf("failed to dial: %v", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
	defer session.Close()
	err = session.Run("mount -t 9p -o trans=virtio hostshare /mnt/share")
	if err != nil {
		return fmt.Errorf("failed to mount: %v", err)
	}

	session, err = client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create session: %v", err)
	}
  err = os.MkdirAll(j.BaseDir(), 0755)
  if err != nil {
    return fmt.Errorf("failed to create base dir: %v", err)
  }
	logFile, err := os.Create(j.FilePath(Log))
  if err != nil {
    return fmt.Errorf("failed to create log file: %v", err)
  }
	defer logFile.Close()
	session.Stdout = logFile
	session.Stderr = logFile
	stdin, err := session.StdinPipe()
	err = session.Run(fmt.Sprintf("/mnt/share/%s", j.Script))
	if err != nil {
		return fmt.Errorf("failed to run script: %v", err)
	}
	stdin.Close()
	return nil
}

func (vm *VirtualMachine) Kill() error {
	err := vm.cmd.Process.Kill()
	if err != nil {
		return fmt.Errorf("failed to kill vm: %v", err)
	}
	return nil
}
