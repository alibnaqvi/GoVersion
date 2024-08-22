package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	// Define the paths for vcs directory and necessary files
	vcsDir := "./vcs"
	commitsDir := filepath.Join(vcsDir, "commits")
	configFilePath := filepath.Join(vcsDir, "config.txt")
	indexFilePath := filepath.Join(vcsDir, "index.txt")
	logFilePath := filepath.Join(vcsDir, "log.txt")

	// Ensure vcs directory and necessary files exist
	createDirIfNotExists(vcsDir)
	createDirIfNotExists(commitsDir)
	createFileIfNotExists(configFilePath)
	createFileIfNotExists(indexFilePath)
	createFileIfNotExists(logFilePath)

	// Define a map of available commands and their descriptions
	commands := map[string]string{
		"config":   "Get and set a username.",
		"add":      "Add a file to the index.",
		"log":      "Show commit logs.",
		"commit":   "Save changes.",
		"checkout": "Restore a file.",
	}

	// Define the help message
	helpMessage := `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`

	// Check the number of arguments
	if len(os.Args) < 2 {
		// If no argument is provided, print the help message
		fmt.Println(helpMessage)
		return
	}

	// Get the command argument
	command := os.Args[1]

	// Handle the --help flag
	if command == "--help" {
		fmt.Println(helpMessage)
		return
	}

	switch command {
	case "config":
		handleConfig(configFilePath)
	case "add":
		handleAdd(indexFilePath)
	case "commit":
		handleCommit(indexFilePath, configFilePath, logFilePath, commitsDir)
	case "log":
		handleLog(logFilePath)
	case "checkout":
		handleCheckout(indexFilePath, commitsDir)
	default:
		if _, exists := commands[command]; !exists {
			fmt.Printf("'%s' is not a SVCS command.\n", command)
		} else {
			fmt.Println(commands[command])
		}
	}
}

// Ensure the directory exists by creating it if it doesn't exist
func createDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, os.ModePerm)
	}
	return nil
}

// Ensure the file exists by creating an empty file if it doesn't exist
func createFileIfNotExists(filePath string) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		file, err := os.Create(filePath)
		if err != nil {
			return err
		}
		file.Close()
	}
	return nil
}

// Handle the config command
func handleConfig(configFilePath string) {
	if len(os.Args) == 2 {
		// No name provided, show the current username
		data, err := ioutil.ReadFile(configFilePath)
		if err != nil || len(data) == 0 {
			fmt.Println("Please, tell me who you are.")
		} else {
			fmt.Printf("The username is %s.\n", strings.TrimSpace(string(data)))
		}
	} else if len(os.Args) == 3 {
		// Name provided, save it to config.txt
		username := os.Args[2]
		err := ioutil.WriteFile(configFilePath, []byte(username), os.ModePerm)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("The username is %s.\n", username)
	}
}

// Handle the add command
func handleAdd(indexFilePath string) {
	if len(os.Args) == 2 {
		data, err := os.ReadFile(indexFilePath)
		if err != nil && !os.IsNotExist(err) {
			log.Println("Error reading index file:", err)
			return
		}
		if len(data) == 0 {
			fmt.Println("Add a file to the index.")
		} else {
			fmt.Println("Tracked files:")
			fmt.Println(strings.TrimSpace(string(data)))
		}
	} else if len(os.Args) == 3 {
		fileName := os.Args[2]
		if _, err := os.Stat(fileName); os.IsNotExist(err) {
			fmt.Printf("Can't find '%s'.\n", fileName)
			return
		}
		data, err := os.ReadFile(indexFilePath)
		if err != nil {
			log.Println("Error reading index file:", err)
			return
		}
		trackedFiles := strings.Split(strings.TrimSpace(string(data)), "\n")
		for _, trackedFile := range trackedFiles {
			if trackedFile == fileName {
				fmt.Printf("The file '%s' is already tracked.\n", fileName)
				return
			}
		}
		trackedFiles = append(trackedFiles, fileName)
		err = os.WriteFile(indexFilePath, []byte(strings.Join(trackedFiles, "\n")), os.ModePerm)
		if err != nil {
			log.Println("Error writing index file:", err)
			return
		}
		fmt.Printf("The file '%s' is tracked.\n", fileName)
	}
}

