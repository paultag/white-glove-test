package main

import (
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"

	"pault.ag/go/archive"
	"pault.ag/go/debian/dependency"
)

// Built-Using Relation between two Sources. This is provided for a source, for
// a binary package with an out of date relation. Package is the binary itself,
// and BuiltUsing is the source of the package that has been updated since the
// last compile of this package.
type Candidate struct {
	Package    archive.Package
	BuiltUsing dependency.Possibility
	Arch       dependency.Arch
	Distance   int
}

type Candidates []Candidate

func (c Candidates) Arches() []dependency.Arch {
	set := map[dependency.Arch]bool{}
	for _, candidate := range c {
		set[candidate.Arch] = true
	}

	ret := []dependency.Arch{}
	for arch, _ := range set {
		ret = append(ret, arch)
	}
	return ret
}

func (c Candidates) Sources() []string {
	set := map[string]bool{}
	for _, candidate := range c {
		set[candidate.BuiltUsing.Name] = true
	}

	ret := []string{}
	for arch, _ := range set {
		ret = append(ret, arch)
	}
	return ret
}

//
//
//
func LoadSources(reader io.Reader) (*archive.SourceMap, error) {
	packages, err := archive.LoadSources(reader)
	if err != nil {
		return nil, err
	}
	return archive.LoadSourceMap(*packages)
}

//
//
//
func LoadPackages(reader io.Reader) (*archive.PackageMap, error) {
	packages, err := archive.LoadPackages(reader)
	if err != nil {
		return nil, err
	}
	return archive.LoadPackageMap(*packages)
}

// source -> Candidate results
type CandidatesMap map[string]Candidates

// arch -> name -> binary
type ArchMap map[string]archive.PackageMap

func LoadSourcesFile(path string) (*archive.SourceMap, error) {
	fd, err := os.Open("Sources")
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return LoadSources(fd)
}

func LoadPackagesFile(path string) (*archive.PackageMap, error) {
	fd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fd.Close()
	return LoadPackages(fd)
}

func main() {
	amap := ArchMap{}
	arches := []string{"amd64", "armhf"}

	smap, err := LoadSourcesFile("Sources")
	if err != nil {
		panic(err)
	}

	for _, arch := range arches {
		bmap, err := LoadPackagesFile(fmt.Sprintf("Packages-%s", arch))
		if err != nil {
			panic(err)
		}
		amap[arch] = *bmap
	}

	cmap := CandidatesMap{}

	for _, packageMap := range amap {
		for name, packages := range packageMap {
			sname, candidates := ProcessBinary(*smap, packages)
			if len(candidates) == 0 {
				continue
			}
			cmap[sname] = append(cmap[name], candidates...)
		}
	}

	for pkg, candidates := range cmap {
		who, err := mail.ParseAddress(candidates[0].Package.Maintainer)
		if err != nil {
			continue
		}
		if who.Address != "pkg-go-maintainers@lists.alioth.debian.org" {
			continue
		}

		carches := candidates.Arches()
		if len(carches) == 1 && carches[0].Is(&dependency.All) {
			continue
		}

		// sources := candidates.Sources()

		if len(arches) == len(carches) {
			carches = []dependency.Arch{dependency.Any}
		}

		fmt.Printf("nmu %s . %s . -m '%s'\n", pkg.Package, join(carches, ", "), "out of date")
	}
}

type stringable interface {
	String() string
}

func join(s []dependency.Arch, sep string) string {
	ret := []string{}
	for _, el := range s {
		ret = append(ret, el.String())
	}
	return strings.Join(ret, sep)
}

func ProcessBinary(smap archive.SourceMap, packages []archive.Package) (string, Candidates) {
	ret := Candidates{}
	latest := packages[0]
	sname := latest.Source
	if sname == "" {
		sname = latest.Package
	}

	for _, possi := range latest.BuiltUsing.GetAllPossibilities() {
		depth, err := smap.Matches(possi)
		if err != nil {
			panic(err)
		}
		if depth <= 0 {
			continue
		}
		ret = append(ret, Candidate{
			Package:    latest,
			BuiltUsing: possi,
			Arch:       latest.Architecture,
			Distance:   depth,
		})
	}

	return sname, ret
}
