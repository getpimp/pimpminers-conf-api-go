// Package pimpminersconf is the API wrapper for interacting with the pimpminers.conf json library.
/*==================================================================
       .__
______ |__| _____ ______  Portable Instant Mining Platform
\____ \|  |/     \____  \       By miners, for miners.
|  |_> >  |  Y Y  \  |_> >
|   __/|__|__|_|  /   __/    Support: forum.getpimp.org
|__|            \/|__|
Copyright (c) 2019 getPiMP.org.  All Rights Reserved.
License: This code is licensed for use with PiMP only.
Description: PiMP OS pimpminers.conf API wrapper in Golang
Interacts with this file: https://github.com/getpimp/pimpminers-conf/pimpminers.conf
==================================================================*/
package pimpminersconf

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"

	"gopkg.in/src-d/go-git.v4"
	"gopkg.in/src-d/go-git.v4/plumbing/object"
)

const remote = "https://raw.githubusercontent.com/getpimp/pimpminers-conf/master/pimpminers.conf"
const stagingFile = "/tmp/pimpminers.conf"
const localGitRepo = "/tmp/pimpminers-conf"
const pimp2repo = "https://update.getpimp.org/pimpup/miners/"

// PimpMiner is the golang representation of the pimpminers.conf json library.
type PimpMiner struct {
	Info           string              `json:"info"`
	Platform       string              `json:"platform"`
	Repotype       string              `json:"repotype"`
	Repo           string              `json:"repo"`
	Folder         string              `json:"folder"`
	Binary         string              `json:"binary"`
	Configure      string              `json:"configure"`
	Menu           string              `json:"menu"`
	Postexec       string              `json:"postexec"`
	Profiles       []pimpMinerProfile  `json:"profiles"`
	MainVersion    string              `json:"MainVersion"`
	DevelVersion   string              `json:"DevelVersion"`
	PageURL        string              `json:"PageURL"`
	PageURLRegex   string              `json:"PageURLRegex"`
	SupportedAlgos []map[string]string `json:"SupportedAlgos"`
	BtcTalk        string              `json:"BtcTalk"`
}

type pimpMinerProfile struct {
	ID     string                   `json:"id"`
	Name   string                   `json:"name"`
	Cfile  string                   `json:"cfile"`
	Config []pimpMinerProfileConfig `json:"Config"`
}

type pimpMinerProfileConfig struct {
	Flags      string `json:"FLAGS"`
	CONF       string `json:"CONF"`
	API        string `json:"API"`
	POOL_TITLE string `json:"POOL_TITLE"`
	TYPE       string `json:"TYPE"`
	Extra      string `json:"EXTRA"`
	Notes      string `json:"NOTES"`
	Template   string `json:"TEMPLATE"`
}

// Load returns a mapstring of PimpMiners populated with data from the pimpminers.conf file
func Load(file string) map[string][]PimpMiner {
	// if no file specified, default to /tmp/pimpminers.conf
	if file == "" {
		file = stagingFile
	}

	// download the file
	if err := DownloadFile(file, remote); err != nil {
		fmt.Println("ERROR downloading the file.")
		panic(err)
	}

	jsonFile, err := os.Open(file) // Open the JSON file
	if err != nil {                // if os.Open returns an error then handle it
		fmt.Println("ERROR opening the file.")
		panic(err)
	}
	defer jsonFile.Close()                   // defer the closing of our jsonFile so that we can parse it later on
	byteValue, _ := ioutil.ReadAll(jsonFile) // read our opened json file as a byte array.
	var miners map[string][]PimpMiner        // Create stringmap of our slice of structs
	json.Unmarshal(byteValue, &miners)       // Read in the JSON
	return miners
}

// GetMiner returns a PimpMiner object with the specified name from the provided mapstring.
// Note: This is for reading, not setting values.
func GetMiner(name string, miners map[string][]PimpMiner) PimpMiner {
	for _, v := range miners {
		if v[0].Info == name {
			return v[0]
		}
	}
	return PimpMiner{}
}

