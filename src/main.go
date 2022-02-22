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
    err := envconfig.Process("golipper", &config)
    if err != nil {
        log.Fatal(err.Error())
        return errorExitCode
    }

    if config.Port == 0 {
        log.Fatal("Please define GOLIPPER_PORT environment variable.")
        return errorExitCode
    }

    ln, err := net.Listen("tcp", ":" + strconv.Itoa(config.Port))
    log.Println("Listening on localhost:" + strconv.Itoa(config.Port))
    if err != nil {
        log.Fatal(err.Error())
        return errorExitCode
    }

    for {
        conn, err := ln.Accept()
        if err != nil {
            log.Fatal(err.Error())
            return errorExitCode
        }
        go handleConnection(conn)
    }
}

func handleConnection(conn net.Conn) error {
    defer conn.Close()

    buffer := make([]byte, 1024)

    message := ""
    for {
        n, err := conn.Read(buffer)
        if err != nil {
            if err != io.EOF {
                return xerrors.Errorf("Connection read error: %w", err)
            }
            break
        }
        message = message + string(buffer[:n])
        log.Print(message)
    }

    err := putStringToClipboard(message)
    if err != nil {
        return err
    }

    writeBytes := []byte("helloasdfasdf")
    _, err = conn.Write(writeBytes)

    if err != nil {
        return err
    }

    return nil
}

func logOutputPipe(out io.ReadCloser) error {
    defer out.Close()
    buffer := make([]byte, 1024)
    message := ""

    for {
        n, err := out.Read(buffer)
        if err != nil {
            if err != io.EOF {
                log.Fatal("Output read error: %w", err)
                return xerrors.Errorf("Output read error: %w", err)
            }
            break
        }
        message = message + string(buffer[:n])
    }

    log.Print(message)
    return nil
}

func putStringToClipboard(content string) error {
    var clipCommand string

    log.Println("Putting string: \"" + content + "\" to clipboard")

    clipCommand = "not found"
    if runtime.GOOS == "windows" {
        clipCommand = "clip.exe"
    }

    cmd := exec.Command(clipCommand)
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

    go logOutputPipe(stderr)
    go logOutputPipe(stdout)


    cmd.Run()
    if err != nil {
        return xerrors.Errorf("Command run error: %w", err)
    }
    exitCode := cmd.ProcessState.ExitCode()

    log.Println("Exit code: " + strconv.Itoa(exitCode))

    return nil
}
