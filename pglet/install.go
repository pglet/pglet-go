package pglet

import (
	"archive/tar"
	"archive/zip"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	PgletVersion = "0.2.2"
)

var (
	pgletExe string
)

func install() error {
	pgletExe = "pglet"
	if runtime.GOOS == "windows" {
		pgletExe = "pglet.exe"
	}

	// check if pglet.exe is in PATH already  (development mode)
	path, err := exec.LookPath(pgletExe)
	if err == nil {
		pgletExe = path
		return nil
	}

	// create temp profile dir
	homeDir, _ := os.UserHomeDir()
	pgletDir := filepath.Join(homeDir, ".pglet")
	pgletBin := filepath.Join(pgletDir, "bin")

	_, err = os.Stat(pgletBin)
	if os.IsNotExist(err) {
		os.MkdirAll(pgletBin, os.ModePerm)
	}

	pgletExe = filepath.Join(pgletBin, pgletExe)

	ver := PgletVersion
	var installedVer string

	_, err = os.Stat(pgletExe)
	if err == nil {
		// read installed version
		cmd := exec.Command(pgletExe, "--version")
		outBytes, err := cmd.CombinedOutput()
		if err != nil {
			return err
		}
		installedVer = string(outBytes)
	}

	fmt.Println("Required version:", ver)
	fmt.Println("Installed version:", installedVer)

	return nil
}

func downloadFile(filepath string, url string) error {

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Create the file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return err
}

func untar(tarball, target string) error {
	reader, err := os.Open(tarball)
	if err != nil {
		return err
	}
	defer reader.Close()
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		path := filepath.Join(target, header.Name)
		info := header.FileInfo()
		if info.IsDir() {
			if err = os.MkdirAll(path, info.Mode()); err != nil {
				return err
			}
			continue
		}

		file, err := os.OpenFile(path, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, info.Mode())
		if err != nil {
			return err
		}
		defer file.Close()
		_, err = io.Copy(file, tarReader)
		if err != nil {
			return err
		}
	}
	return nil
}

func unzip(src string, dest string) ([]string, error) {

	var filenames []string

	r, err := zip.OpenReader(src)
	if err != nil {
		return filenames, err
	}
	defer r.Close()

	for _, f := range r.File {

		// Store filename/path for returning and using later on
		fpath := filepath.Join(dest, f.Name)

		// Check for ZipSlip. More Info: http://bit.ly/2MsjAWE
		if !strings.HasPrefix(fpath, filepath.Clean(dest)+string(os.PathSeparator)) {
			return filenames, fmt.Errorf("%s: illegal file path", fpath)
		}

		filenames = append(filenames, fpath)

		if f.FileInfo().IsDir() {
			// Make Folder
			os.MkdirAll(fpath, os.ModePerm)
			continue
		}

		// Make File
		if err = os.MkdirAll(filepath.Dir(fpath), os.ModePerm); err != nil {
			return filenames, err
		}

		outFile, err := os.OpenFile(fpath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			return filenames, err
		}

		rc, err := f.Open()
		if err != nil {
			return filenames, err
		}

		_, err = io.Copy(outFile, rc)

		// Close the file without defer to close before next iteration of loop
		outFile.Close()
		rc.Close()

		if err != nil {
			return filenames, err
		}
	}
	return filenames, nil
}
