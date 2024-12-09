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
	vcsDir := "./vcs"
	commitsDir := filepath.Join(vcsDir, "commits")
	configFilePath := filepath.Join(vcsDir, "config.txt")
	indexFilePath := filepath.Join(vcsDir, "index.txt")
	logFilePath := filepath.Join(vcsDir, "log.txt")

	createDirIfNotExists(vcsDir)
	createDirIfNotExists(commitsDir)
	createFileIfNotExists(configFilePath)
	createFileIfNotExists(indexFilePath)
	createFileIfNotExists(logFilePath)

	commands := map[string]string{
		"config":   "Get and set a username.",
		"add":      "Add a file to the index.",
		"log":      "Show commit logs.",
		"commit":   "Save changes.",
		"checkout": "Restore a file.",
	}

	helpMessage := `These are SVCS commands:
config     Get and set a username.
add        Add a file to the index.
log        Show commit logs.
commit     Save changes.
checkout   Restore a file.`

	if len(os.Args) < 2 {
		fmt.Println(helpMessage)

		return
	}

	command := os.Args[1]

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

func createDirIfNotExists(dirPath string) error {
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		return os.MkdirAll(dirPath, os.ModePerm)
	}

	return nil
}

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

func handleConfig(configFilePath string) {
	if len(os.Args) == 2 {
		data, err := ioutil.ReadFile(configFilePath)

		if err != nil || len(data) == 0 {
			fmt.Println("Please, tell me who you are.")
		} else {
			fmt.Printf("The username is %s.\n", strings.TrimSpace(string(data)))
		}
	} else if len(os.Args) == 3 {
		username := os.Args[2]
		err := ioutil.WriteFile(configFilePath, []byte(username), os.ModePerm)

		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("The username is %s.\n", username)
	}
}

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

func handleCommit(indexFilePath, configFilePath, logFilePath, commitsDir string) {
	if len(os.Args) < 3 {
		fmt.Println("Message was not passed.")

		return
	}

	message := os.Args[2]

	username, err := os.ReadFile(configFilePath)

	if err != nil || len(username) == 0 {
		fmt.Println("Please, tell me who you are.")

		return
	}

	trackedFiles, err := os.ReadFile(indexFilePath)

	if err != nil || len(trackedFiles) == 0 {
		fmt.Println("No files added to the index.")

		return
	}

	files := strings.Split(strings.TrimSpace(string(trackedFiles)), "\n")

	hash := computeFilesHash(files)

	entries, err := os.ReadDir(commitsDir)

	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			commitDir := filepath.Join(commitsDir, entry.Name())

			if hash == computeCommitHash(commitDir, files) {
				fmt.Println("Nothing to commit.")

				return
			}
		}
	}

	commitID := fmt.Sprintf("%x", hash)

	newCommitDir := filepath.Join(commitsDir, commitID)
	err = os.Mkdir(newCommitDir, os.ModePerm)

	if err != nil {
		log.Fatal(err)
	}

	for _, file := range files {
		err = copyFile(file, filepath.Join(newCommitDir, file))

		if err != nil {
			log.Fatal(err)
		}
	}

	logEntry := fmt.Sprintf("commit %s\nAuthor: %s\n%s\n\n", commitID, strings.TrimSpace(string(username)), message)
	err = appendToFile(logFilePath, logEntry)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Changes are committed.")
}

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

func appendToFile(filePath, content string) error {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_WRONLY, os.ModeAppend)

	if err != nil {
		return err
	}

	defer f.Close()

	_, err = f.WriteString(content)

	return err
}

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

func handleCheckout(indexFilePath, commitsDir string) {
	if len(os.Args) != 3 {
		fmt.Println("Commit id was not passed.")

		return
	}

	commitID := os.Args[2]
	commitDir := filepath.Join(commitsDir, commitID)

	if _, err := os.Stat(commitDir); os.IsNotExist(err) {
		fmt.Println("Commit does not exist.")

		return
	}

	trackedFiles, err := os.ReadFile(indexFilePath)

	if err != nil || len(trackedFiles) == 0 {
		fmt.Println("No files in the index.")

		return
	}

	files := strings.Split(strings.TrimSpace(string(trackedFiles)), "\n")

	for _, file := range files {
		src := filepath.Join(commitDir, file)
		dst := file

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
