package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// CommitType represents the type of change in a structured commit message.
type CommitType string

const (
	CommitAdd    CommitType = "add"
	CommitUpdate CommitType = "update"
	CommitDelete CommitType = "delete"
	CommitRename CommitType = "rename"
)

// CommitScope represents the scope of a change.
type CommitScope string

const (
	ScopeRequest    CommitScope = "request"
	ScopeCollection CommitScope = "collection"
	ScopeEnv        CommitScope = "env"
)

// GitLogEntry represents a single commit in the collection history.
type GitLogEntry struct {
	Hash      string
	ShortHash string
	Message   string
	Author    string
	When      time.Time
}

// CollectionRepo wraps git operations for a single collection.
type CollectionRepo struct {
	Path string // path to the collection directory (which contains .git)
}

// getRepoPath returns the path to the collection's directory.
func collectionGitDir(name string) string {
	return filepath.Join(getCollectionsDir(), name)
}

// openRepo opens an existing git repository at the given path.
func openRepo(path string) (*git.Repository, error) {
	return git.PlainOpen(path)
}

// InitCollectionRepo initializes a new git repo for a collection.
// It creates a .gitignore that excludes sensitive files.
func InitCollectionRepo(name string) (*CollectionRepo, error) {
	path := collectionGitDir(name)
	if err := os.MkdirAll(path, 0o755); err != nil {
		return nil, fmt.Errorf("create collection dir: %w", err)
	}

	repo, err := git.PlainInit(path, false)
	if err != nil {
		return nil, fmt.Errorf("git init: %w", err)
	}

	// Create .gitignore to protect secrets
	gitignore := []byte("# Never commit sensitive values\n*.secret\n*.key\n.env.local\n")
	if err := os.WriteFile(filepath.Join(path, ".gitignore"), gitignore, 0o644); err != nil {
		return nil, fmt.Errorf("write .gitignore: %w", err)
	}

	// Initial commit
	w, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("worktree: %w", err)
	}

	if _, err := w.Add(".gitignore"); err != nil {
		return nil, fmt.Errorf("git add .gitignore: %w", err)
	}

	_, err = w.Commit("init(collection): initialize collection", &git.CommitOptions{
		Author: defaultSignature(),
	})
	if err != nil {
		return nil, fmt.Errorf("initial commit: %w", err)
	}

	return &CollectionRepo{Path: path}, nil
}

// OpenCollectionRepo opens an existing collection repo.
func OpenCollectionRepo(name string) (*CollectionRepo, error) {
	path := collectionGitDir(name)
	if _, err := os.Stat(filepath.Join(path, ".git")); os.IsNotExist(err) {
		return nil, fmt.Errorf("collection %q has no git repo", name)
	}
	return &CollectionRepo{Path: path}, nil
}

// SaveRequest stages and auto-commits changes to a request file.
func (cr *CollectionRepo) SaveRequest(req Request, commitType CommitType) error {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	// Write the request file
	if err := SaveRequest(req); err != nil {
		return fmt.Errorf("save request file: %w", err)
	}

	// Stage all changes
	if _, err := w.Add("."); err != nil {
		return fmt.Errorf("git add: %w", err)
	}

	// Check if there are staged changes
	status, err := w.Status()
	if err != nil {
		return fmt.Errorf("git status: %w", err)
	}
	if status.IsClean() {
		return nil // nothing to commit
	}

	msg := fmt.Sprintf("%s(request): %s %s", commitType, req.Method, req.Name)
	_, err = w.Commit(msg, &git.CommitOptions{
		Author: defaultSignature(),
	})
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	return nil
}

// DeleteRequest stages and commits the deletion of a request file.
func (cr *CollectionRepo) DeleteRequest(collection, requestID, requestName string) error {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	// Remove the file
	reqPath := filepath.Join("requests", requestID+".json")
	if _, err := w.Remove(reqPath); err != nil {
		return fmt.Errorf("git remove: %w", err)
	}

	msg := fmt.Sprintf("delete(request): remove %s", requestName)
	_, err = w.Commit(msg, &git.CommitOptions{
		Author: defaultSignature(),
	})
	if err != nil {
		return fmt.Errorf("git commit: %w", err)
	}

	// Also delete from filesystem
	return os.Remove(filepath.Join(cr.Path, reqPath))
}

// Log returns the commit history for this collection.
func (cr *CollectionRepo) Log(maxCount int) ([]GitLogEntry, error) {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return nil, fmt.Errorf("get HEAD: %w", err)
	}

	cIter, err := repo.Log(&git.LogOptions{
		From:  head.Hash(),
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("git log: %w", err)
	}

	var entries []GitLogEntry
	err = cIter.ForEach(func(c *object.Commit) error {
		if maxCount > 0 && len(entries) >= maxCount {
			return fmt.Errorf("limit reached") // stop iteration
		}
		entries = append(entries, GitLogEntry{
			Hash:      c.Hash.String(),
			ShortHash: c.Hash.String()[:7],
			Message:   c.Message,
			Author:    c.Author.Name,
			When:      c.Author.When,
		})
		return nil
	})
	if err != nil && err.Error() != "limit reached" {
		return nil, err
	}

	return entries, nil
}

