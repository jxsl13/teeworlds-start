package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
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
	//errShutdown   = errors.New("shutdown")
	//shutdownRegex = regexp.MustCompile(`^\[.+]\[.+]: .+=(\d+) rcon='shutdown'$`)

	cfgSplitRegex = regexp.MustCompile(`autoexec_(.+)_([^_]+)\.cfg$`)

	debug *bool

	executablesRegex *regexp.Regexp
	configRegex      *regexp.Regexp

	executablesDir string
	configsDir     string
	startTimes     []time.Time = make([]time.Time, 0)
	stopTimes      []time.Time = make([]time.Time, 0)
)

func printUsage() {
	log.Print(`
Usage: 
	./teeworlds-start --executables 'zcatch_.*' --configs '_t0\d.*'

	--configs-dir		./configs
	--executables-dir	./executables

	--times '2021-04-01-13.00.00,2021-04-02-18.00.00'
		Provide a start stop schedule list separated by ','
	
	`)
}

func init() {

	help := flag.Bool("help", false, "show help screen")
	debug = flag.Bool("debug", false, "use it to show more information")
	cfgDir := flag.String("configs-dir", "./configs", "directorythat contains your configuration files (autoexec.cfg).")
	cfg := flag.String("configs", ".*", "regular expression to match the config file suffix")
	execDir := flag.String("executables-dir", "./executables", "directory that contains your teeworlds gameserver executables (zcatch_srv, gctf_srv).")
	exec := flag.String("executables", ".*", "regular expression to match the server executable")
	times := flag.String("times", "", "start,stop,start,stop.. dates in the format of 2006-12-30-15.04.05")
	flag.Parse()

	if help != nil && *help {
		printUsage()
		os.Exit(0)
	}
	executablesDir = *execDir
	configsDir = *cfgDir

	var err error
	executablesRegex, err = regexp.Compile(*exec)
	if err != nil {
		log.Fatalln(err)
	}

	configRegex, err = regexp.Compile(*cfg)
	if err != nil {
		log.Fatalln(err)
	}

	if *times != "" {
		parts := strings.Split(*times, ",")
		if len(parts)%2 != 0 {
			fmt.Println("--times requires the number of dates to be a multiple of 2")
			os.Exit(1)
		}

		formats := []string{
			"2006-01-02-15.04.05",
			"2006-01-02-15.04",
			"2006-01-02-15",
			"2006-01-02",
		}

		for idx, part := range parts {
			var (
				t   time.Time
				err error
			)
			for _, format := range formats {
				t, err = time.Parse(format, part)
				if err != nil {
					continue
				}
				break
			}

			if err != nil {
				fmt.Printf("'%s' is not a valid date and time value\n", part)
				os.Exit(1)
			} else if idx%2 == 0 {
				// startup
				startTimes = append(startTimes, t)
			} else {
				//shutdown
				startupSchedule := startTimes[len(startTimes)-1]
				if startupSchedule == t || startupSchedule.After(t) {
					fmt.Printf("WARING: shutdown schedule %s lies before or at the startup schedule %s\n", t.String(), startupSchedule.String())
					fmt.Println("WARNING: moving shutdown schedule 5 seconds after startup schedule.")
					t = startupSchedule.Add(5 * time.Second)
				}
				stopTimes = append(stopTimes, t)
			}

		}
	}
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()
	defer printUsage()

	if len(startTimes) > 0 && len(startTimes) == len(stopTimes) {
		log.Printf("scheduled startups: %v\n", startTimes)
		log.Printf("scheduled shutdown: %v\n", stopTimes)
	} else if len(startTimes) > 0 && len(startTimes) != len(stopTimes) {
		log.Println("stat/stop times mismatch")
		os.Exit(1)
	}

	os.Setenv("PATH", buildPathEnv(os.Getenv("PATH"), executablesDir))
	wg := sync.WaitGroup{}
	cfgs := constructConfigs(ctx, executablesDir, configsDir, executablesRegex, configRegex, startTimes, stopTimes)

	wg.Add(len(cfgs))
	for idx, c := range cfgs {
		c.StartupOffset = time.Second * time.Duration(idx)
		go func(c Config) {
			defer wg.Done()

			err := c.Run()
			if err != nil {
				log.Printf("unexpected shutdown: %s: %v\n", c.Cmd(), err)
				return
			}
			log.Printf("successful shutdown: %s\n", c.Cmd())
		}(c)
	}

	wg.Wait()
	log.Println("finished execution.")
}

func DebugPrintln(a ...interface{}) {
	if *debug {
		log.Println(a...)
	}
}

func DebugPrintf(format string, a ...interface{}) {
	if *debug {
		log.Printf(format, a...)
	}
}

type Config struct {
	ConfigFile string
	Executable string
	ID         string
	Log        io.Writer

	StartupOffset time.Duration
	StartTimes    []time.Time
	StopTimes     []time.Time

	ShutdownContext context.Context
}

func (c *Config) Cmd() string {
	return fmt.Sprintf("./%s -f %s", c.Executable, c.ConfigFile)
}

func (c *Config) LogFile() (*os.File, error) {
	fileName := filepath.Base(c.ConfigFile)
	fileNameWithoutExt := fileName[:len(fileName)-len(filepath.Ext(fileName))]
	logFilePath := fmt.Sprintf("./logs/%s-%s.log", fileNameWithoutExt, time.Now().Format("2006-01-02-15.04.05"))
	f, err := os.OpenFile(logFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, fmt.Errorf("failed to create log file: %w", err)
	}
	return f, nil
}

