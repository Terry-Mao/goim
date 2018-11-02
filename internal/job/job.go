package job

import "github.com/Terry-Mao/goim/internal/job/conf"

// Job is push job.
type Job struct {
	c *conf.Config
}

// New new a push job.
func New(c *conf.Config) *Job {
	j := &Job{
		c: c,
	}
	return j
}