// SetRepo updates a PimpMiner object's repo property with the specified filename.
// Note: This is for pimpup 2.x.
func SetRepo(name string, repo string, miners map[string][]PimpMiner) string {
	out := ""
	for _, v := range miners {
		if v[0].Info == name {
			out = pimp2repo + repo + ".tgz" // for output / return value
			v[0].Repo = out                 // update the object
		}
	}
	return out
}

// checkErr checks if there was an error, and if it does, prints it to the screen
func checkErr(err error) { // to keep code clean
	if err != nil {
		fmt.Println(err.Error()) // output the error
	}
}

// DownloadFile will download a url to a stagingFile file.
func DownloadFile(filepath string, url string) error {
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

// RunCommand runs the specified string as a command in a bash shell, and returns its output as a string.
func RunCommand(cmd string) string {
	out, err := exec.Command("bash", "-c", cmd).Output()
	checkErr(err)
	return string(out)
}

// Save takes a map of strings to Miners object, marshals it into json, and then saves it as a file.
func Save(jsonfile string, m map[string][]PimpMiner) {
	json, err := json.Marshal(m)
	checkErr(err)
	f, err := os.Create(jsonfile)
	checkErr(err)
	defer f.Close()
	out := []byte(PrettyPrint(string(json)))
	f.Write(out)
}

// FileExists takes a filename string and returns it if it exists, or empty string if it does not.
func FileExists(file string) string {
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		// path/to/whatever exists
		return file
	}
	return ""
}

// PrettyPrint takes a string of json in and returns a prettier string of json out.
func PrettyPrint(in string) string {
	var out bytes.Buffer
	err := json.Indent(&out, []byte(in), "", "\t")
	if err != nil {
		return in
	}
	return out.String()
}

// Clone will clone the pimpminers-conf repo to /tmp/pimpminers-conf.
func Clone() *git.Repository {
	// backup existing dir and move out of the way.
	move := fmt.Sprintf("mv %s %s.old", localGitRepo, localGitRepo)
	RunCommand(move)
	r, err := git.PlainClone(localGitRepo, false, &git.CloneOptions{
		URL:      "https://github.com/getpimp/pimpminers-conf.git",
		Progress: os.Stdout,
	})
	checkErr(err)
	return r
}

// Commit will commit the pimpminers-conf repo to git. (Maintainers only.) Returns true if success.
func Commit(r *git.Repository, msg string) bool {
	// check for differences.
	diffCmd := fmt.Sprintf("diff -U 0 %s %s/pimpminers.conf | grep -v ^@ | wc -l", stagingFile, localGitRepo)
	diff := RunCommand(diffCmd)
	diff = strings.TrimSpace(diff)
	if diff == "0" {
		fmt.Println("No changes to commit.")
	} else {
		// copy file from staging into worktree
		diffCmd = fmt.Sprintf("diff -U 0 %s %s/pimpminers.conf | grep -v ^@", stagingFile, localGitRepo)
		diff = RunCommand(diffCmd)
		diff = strings.TrimSpace(diff)
		fmt.Println("Changes:")
		fmt.Println(diff)
		fmt.Println("\nCommitting changes... ")
		copy := fmt.Sprintf("cp %s %s/pimpminers.conf", stagingFile, localGitRepo)
		RunCommand(copy)
		w, err := r.Worktree()
		checkErr(err)
		// add files
		_, err = w.Add("pimpminers.conf")
		checkErr(err)
		// commit
		_, err = w.Commit(msg, &git.CommitOptions{
			Author: &object.Signature{
				Name:  "pimplabops",
				Email: "labops@getpimp.org",
				When:  time.Now(),
			},
		})
		checkErr(err)
		if err != nil {
			return false
		}
	}
	return true
}
