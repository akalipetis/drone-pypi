package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

// Params desribes how to upload a Python module to PyPI.
type Params struct {
	Distributions []string `json:"distributions"`
	Password      *string  `json:"password,omitempty"`
	Repository    *string  `json:"repository,omitempty"`
	Username      *string  `json:"username,omitempty"`
	UploadPath    *string  `json:"upload_path,omitempty"`
}

func main() {
	// Parse parameters from environment
	repository := os.Getenv("PLUGIN_REPOSITORY")
	username := os.Getenv("PLUGIN_USERNAME")
	password := os.Getenv("PLUGIN_PASSWORD")
	uploadPath := os.Getenv("PLUGIN_UPLOAD_PATH")
	var distributions []string
	if dString, dExists := os.LookupEnv("PLUGIN_DISTRIBUTIONS"); dExists == true {
		distributions = strings.Split(dString, ",")
	} else {
		distributions = []string{}
	}
	v := Params{
		Repository:    &repository,
		Username:      &username,
		Password:      &password,
		Distributions: distributions,
		UploadPath:    &uploadPath,
	}

	err := v.Deploy()
	if err != nil {
		log.Fatal(err)
	}
}

// Deploy creates a PyPI configuration file and uploads a module.
func (v *Params) Deploy() error {
	err := v.CreateConfig()
	if err != nil {
		return err
	}
	err = v.UploadDist()
	if err != nil {
		return err
	}
	return nil
}

// CreateConfig creates a PyPI configuration file in the home directory of
// the current user.
func (v *Params) CreateConfig() error {
	f, err := os.Create(path.Join(os.Getenv("HOME"), ".pypirc"))
	if err != nil {
		return err
	}
	defer f.Close()
	buf := bufio.NewWriter(f)
	err = v.WriteConfig(buf)
	if err != nil {
		return err
	}
	buf.Flush()
	return nil
}

// UploadDist executes a distutils command to upload a python module.
func (v *Params) UploadDist() error {
	cmd := v.Upload()
	cmd.Dir = path.Join(os.Getenv("PWD"), *v.UploadPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("$", strings.Join(cmd.Args, " "))
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}

// WriteConfig writes a .pypirc to a supplied io.Writer.
func (v *Params) WriteConfig(w io.Writer) error {
	repository := "https://pypi.python.org/pypi"
	if v.Repository != nil {
		repository = *v.Repository
	}
	username := "guido"
	if v.Username != nil {
		username = *v.Username
	}
	password := "secret"
	if v.Password != nil {
		password = *v.Password
	}
	_, err := io.WriteString(w, fmt.Sprintf(`[distutils]
index-servers =
    pypi

[pypi]
repository: %s
username: %s
password: %s
`, repository, username, password))
	return err
}

// Upload creates a distutils upload command.
func (v *Params) Upload() *exec.Cmd {
	distributions := []string{"sdist"}
	if len(v.Distributions) > 0 {
		distributions = v.Distributions
	}
	args := []string{"setup.py"}
	for i := range distributions {
		args = append(args, distributions[i])
	}
	args = append(args, "upload")
	args = append(args, "-r")
	args = append(args, "pypi")
	return exec.Command("python", args...)
}
