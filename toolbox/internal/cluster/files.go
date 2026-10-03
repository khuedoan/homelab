package cluster

import (
	"crypto/rand"
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

func readFile(files *sftp.Client, name string) (string, error) {
	file, err := files.Open(name)
	if err != nil {
		return "", err
	}
	defer file.Close()
	data, err := io.ReadAll(io.LimitReader(file, 1<<20))
	if err != nil {
		return "", err
	}
	if len(data) == 1<<20 {
		return "", fmt.Errorf("remote file %s exceeds size limit", name)
	}
	return string(data), nil
}

func publishToken(files *sftp.Client, destination, token string) error {
	stage := path.Join(path.Dir(destination), ".incoming-"+rand.Text())
	file, err := files.OpenFile(stage, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
	if err != nil {
		return err
	}
	defer files.Remove(stage)
	defer file.Close()
	if err := file.Chmod(0600); err != nil {
		return err
	}
	if _, err := io.WriteString(file, token); err != nil {
		return err
	}
	if err := file.Sync(); err != nil {
		return err
	}
	if err := file.Close(); err != nil {
		return err
	}
	if err := files.Link(stage, destination); err != nil {
		return err
	}
	dir, err := files.Open(path.Dir(destination))
	if err != nil {
		return err
	}
	defer dir.Close()
	return dir.Sync()
}
