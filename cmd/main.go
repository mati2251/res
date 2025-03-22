package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"golang.org/x/crypto/ssh"
)

func help(cmd string) {
	fmt.Printf("Usage: ./%s <script>\n", cmd)
	os.Exit(1)
}

func main() {
	if len(os.Args) < 2 {
		help(os.Args[0])
		os.Exit(1)
	}

	script := os.Args[1]
	qemu_args := []string{
		"-m", "2048",
		"-hda", "/home/mateusz/Images/qemu/debian/clean-ssh.qcow2",
		"-device", "qemu-xhci",
		"-no-reboot",
		"-nic", "user,hostfwd=tcp::2222-:22",
		"--enable-kvm",
		"-display", "none",
		"-pidfile", "/tmp/qemu.pid",
		"-fsdev", "local,id=fs1,path=/home/mateusz/Test/,security_model=mapped",
		"-device", "virtio-9p-pci,fsdev=fs1,mount_tag=hostshare",
	}

	cmd := exec.Command("qemu-system-x86_64", qemu_args...)
	err := cmd.Start()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
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
		client, err = ssh.Dial("tcp", "localhost:2222", config)
		if err == nil {
			break
		}
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer session.Close()
	err = session.Run("mount -t 9p -o trans=virtio hostshare /mnt/share")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	session, err = client.NewSession()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	session.Stdout = os.Stdout
	session.Stderr = os.Stderr
	stdin, err := session.StdinPipe()
	err = session.Run(fmt.Sprintf("/mnt/share/%s", script))
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	stdin.Close()
	err = cmd.Process.Kill()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