func (c *Config) runSingleWithRestart(shutdownContext context.Context) (err error) {
	restartCounter := 0
	timeUntilRestart := time.Duration(0)
	for {
		select {
		case <-shutdownContext.Done():
			log.Printf("closing restart routine: %s\n", c.Cmd())
			return
		default:
			if restartCounter == 0 {
				log.Printf("starting: %s\n", c.Cmd())
			} else {
				log.Printf("restarting: %s\n", c.Cmd())
			}

			if restartCounter > 5 && timeUntilRestart/time.Duration(restartCounter) < 60*time.Second {
				log.Printf("stopped: %s: reason: too many restarts within a short period\n", c.Cmd())
				return errors.New("too many restarts within a short period")
			}

			start := time.Now()
			err := c.runSingle(shutdownContext)
			timeUntilRestart = time.Since(start)
			if err != nil {
				log.Printf("stopped: %s: reason: %v\n", c.Cmd(), err)
				if strings.Contains(err.Error(), "exec format error") {
					log.Printf("please use a different executable '%s', as it seems not to have been built for your operating system.\n", c.Executable)
					return err
				} else if strings.Contains(err.Error(), "exit status 255") {
					log.Printf("make sure that your defined ports are not blocked: %s\n", c.ConfigFile)
					time.Sleep(10 * time.Second)
				}
			} else {
				log.Printf("stopped: %s: reason: manual shutdown\n", c.Cmd())
			}

			time.Sleep(3 * time.Second)
		}
	}
}

func (c *Config) runSingle(shutdownContext context.Context) (err error) {
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("recovered run: %v", r)
		}
	}()

	cmd := exec.CommandContext(shutdownContext, c.Executable, "-f", c.ConfigFile)
	logFile, err := c.LogFile()
	if err != nil {
		return err
	}
	defer logFile.Close()

	var stderr bytes.Buffer
	// write output into file
	cmd.Stdout = logFile
	cmd.Stderr = &stderr

	err = cmd.Run()

	if err != nil && stderr.Len() > 0 {
		log.Printf("error: %s\n%s\n", c.Cmd(), stderr.String())
	} else {
		log.Printf("stopped: %s\n", c.Cmd())
	}

	select {
	case <-shutdownContext.Done():
		// command stopped due to shutdown context.
		return nil
	default:
		// command stopped due to error
		return err
	}
}

func (c *Config) Run() (err error) {

	if len(c.StartTimes) == 0 {
		select {
		case <-c.ShutdownContext.Done():
			log.Printf("shutdown before starting: %s\n", c.Cmd())
			return nil
		case <-time.After(c.StartupOffset):
			return c.runSingleWithRestart(c.ShutdownContext)
		}
	}

	// start/stop logic
	if len(c.StartTimes) != len(c.StopTimes) {
		return errors.New("start/stop times mismatch")
	}

	select {
	case <-c.ShutdownContext.Done():
		log.Printf("shutdown before starting: %s\n", c.Cmd())
		return nil
	case <-time.After(c.StartupOffset / 4):
		// make console output ordered
	}

	for idx, start := range c.StartTimes {
		now := time.Now()
		offsetStartUp := start.Add(c.StartupOffset)
		durationUntilNextStartup := offsetStartUp.Sub(now)
		if durationUntilNextStartup < 0 {
			durationUntilNextStartup = c.StartupOffset
		}
		shutdownDeadline := c.StopTimes[idx]
		deadlineContext, _ := context.WithDeadline(c.ShutdownContext, shutdownDeadline)

		log.Printf("startup scheduled: %v: %s (%s)", offsetStartUp, c.Cmd(), durationUntilNextStartup.String())

		select {
		case <-c.ShutdownContext.Done():
			log.Printf("shutdown: %s\n", c.Cmd())
			return nil
		case <-time.After(durationUntilNextStartup):
			log.Printf("scheduled startup: %s\n", c.Cmd())
			log.Printf("will be shut down: %v: %s (%s)\n", shutdownDeadline, c.Cmd(), time.Until(shutdownDeadline).String())
			err := c.runSingleWithRestart(deadlineContext)
			if err != nil {
				log.Printf("unexpected shutdown: %s: %v\n", c.Cmd(), err)
				return err
			}
			log.Printf("scheduled shutdown: %s\n", c.Cmd())
		}
	}

	log.Printf("exhausted startup schedules: %s\n", c.Cmd())
	return nil
}

func constructConfigs(shutdownContext context.Context, execPath, cfgPath string, execRegex, cfgRegex *regexp.Regexp, startupTimes, stopTimes []time.Time) []Config {

	execMatch := execRegex.String()

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
				log.Printf("found executable (matchig '%s'): %s\n", execMatch, executable)
				executables[executable] = true
			} else {
				log.Printf("skipped executable (not matching '%s'): %s\n", execMatch, executable)
			}
		}
	}

	cf, err := os.ReadDir(cfgPath)
	if err != nil {
		log.Fatalf("filed to get files from current dir: %v\n", err)
	}

	// get config files that actually match the executable
	for idx, file := range cf {
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

				ShutdownContext: shutdownContext,

				StartupOffset: time.Duration(idx) * time.Second,
				StartTimes:    startupTimes,
				StopTimes:     stopTimes,
			})
		}
	}
	return configFiles
}

func buildPathEnv(directories ...string) string {
	return strings.Join(directories, ":")
}
