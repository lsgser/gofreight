package main

import (
	"fmt"
	"os"
	"os/exec"
)

func registerScheduleCommands() {
	register(command{
		Name: "schedule:run", Category: "schedule", Description: "Run due scheduled tasks",
		Run: handleScheduleRun,
	})
	register(command{
		Name: "schedule:list", Category: "schedule", Description: "List scheduled tasks",
		Run: handleScheduleList,
	})
}

func handleScheduleRun(args []string) {
	fmt.Println("Running scheduler — define tasks in bootstrap/schedule.go and call schedule:run from cron")
	fmt.Println("Example crontab: * * * * * cd /app && gofreight schedule:run")
	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "GOFREIGHT_SCHEDULE_RUN=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		os.Exit(1)
	}
}

func handleScheduleList(args []string) {
	fmt.Println("Scheduled tasks are defined in bootstrap/schedule.go")
	fmt.Println("Use app scheduler via schedule.New() — see docs/scheduling.md")
}

func registerRouteCacheCommand() {
	register(command{
		Name: "route:cache", Category: "route", Description: "Cache route metadata for URL generation",
		Run: handleRouteCache,
	})
}

func handleRouteCache(args []string) {
	cmd := exec.Command("go", "run", ".")
	cmd.Env = append(os.Environ(), "GOFREIGHT_ROUTE_CACHE=1")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "route:cache failed — ensure you are in an app directory with main.go\n")
		os.Exit(1)
	}
}
