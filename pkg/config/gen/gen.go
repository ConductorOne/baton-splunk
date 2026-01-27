package main

import (
	cfg "github.com/conductorone/baton-splunk/pkg/config"
	"github.com/conductorone/baton-sdk/pkg/config"
)

func main() {
	config.Generate("splunk", cfg.Config)
}
