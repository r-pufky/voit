package voit

import (
	"fmt"

	"github.com/k0kubun/pp/v3"
)

// Resolve name collisions after task modified files with specified operation.
func (files VoitFiles) ResolveCollisions(vCfg Config, verbose bool) {
	collisions := make(map[string]int)

	for _, f := range files {
		desc := f.Mark.Desc.Text
		f.Format(&vCfg)

		for {
			if _, exists := collisions[f.Target]; !exists {
				collisions[f.Target] = 1
				break // No Collision.
			}

			count := collisions[f.Target]
			collisions[f.Target]++

			if desc != "" {
				// Standard collision.
				f.Mark.Desc.Text = fmt.Sprintf("%s_%d", desc, count)
			} else if len(f.Mark.Tags.Items) > 0 {
				// No description, tags.
				f.Mark.Desc.Text = fmt.Sprintf("%d", count)
			} else {
				// No description, no tags.
				f.Mark.Desc.Text = fmt.Sprintf("%d", count)
			}
			f.Format(&vCfg)
		}

		if verbose && f.Matched {
			pp.Printf("Matched: %v\n", f)
		}
	}
}
