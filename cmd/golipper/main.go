package main

import (
	"io"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"github.com/kelseyhightower/envconfig"
	"golang.org/x/xerrors"
)

type Configuration struct {
	Port int
}

func main() {
	os.Exit(mainModule())
}

func mainModule() int {
	errorExitCode := 1
	var config Configuration
	var port = 8377
	err := envconfig.Process("golipper", &config)
	if err != nil {
		log.Fatal(err.Error())
		return errorExitCode
	}

	if config.Port == 0 {
		log.Println("GOLIPPER_PORT environment variable is not defined. Using default port: " + strconv.Itoa(port))
	} else {
		port = config.Port
	}

	ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	log.Println("Listening on localhost:" + strconv.Itoa(port))
	if err != nil {
		log.Fatal(err.Error())
		return errorExitCode
	}

	for {
		err = loop(ln)
		if err != nil {
			log.Fatal(err.Error())
			return errorExitCode
		}
	}

}

func loop(ln net.Listener) error {

	buffer := make([]byte, 1024*1024*1024) // GB

	for {
		message := ""
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err.Error())
			return err
		}
		defer conn.Close()

		n, err := conn.Read(buffer)
		if err != nil {
			if err != io.EOF {
				return xerrors.Errorf("Connection read error: %w", err)
			}
			break
		}
		message = message + string(buffer[:n])
		log.Print("Recieved message:\n" + message)

		err = putStringToClipboard(message)
		if err != nil {
			return err
		}
		err = conn.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func logOutputPipe(out io.ReadCloser, typeName string) error {
	defer out.Close()
	buffer := make([]byte, 1024)
	message := ""

	for {
		n, err := out.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Fatal("Output read error: %w", err)
				// return xerrors.Errorf("Output read error: %w", err)
			}
			break
		}
		message = message + string(buffer[:n])
	}

	log.Print(typeName + ": " + message)
	return nil
}

func putStringToClipboard(content string) error {
	var clipCommand string
	var cmd *exec.Cmd

	log.Println("Putting string to clipboard:\n" + content)

	if runtime.GOOS == "windows" {
		clipCommand = "clip.exe"
		cmd = exec.Command(clipCommand)
	} else if runtime.GOOS == "darwin" {
		clipCommand = "pbcopy"
		cmd = exec.Command("sh", "-c", clipCommand)
	} else if runtime.GOOS == "linux" {
		clipCommand = "xsel -ip && xsel -op | xsel -ib"
		cmd = exec.Command("sh", "-c", clipCommand)
	} else {
		return xerrors.New("Unsupported platform error")
	}

	stdin, err := cmd.StdinPipe()

	if err != nil {
		return xerrors.Errorf("Stdin pipe error: %w", err)
	}

	go func() {
		defer stdin.Close()
		_, err := io.WriteString(stdin, content)
		if err != nil {
			log.Fatal(xerrors.Errorf("Stdin put error: %w", err).Error())
		}
	}()

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return xerrors.Errorf("Stderr pipe error: %w", err)
	}
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		return xerrors.Errorf("Stdout pipe error: %w", err)
	}

	go logOutputPipe(stdout, "stdout")
	go logOutputPipe(stderr, "stderr")

	cmd.Run()
	if err != nil {
		return xerrors.Errorf("Command run error: %w", err)
	}
	exitCode := cmd.ProcessState.ExitCode()

	log.Println("Exit code: " + strconv.Itoa(exitCode))

	return nil
}
