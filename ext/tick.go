package ext

import "time"

func Tick(duration time.Duration, action func() error) error {
	for {
		select {
		case <-time.Tick(duration):
			if err := action(); err != nil {
				return err
			}
		}
	}
}
