package main

import (
	"fmt"
	"net/mail"

	"pault.ag/go/white-glove-test/repo"
	"pault.ag/go/white-glove-test/untangle"
)

func main() {
	// r := repo.Repo{Base: "http://archive.paultag.house/debian/"}
	r := repo.Repo{Base: "http://proxy:3142/debian/"}

	sourcesR, closer, err := r.Sources("unstable", "main")
	if err != nil {
		panic(err)
	}
	defer closer()

	sources, err := untangle.LoadSourceMap(*sourcesR)
	if err != nil {
		panic(err)
	}

	packages, closer, err := r.Packages("unstable", "main", "binary-amd64")
	if err != nil {
		panic(err)
	}
	defer closer()

	binaries, err := untangle.LoadBinaryMap(*packages)
	if err != nil {
		panic(err)
	}

	for _, binary := range *binaries {
		latest := binary[0]
		for _, possi := range latest.BuiltUsing.GetAllPossibilities() {
			who, err := mail.ParseAddress(latest.Maintainer)
			if err != nil {
				// Some people use Foo (Bar) <baz>, and the parser chokes on
				// (Bar), since I don't think that's actually valid
				continue
			}

			if who.Address != "pkg-go-maintainers@lists.alioth.debian.org" {
				continue
			}

			distance, err := sources.Matches(possi)
			if err != nil {
				panic(err)
			}

			if distance == 0 {
				continue
			}

			fmt.Printf("%d %s -> %s\n", distance, latest.Package, possi.Name)
		}
	}
}
