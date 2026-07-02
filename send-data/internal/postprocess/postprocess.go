package postprocess

import (
	"log"
	"os"
	"os/exec"
)

type Runner struct {
	enabled bool
	script  string
}

func New(enabled bool, script string) *Runner {
	return &Runner{
		enabled: enabled,
		script:  script,
	}
}

func (r *Runner) Run(dataDir string) {
	if !r.enabled || r.script == "" {
		return
	}

	go func() {
		cmd := exec.Command(r.script, dataDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		if err := cmd.Run(); err != nil {
			log.Printf("post-process failed: script=%s dir=%s err=%v", r.script, dataDir, err)
		}
	}()
}
