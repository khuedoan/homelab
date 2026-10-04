package cluster

import (
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"regexp"

	"github.com/pkg/sftp"
)

const tokenDir = "/var/lib/rancher/k3s/enrollment"

const tokenPath = tokenDir + "/token"

var serverTokenFormat = regexp.MustCompile(`^K10[0-9a-f]{64}::server:[^\s]+$`)

func validToken(token string) bool {
	return len(token) <= 4096 && serverTokenFormat.MatchString(token)
}

func exists(files *sftp.Client, name string) (bool, error) {
	_, err := files.Lstat(name)
	if os.IsNotExist(err) {
		return false, nil
	}
	return err == nil, err
}

func securePath(files *sftp.Client, name string, mode os.FileMode, directory bool) error {
	info, err := files.Lstat(name)
	if err != nil {
		return err
	}
	owner, ok := info.Sys().(*sftp.FileStat)
	if !ok || owner.UID != 0 || owner.GID != 0 || info.Mode().Perm() != mode || info.IsDir() != directory || !directory && !info.Mode().IsRegular() {
		return fmt.Errorf("unsafe type, ownership or permissions on %s", name)
	}
	return nil
}

func readFile(files *sftp.Client, name string) (data string, err error) {
	file, err := files.Open(name)
	if err != nil {
		return "", err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	contents, err := io.ReadAll(io.LimitReader(file, 1<<20))
	if err != nil {
		return "", err
	}
	if len(contents) == 1<<20 {
		return "", fmt.Errorf("remote file %s exceeds size limit", name)
	}
	return string(contents), nil
}

func publishToken(files *sftp.Client, destination, token string) (err error) {
	stage := path.Join(path.Dir(destination), ".incoming-"+rand.Text())
	file, err := files.OpenFile(stage, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, files.Remove(stage)) }()
	closed := false
	defer func() {
		if !closed {
			err = errors.Join(err, file.Close())
		}
	}()
	if err := file.Chmod(0600); err != nil {
		return err
	}
	if _, err := io.WriteString(file, token); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	err = file.Close()
	closed = true
	if err != nil {
		return err
	}
	if err := files.Link(stage, destination); err != nil {
		return err
	}
	dir, err := files.Open(path.Dir(destination))
	if err != nil {
		return err
	}
	defer func() { err = errors.Join(err, dir.Close()) }()
	return dir.Sync()
}
