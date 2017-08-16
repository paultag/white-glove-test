package main

import (
	"fmt"
	"net/mail"
	"strings"

	"pault.ag/go/archive"
	"pault.ag/go/debian/dependency"
	"pault.ag/go/white-glove-test/repo"
)

type Node struct {
	Package    archive.Package
	BuiltUsing dependency.Possibility
	Distance   int
	Arch       string
}

func main() {
	r := repo.Repo{Base: "http://mirror.cc.columbia.edu/debian/"}

	sources, err := r.LoadSourceMap("unstable", "main")
	if err != nil {
		panic(err)
	}

	outdated := map[string][]Node{}

	for arch, binaries := range *bmap {
		for _, binary := range binaries {
			latest := binary[0]
			who, err := mail.ParseAddress(latest.Maintainer)
			if err != nil {
				// Some people use Foo (Bar) <baz>, and the parser chokes on
				// (Bar), since I don't think that's actually valid
				continue
			}

			if who.Address != "pkg-go-maintainers@lists.alioth.debian.org" {
				continue
			}

			if len(latest.BuiltUsing.Relations) == 0 {
				continue
			}

			for _, possi := range latest.BuiltUsing.GetAllPossibilities() {
				distance, err := sources.Matches(possi)
				if err != nil {
					panic(err)
				}

				if distance == 0 {
					continue
				}

				outdated[latest.Package] = append(outdated[latest.Package], Node{
					Package:    latest,
					BuiltUsing: possi,
					Distance:   distance,
					Arch:       arch,
				})
			}
		}
	}

	for pkg, results := range outdated {
		arches := map[string][]string{}

		fmt.Printf("# %s\n", pkg)
		for _, result := range results {
			arches[result.Arch] = append(arches[result.Arch], result.BuiltUsing.Name)
		}

		a := []string{}
		for arch, why := range arches {
			chunks := strings.Split(arch, "-")
			a = append(a, chunks[1])
			fmt.Printf("# out of date on %s: %s\n", arch, strings.Join(why, ", "))
		}
		fmt.Printf("nmu %s . %s . unstable . -m \"Out of date Built-Using\"\n\n", pkg, strings.Join(a, " "))
	}
}