// Diff returns a unified diff string for a file between two commits.
// If fromHash is empty, shows the full file as added.
func (cr *CollectionRepo) Diff(filePath, fromHash, toHash string) (string, error) {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return "", fmt.Errorf("open repo: %w", err)
	}

	getFileContent := func(hash, path string) (string, error) {
		commit, err := repo.CommitObject(plumbing.NewHash(hash))
		if err != nil {
			return "", err
		}
		tree, err := commit.Tree()
		if err != nil {
			return "", err
		}
		f, err := tree.File(path)
		if err != nil {
			return "", err // file doesn't exist in this version
		}
		return f.Contents()
	}

	var oldContent, newContent string

	if fromHash != "" {
		oldContent, err = getFileContent(fromHash, filePath)
		if err != nil {
			return "", fmt.Errorf("read old file: %w", err)
		}
	}

	if toHash != "" {
		newContent, err = getFileContent(toHash, filePath)
		if err != nil {
			return "", fmt.Errorf("read new file: %w", err)
		}
	} else {
		head, err := repo.Head()
		if err != nil {
			return "", fmt.Errorf("get HEAD: %w", err)
		}
		newContent, err = getFileContent(head.Hash().String(), filePath)
		if err != nil {
			return "", fmt.Errorf("read HEAD file: %w", err)
		}
	}

	return buildUnifiedDiff(filePath, oldContent, newContent), nil
}

// DiffFile returns the content of a file at a specific commit.
func (cr *CollectionRepo) DiffFile(filePath, commitHash string) (string, error) {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return "", fmt.Errorf("open repo: %w", err)
	}

	hash := plumbing.NewHash(commitHash)
	commit, err := repo.CommitObject(hash)
	if err != nil {
		return "", fmt.Errorf("commit: %w", err)
	}

	tree, err := commit.Tree()
	if err != nil {
		return "", fmt.Errorf("tree: %w", err)
	}

	file, err := tree.File(filePath)
	if err != nil {
		return "", fmt.Errorf("file: %w", err)
	}

	contents, err := file.Contents()
	if err != nil {
		return "", fmt.Errorf("contents: %w", err)
	}

	return contents, nil
}

// Checkout reverts a file to its state at a given commit and writes it to disk.
func (cr *CollectionRepo) Checkout(filePath, commitHash string) error {
	content, err := cr.DiffFile(filePath, commitHash)
	if err != nil {
		return fmt.Errorf("read file at commit: %w", err)
	}

	absPath := filepath.Join(cr.Path, filePath)
	if err := os.MkdirAll(filepath.Dir(absPath), 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}

	if err := os.WriteFile(absPath, []byte(content), 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}

// buildUnifiedDiff produces a simple unified diff string.
func buildUnifiedDiff(path, old, new string) string {
	oldLines := strings.Split(old, "\n")
	newLines := strings.Split(new, "\n")

	var out strings.Builder
	fmt.Fprintf(&out, "--- a/%s\n", path)
	fmt.Fprintf(&out, "+++ b/%s\n", path)

	// Simple line-by-line diff
	maxLen := max(len(newLines), len(oldLines))

	for i := 0; i < maxLen; i++ {
		var oldLine, newLine string
		if i < len(oldLines) {
			oldLine = oldLines[i]
		}
		if i < len(newLines) {
			newLine = newLines[i]
		}

		if oldLine == newLine {
			fmt.Fprintf(&out, " %s\n", oldLine)
		} else {
			if oldLine != "" {
				fmt.Fprintf(&out, "-%s\n", oldLine)
			}
			if newLine != "" {
				fmt.Fprintf(&out, "+%s\n", newLine)
			}
		}
	}

	return out.String()
}

// Branches returns a list of branch names in the collection repo.
func (cr *CollectionRepo) Branches() ([]string, error) {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return nil, fmt.Errorf("open repo: %w", err)
	}

	branches, err := repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("list branches: %w", err)
	}

	var names []string
	err = branches.ForEach(func(ref *plumbing.Reference) error {
		name := strings.TrimPrefix(ref.Name().String(), "refs/heads/")
		names = append(names, name)
		return nil
	})
	if err != nil {
		return nil, err
	}

	return names, nil
}

// CreateBranch creates a new branch from HEAD.
func (cr *CollectionRepo) CreateBranch(name string) error {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return fmt.Errorf("get HEAD: %w", err)
	}

	ref := plumbing.NewHashReference(
		plumbing.NewBranchReferenceName(name),
		head.Hash(),
	)

	if err := repo.Storer.SetReference(ref); err != nil {
		return fmt.Errorf("create branch: %w", err)
	}

	return nil
}

// CheckoutBranch switches to an existing branch.
func (cr *CollectionRepo) CheckoutBranch(name string) error {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return fmt.Errorf("open repo: %w", err)
	}

	w, err := repo.Worktree()
	if err != nil {
		return fmt.Errorf("worktree: %w", err)
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: plumbing.NewBranchReferenceName(name),
	})
	if err != nil {
		return fmt.Errorf("checkout branch: %w", err)
	}

	return nil
}

// CurrentBranch returns the name of the currently checked out branch.
func (cr *CollectionRepo) CurrentBranch() (string, error) {
	repo, err := openRepo(cr.Path)
	if err != nil {
		return "", fmt.Errorf("open repo: %w", err)
	}

	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("get HEAD: %w", err)
	}

	name := head.Name().String()
	if head.Name().IsBranch() {
		return strings.TrimPrefix(name, "refs/heads/"), nil
	}

	return name, nil
}

// defaultSignature returns the standard author identity for commits.
func defaultSignature() *object.Signature {
	return &object.Signature{
		Name:  "hilo",
		Email: "hilo@localhost",
		When:  time.Now(),
	}
}
