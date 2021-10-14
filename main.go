package main

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"path"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
)

var (
	ErrShutdown   = errors.New("shutdown")
	ErrUnknown    = errors.New("unknown")
	ctx           context.Context
	cancel        context.CancelFunc
	cfgSplitRegex = regexp.MustCompile(`autoexec_(.+)_([^_]+)\.cfg$`)
	shutdownRegex = regexp.MustCompile(`^\[.+]\[.+]: .+=(\d+) rcon='shutdown'$`)
)

func printUsage() {
	log.Print(`
Usage: 
	./teeworlds-start [zcatch_srv] ['-t0\d']
	
	1. Executable
	2. Optional regular expression to match the teeworlds server executable 
	3. requires 2 and adds a regular expression for config files.
	`)
}

func init() {
	ctx, cancel = signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
}

type Config struct {
	ConfigFile string
	Executable string
	ID         string
	Log        io.Writer
}

func (c *Config) Cmd() string {
	return fmt.Sprintf("./%s -f %s", c.Executable, c.ConfigFile)
}

func (c *Config) LogFile() (*os.File, error) {
	fileName := filepath.Base(c.ConfigFile)
	fileNameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	logFilePath := fmt.Sprintf("./logs/%s-%s.log", fileNameWithoutExt, time.Now().Format("2006-01-02-15:04:05.000"))
	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	return f, nil
}

func (c *Config) Run() (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("recovered run: %v", r)
		}
	}()

	cmd := exec.CommandContext(ctx, c.Executable, "-f", c.ConfigFile)
	logFile, err := c.LogFile()
	if err != nil {
		return err
	}
	defer func() {
		err := logFile.Close()
		if err != nil {
			log.Printf("failed to close log file: %s: %v\n", logFile.Name(), err)
		}
	}()
	logWriter, errC := analyzeLogs(logFile) // analyze server logs
	var stderr bytes.Buffer
	// write output into file
	cmd.Stdout = logWriter
	cmd.Stderr = &stderr

	err = cmd.Run()
	if err != nil && stderr.Len() > 0 {
		log.Printf("error: %s\n%s\n", c.Cmd(), stderr.String())
	} else {
		log.Printf("stopped: %s\n", c.Cmd())
	}
	logErr := <-errC
	if logErr != nil {
		return logErr
	}
	return err
}

func analyzeLogs(logFile ...io.Writer) (io.Writer, <-chan error) {
	errC := make(chan error, 1)

	pipeReader, pipeWriter := io.Pipe()
	logWriter := io.MultiWriter(io.MultiWriter(logFile...), pipeWriter)

	go func() {
		defer close(errC)
		defer pipeReader.Close()
		defer pipeWriter.Close()

		lineScanner := bufio.NewScanner(pipeReader)
		for lineScanner.Scan() {
			line := lineScanner.Text()
			if shutdownRegex.MatchString(line) {
				// admin stopped server ingame/econ
				errC <- ErrShutdown
				return
			}
		}
		// unknown shutdown reason
		errC <- nil
	}()

	return logWriter, errC
}

func constructConfigs(execPath, cfgPath, execMatch, cfgMatch string) []Config {
	execRegex := regexp.MustCompile(execMatch)
	cfgRegex := regexp.MustCompile(cfgMatch)

	// read files in current dir
	ef, err := os.ReadDir(execPath)
	if err != nil {
		log.Fatalf("failed to get files from current dir: %v\n", err)
	}
	configFiles := make([]Config, 0, 4)
	executables := make(map[string]bool, 1)

	// step 1 get executables from file list
	for _, file := range ef {
		if file.IsDir() {
			continue
		}

		fileName := file.Name()
		ext := filepath.Ext(fileName)
		if ext == "" || ext == "exe" {
			executable := filepath.Base(fileName)
			if execRegex.MatchString(executable) {
				log.Printf("found executable: %s\n", executable)
			} else {
				log.Printf("skipped executable: %s\n", executable)
				continue
			}
			executables[executable] = true
		}
	}

	cf, err := os.ReadDir(cfgPath)
	if err != nil {
		log.Fatalf("filed to get files from current dir: %v\n", err)
	}

	// get config files that actually match the executable
	for _, file := range cf {
		if file.IsDir() {
			continue
		}

		fileName := filepath.Base(file.Name())
		matches := cfgSplitRegex.FindStringSubmatch(fileName)
		if len(matches) > 0 {
			exec := matches[1]
			if !executables[exec] {
				// no matching executable, skip autoexec_(execuable)_(id).cfg file
				log.Printf("no executable found for config (expected '%s'): %s\n", exec, fileName)
				continue
			} else if !cfgRegex.MatchString(fileName) {
				log.Printf("skipped config due to regex mismatch: %s\n", fileName)
				continue
			}

			configFiles = append(configFiles, Config{
				ConfigFile: path.Join(cfgPath, fileName),
				Executable: path.Join(execPath, exec),
				ID:         matches[2],
			})
		}
	}
	return configFiles
}

func buildPathEnv(directories ...string) string {
	return strings.Join(directories, ":")
}

func main() {
	defer cancel()
	printUsage()

	executablesPath := "./executables"
	configsPath := "./configs"
	execRegex := ".*"
	cfgRegex := ".*"

	if len(os.Args) >= 2 {
		execRegex = os.Args[1]
	}
	if len(os.Args) >= 3 {
		cfgRegex = os.Args[2]
	}

	os.Setenv("PATH", buildPathEnv(os.Getenv("PATH"), executablesPath))
	wg := sync.WaitGroup{}
	cfgs := constructConfigs(executablesPath, configsPath, execRegex, cfgRegex)

	wg.Add(len(cfgs))
	for idx, c := range cfgs {
		go func(index int, c Config) {
			defer wg.Done()

			select {
			case <-ctx.Done():
				log.Printf("closing restart routine early: %s \n", c.Cmd())
				return
			case <-time.After(time.Duration(index) * time.Second):
				// continue
			}

			restartCounter := 0
			timeUntilRestart := time.Duration(0)
			for {
				select {
				case <-ctx.Done():
					log.Printf("closing restart routine: %s \n", c.Cmd())
					return
				default:
					if restartCounter == 0 {
						log.Printf("starting: %s\n", c.Cmd())
					} else {
						log.Printf("restarting: %s\n", c.Cmd())
					}

					if restartCounter > 5 && timeUntilRestart/time.Duration(restartCounter) < 60*time.Second {
						log.Printf("stopped: %s: reason: too many restarts within a short period\n", c.Cmd())
						return
					}

					start := time.Now()
					err := c.Run()
					timeUntilRestart = time.Since(start)
					if err != nil {
						log.Printf("stopped: %s: reason: %v\n", c.Cmd(), err)
						if strings.Contains(err.Error(), "exec format error") {
							log.Printf("slease use a different executable '%s', as it seems not to have been built for your operating system.\n", c.Executable)
							return
						} else if strings.Contains(err.Error(), "exit status 255") {
							log.Printf("make sure that your defined ports are not blocked: %s\n", c.ConfigFile)
							time.Sleep(10 * time.Second)
						}
					} else {
						log.Printf("stopped: %s: reason: %v\n", c.Cmd(), err)
					}

					time.Sleep(3 * time.Second)
				}
			}
		}(idx, c)
	}

	wg.Wait()
	log.Println("finished execution.")
}