// Handle the commit command
func handleCommit(indexFilePath, configFilePath, logFilePath, commitsDir string) {
	if len(os.Args) < 3 {
		fmt.Println("Message was not passed.")
		return
	}
	message := os.Args[2]

	// Read the username from config.txt
	username, err := os.ReadFile(configFilePath)
	if err != nil || len(username) == 0 {
		fmt.Println("Please, tell me who you are.")
		return
	}

	// Read the tracked files from index.txt
	trackedFiles, err := os.ReadFile(indexFilePath)
	if err != nil || len(trackedFiles) == 0 {
		fmt.Println("No files added to the index.")
		return
	}
	files := strings.Split(strings.TrimSpace(string(trackedFiles)), "\n")

	// Compute the hash of the current state of the files
	hash := computeFilesHash(files)

	// Check if there are previous commits
	entries, err := os.ReadDir(commitsDir)
	if err != nil {
		log.Fatal(err)
	}

	// If there's a previous commit, check if the hash matches the last commit
	for _, entry := range entries {
		if entry.IsDir() {
			commitDir := filepath.Join(commitsDir, entry.Name())
			if hash == computeCommitHash(commitDir, files) {
				fmt.Println("Nothing to commit.")
				return
			}
		}
	}

	// Generate the new commit ID
	commitID := fmt.Sprintf("%x", hash)

	// Create the new commit directory
	newCommitDir := filepath.Join(commitsDir, commitID)
	err = os.Mkdir(newCommitDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	// Copy all tracked files to the new commit directory
	for _, file := range files {
		err = copyFile(file, filepath.Join(newCommitDir, file))
		if err != nil {
			log.Fatal(err)
		}
	}

	// Log the commit in log.txt
	logEntry := fmt.Sprintf("commit %s\nAuthor: %s\n%s\n\n", commitID, strings.TrimSpace(string(username)), message)
	err = appendToFile(logFilePath, logEntry)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Changes are committed.")
}

// Compute the hash of a set of files
func computeFilesHash(files []string) [32]byte {
	h := sha256.New()
	for _, file := range files {
		data, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatal(err)
		}
		h.Write(data)
	}
	return sha256.Sum256(h.Sum(nil))
}

// Compute the hash of the files in a commit directory
func computeCommitHash(commitDir string, files []string) [32]byte {
	h := sha256.New()
	for _, file := range files {
		data, err := os.ReadFile(filepath.Join(commitDir, file))
		if err != nil {
			log.Fatal(err)
		}
		h.Write(data)
	}
	return sha256.Sum256(h.Sum(nil))
}

// Copy a file from source to destination
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	return err
}

// Append content to a file
func appendToFile(filePath, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(content)
	return err
}

// Handle the log command
func handleLog(logFilePath string) {
	data, err := os.ReadFile(logFilePath)
	if err != nil || len(data) == 0 {
		fmt.Println("No commits yet.")
		return
	}
	logEntries := strings.Split(string(data), "\n\n")
	for _, entry := range logEntries {
		if entry != "" {
			fmt.Println("-----")
			fmt.Println(entry)
		}
	}
}

// Handle the checkout command
func handleCheckout(indexFilePath, commitsDir string) {
	if len(os.Args) != 3 {
		fmt.Println("Commit id was not passed.")
		return
	}

	commitID := os.Args[2]
	commitDir := filepath.Join(commitsDir, commitID)

	// Check if the commit directory exists
	if _, err := os.Stat(commitDir); os.IsNotExist(err) {
		fmt.Println("Commit does not exist.")
		return
	}

	// Read the list of tracked files from index.txt
	trackedFiles, err := os.ReadFile(indexFilePath)
	if err != nil || len(trackedFiles) == 0 {
		fmt.Println("No files in the index.")
		return
	}

	files := strings.Split(strings.TrimSpace(string(trackedFiles)), "\n")

	// Restore the files from the commit directory
	for _, file := range files {
		src := filepath.Join(commitDir, file)
		dst := file

		// Check if the file exists in the commit directory
		if _, err := os.Stat(src); os.IsNotExist(err) {
			fmt.Printf("File '%s' not found in commit.\n", file)
			continue
		}

		err = copyFile(src, dst)
		if err != nil {
			log.Fatal(err)
		}
	}

	fmt.Printf("Switched to commit %s.\n", commitID)
}
