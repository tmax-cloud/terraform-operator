package util

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/prometheus/common/log"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	restclient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

// ExecPodCmd exec command on specific pod and wait the command's output.
func ExecPodCmd(client kubernetes.Interface, config *restclient.Config, podName string, podNamespace string,
	command string, stdin io.Reader, stdout io.Writer, stderr io.Writer) error {

	cmd := []string{
		//"sh",
		"/bin/sh",
		"-c",
		command,
	}
	req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
		Namespace(podNamespace).SubResource("exec")
	//req := client.CoreV1().RESTClient().Post().Resource("pods").Name(podName).
	//	Namespace("default").SubResource("exec")

	option := &v1.PodExecOptions{
		Command: cmd,
		Stdin:   true,
		Stdout:  true,
		Stderr:  true,
		TTY:     true,
	}
	if stdin == nil {
		option.Stdin = false
	}
	req.VersionedParams(
		option,
		scheme.ParameterCodec,
	)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		return err
	}

	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  stdin,
		Stdout: stdout,
		Stderr: stderr,
	})
	if err != nil {
		return err
	}
	return nil
}

// ReadIDFromFile returns a Cloud Resource ID from Terraform State File
func ReadIDFromFile(filename string) (string, error) {
	var matched string // line with id
	var id string

	input, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Error(err, "Failed to read Terraform State File")
		return "", err
	}

	lines := strings.Split(string(input), "\n")

	for i, line := range lines {
		if strings.Contains(line, "\"id\"") {
			matched = lines[i]
		}
	}

	t := strings.Split(matched, "\"")

	if len(t) >= 4 {
		id = t[3]
	} else {
		err = errors.New("Index out of range")
		return "", err
	}

	return id, nil
}

// Initialize Terraform Working Directory
func InitTerraform_CLI(targetDir string, cloudType string) error {
	// Download Terraform Plugin (e.g. AWS, Azure, GCP)
	//cmd1 := exec.Command("bash", "-c", "terraform init")
	//cmd1.Dir = targetDir
	//stdoutStderr1, err1 := cmd1.CombinedOutput()
	//fmt.Printf("%s\n", stdoutStderr1)

	// Select the Terraform Plugin (cloudType: AWS, Azure, GCP)
	orgDir := HCL_DIR + "/" + ".terraform" + cloudType
	dstDir := targetDir + "/" + ".terraform"

	// Make the Destination Directory for plugin
	if _, err := os.Stat(dstDir); os.IsNotExist(err) {
		err = os.Mkdir(dstDir, 0755)
		if err != nil {
			return err
		}
		// Copy the Terraform Plugin (e.g. AWS, Azure, GCP) at Woring Directory
		err = copy(orgDir, dstDir)
		if err != nil {
			return err
		}
	}
	return nil
}

// Execute Terraform (Apply / Destroy)
func ExecuteTerraform_CLI(targetDir string, isDestroy bool) error {

	// Provision the Resources by Terraform
	cmd := exec.Command("bash", "-c", "terraform apply -auto-approve")

	// Swith the command from "apply" to "destroy"
	if isDestroy {
		// Destroy the Reosource by Terraform
		cmd = exec.Command("bash", "-c", "terraform destroy -auto-approve")
	}

	cmd.Dir = targetDir
	stdoutStderr, err := cmd.CombinedOutput()

	fmt.Printf("%s\n", stdoutStderr)

	return err
}

// Copy a Dierectory (preserve directory structure)
func copy(source, destination string) error {
	var err error = filepath.Walk(source, func(path string, info os.FileInfo, err error) error {
		var relPath string = strings.Replace(path, source, "", 1)
		if relPath == "" {
			return nil
		}
		if info.IsDir() {
			return os.Mkdir(filepath.Join(destination, relPath), 0755)
		} else {
			var data, err1 = ioutil.ReadFile(filepath.Join(source, relPath))
			if err1 != nil {
				return err1
			}
			return ioutil.WriteFile(filepath.Join(destination, relPath), data, 0777)
		}
	})
	return err
}
