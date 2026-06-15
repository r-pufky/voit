package voit

import (
	"fmt"
	"os"

	"github.com/k0kubun/pp/v3"
)

// TODO - io reader/writer for prompts in functions should accept w/r for testing purposes.

// Resolve name collisions after task modified files with specified operation.
func (files VoitFiles) ResolveCollisions(vCfg Config, verbose bool) {
	collisions := make(map[string]int)

	for i := range files { // Modifying must use index to reference.
		desc := files[i].Mark.Desc.Text
		files[i].Format(&vCfg)

		for {
			_, sliceCollision := collisions[files[i].Target]
			_, err := os.Stat(files[i].Target)
			FSCollision := !os.IsNotExist(err)

			if verbose {
				pp.Printf("\nSlice collision: %+v\nFS collision:    %+v\n", sliceCollision, FSCollision)
			}

			// Mark seen, no collision on FS or VoitFiles.
			if !sliceCollision && !FSCollision {
				collisions[files[i].Target] = 1
				break
			}

			count := collisions[files[i].Target]
			// Collision occurred and target has not been seen. This is a FS
			// collision as slice would have been marked above. Set count to 1 and
			// mark target seen.
			if count == 0 {
				count = 1
			}
			collisions[files[i].Target] = count + 1

			if desc != "" {
				files[i].Mark.Desc.Text = fmt.Sprintf("%s %d", desc, count)
			} else {
				files[i].Mark.Desc.Text = fmt.Sprintf("%d", count)
			}
			files[i].Format(&vCfg)
		}

		if verbose && files[i].Matched {
			pp.Printf("Matched: %v\n", files[i])
		}
	}
}
