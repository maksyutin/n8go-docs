package utils

import (
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"n8go-docs/diagnostics"
)

func copyFile(src string, dst string) error {
	diagnostics.Debug(func() {
		log.Println("copying\n", src, "to", dst)
	})

	err := os.MkdirAll(filepath.Dir(dst), os.ModePerm)
	if err != nil {
		return err
	}

	inFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer inFile.Close()

	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, inFile)
	if err != nil {
		return err
	}

	return nil
}

func CopyDirContents(srcDir string, dstDir string, predicate func(ext string) bool) error {
	return CopyDirContentsWithHook(srcDir, dstDir, predicate, nil)
}

// CopyDirContentsWithHook copies files matching predicate from srcDir to dstDir.
// onCopy is called before each file copy; returning false skips the file.
func CopyDirContentsWithHook(srcDir string, dstDir string, predicate func(ext string) bool, onCopy func(relPath string) bool) error {
	srcDirAbs, err := filepath.Abs(srcDir)
	if err != nil {
		return err
	}

	dstDirAbs, err := filepath.Abs(dstDir)
	if err != nil {
		return err
	}

	return filepath.WalkDir(srcDirAbs, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !predicate(filepath.Ext(path)) {
			return nil
		}

		relativePath := path[len(srcDirAbs):]
		relSlash := filepath.ToSlash(relativePath[1:]) // strip leading separator

		if onCopy != nil && !onCopy(relSlash) {
			return nil
		}

		return copyFile(path, filepath.Join(dstDirAbs, relativePath))
	})
}
