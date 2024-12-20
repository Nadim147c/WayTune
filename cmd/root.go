package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	Version = "Git"

	logFile *os.File
)

var Log = func(a ...any) {
	WriteLog("WayTune", a...)
}

var rootCmd = &cobra.Command{
	Use:     "waytune",
	Version: Version,
	Short:   "A collection of custom  waybar modules",
}

func Execute() {
	defer logFile.Close()
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.SetEnvPrefix("waytune")
	viper.BindEnv("log_file")

	logFilePath := viper.GetString("log_file")

	if logFilePath == "" {
		fmt.Fprintln(os.Stderr, "No log file specified. Exiting.")
		return
	}

	if err := os.MkdirAll(filepath.Dir(logFilePath), os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create log file directories: %v\n", err)
		return
	}

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to open log file: %v\n", err)
		return
	}
	logFile = file
}

func WriteLog(scope string, a ...any) {
	fmt.Fprintf(os.Stderr, "[%s] %v\n", scope, fmt.Sprint(a...))
	if logFile != nil {
		timestamp := time.Now().UTC().Format(time.DateTime)
		fmt.Fprintf(logFile, "[%s] [%s] %v\n", timestamp, scope, fmt.Sprint(a...))
	}
}

func RunCommand(bin string, args ...string) error {
	fmt.Fprintf(os.Stderr, "[sys] running: %s ", bin)
	for _, a := range args {
		fmt.Fprintf(os.Stderr, "%q ", a)
	}
	fmt.Print("\n")

	cmd := exec.Command(bin, args...)

	cmd.Stdout = nil
	cmd.Stderr = nil

	return cmd.Run()
}

func Output(bin string, args ...string) (string, error) {
	cmd := exec.Command(bin, args...)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(string(output), "\n"), nil
}

func UpdateWaybar() error {
	return SendSignal("^waybar$", SIGRTMIN+4)
}

func SendSignal(processName string, signal int) error {
	cmd := exec.Command("pgrep", processName)
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("failed to find processes matching %q: %w", processName, err)
	}

	pidStrings := strings.Fields(string(output))
	if len(pidStrings) == 0 {
		return fmt.Errorf("no processes found matching %q", processName)
	}

	for _, pidStr := range pidStrings {
		pid, err := strconv.Atoi(pidStr)
		if err != nil {
			return fmt.Errorf("invalid PID %q: %w", pidStr, err)
		}

		// Send the signal
		err = syscall.Kill(pid, syscall.Signal(signal))
		if err != nil {
			return fmt.Errorf("failed to send signal to PID %d: %w", pid, err)
		}
	}

	return nil
}
