package repo

import (
	"io"
	"log"
	"net/http"
	"path"

	"pault.ag/go/archive"
	"xi2.org/x/xz"
)

type Repo struct {
	Base string
}

func (r Repo) createURL(resource ...string) string {
	return r.Base + path.Join(resource...)
}

func (r Repo) getXZ(resource ...string) (io.Reader, func() error, error) {
	path := r.createURL(resource...)
	log.Println("Fetching path", path)

	resp, err := http.Get(path)
	if err != nil {
		return nil, nil, err
	}
	reader, err := xz.NewReader(resp.Body, 0)
	if err != nil {
		return nil, nil, err
	}
	return reader, resp.Body.Close, nil
}

func (r Repo) Packages(suite, component, arch string) (*archive.Packages, func() error, error) {
	reader, closer, err := r.getXZ("dists", suite, component, arch, "Packages.xz")
	if err != nil {
		return nil, closer, err
	}
	packages, err := archive.LoadPackages(reader)
	return packages, closer, err
}

func (r Repo) Sources(suite, component string) (*archive.Sources, func() error, error) {
	reader, closer, err := r.getXZ("dists", suite, component, "source", "Sources.xz")
	if err != nil {
		return nil, closer, err
	}
	packages, err := archive.LoadSources(reader)
	return packages, closer, err
}
