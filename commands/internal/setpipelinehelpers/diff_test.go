package setpipelinehelpers_test

import (
	"fmt"

	"gopkg.in/yaml.v2"

	sph "github.com/concourse/fly/commands/internal/setpipelinehelpers"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gbytes"
)

type DiffPart struct {
	Name string
}

var _ = Describe("Diff", func() {
	Describe("Render", func() {
		var diff sph.Diff
		BeforeEach(func() {
			diff = sph.Diff{
				Before: DiffPart{Name: "beforeName"},
				After:  DiffPart{Name: "afterName"},
			}
		})

		FIt("returns true", func() {
			payload, err := yaml.Marshal(diff.Before)
			Expect(err).NotTo(HaveOccurred())
			fmt.Println(string(payload))

			// writeBytes := []byte{}
			// buffer := bytes.NewBuffer(writeBytes)
			// readBytes := []byte{}
			// _, err = buffer.Read(readBytes)
			// Expect(err).NotTo(HaveOccurred())
			// Expect(readBytes).To(Equal("foo"))
			buffer := gbytes.NewBuffer()
			diff.Render(buffer, "someLabel")
			Expect(buffer).To(gbytes.Say("name : <redacted>"))
			Expect(true).To(BeTrue())
		})
	})
})
