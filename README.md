# Simple Version Control System (SVCS)

A basic version control system implemented in Go. This system allows users to manage and track changes to files, commit changes, and switch between different versions of the files.

## Features

- **Configuration**: Set and get a username for commits.
- **Add**: Add files to the index to track changes.
- **Commit**: Save changes with a commit message.
- **Log**: View commit history.
- **Checkout**: Restore files to a specific commit state.

## Getting Started

### Prerequisites

- Go installed on your machine.
- An IDE installed. (GoLand prefered).

### Installation

1. Clone the repository

2. CD into the project directory (or open the directory from you IDE)

3. Build the project using:

```bash
    go build -o svcs
```

### Usage

#### Configuration

Set or get your username:

```bash
./svcs config [username]
```

To set a username:

```bash
./svcs config JohnDoe
```

To get the current username:

```bash
./svcs config
```

#### Adding Files

Add a file to the index:

```bash
./svcs add [filename]
```

Example:

```bash
./svcs add file1.txt
```

#### Committing Changes

Commit the changes with a message:

```bash
./svcs commit [commit-message]
```

Example:

```bash
./svcs commit "Initial commit"
```

#### Viewing Commit Logs

Show the commit history:

```bash
./svcs log
```

#### Checking Out Commits

Restore files to a specific commit:

```bash
./svcs checkout [commit-id]
```

Example:

```bash
./svcs checkout 0b4f05fcd3e1dcc47f58fed4bb189196f99da89a
```

### Directory Structure:

.
├── vcs
│   ├── commits
│   │   ├── [commit-id]
│   │   │   ├── [file1.txt]
│   │   │   └── [file2.txt]
│   │   └── ...
│   ├── config.txt
│   ├── index.txt
│   └── log.txt
├── file1.txt
├── file2.txt
└── untracked_file.txt

* commits/ directory contains commit snapshots.
* config.txt stores the username.
* index.txt lists tracked files.
* log.txt records commit messages and history.

## Notes

* Ensure you add files to the index before committing.
* The checkout command only restores files that are tracked in the index.

## Contributing

Feel free to open issues or submit pull requests. Contributions are welcome!

## Any Problems?

For more information or if you encounter any issues, please contact ali.b.naqvi@berkley.edu or open an issue on GitHub.
