package main

import (
	"os"

	"github.com/cardil/qe-clusterlogging/internal/collector"
)

func main() {
	collector.ServeOrDie(os.Exit)
}
