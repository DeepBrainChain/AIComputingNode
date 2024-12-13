package selfupdate

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

// go test -v -timeout 360s -count=1 -run TestUpdateGithubLatestRelease AIComputingNode/pkg/selfupdate
func TestUpdateGithubLatestRelease(t *testing.T) {
	var cur_version = "v0.1.8"
	t.Log("Operating System:", runtime.GOOS)
	t.Log("Architecture:", runtime.GOARCH)
	ctx := context.Background()

	// 1. Detect github latest release
	glr, err := DetectLatestGithubRelease(ctx, 15*time.Second)
	if err != nil {
		t.Fatalf("Failed to detect github latest release: %v", err)
	}

	t.Logf("Getting latest release: %v", glr)
	t.Logf("Current version %v, github latest version %v, tag_name %v, prerelease %v, published at %v",
		cur_version, glr.Name, glr.TagName, glr.PreRelease, glr.PublishedAt)
	if cur_version == glr.TagName {
		t.Fatalf("Already latest, no need to upgrade")
	}

	asset := GithubReleaseAsset{}
	checksum := GithubReleaseAsset{}
	os_arch_ext := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		os_arch_ext += ".exe"
	}
	for _, ass := range glr.Assets {
		if strings.HasSuffix(ass.Name, os_arch_ext) {
			asset = ass
		}
		if ass.Name == "checksums.txt" {
			checksum = ass
		}
	}

	if asset.Url == "" || asset.Name == "" {
		t.Fatalf("github latest release is empty: %v", asset)
	}
	t.Log("Get latest asset:", asset)

	if checksum.Url == "" {
		t.Fatal("Url of github latest release's checksum is empty")
	}
	t.Logf("Get checksum asset: %v", checksum)

	// 2. Get sha256sum value of asset from github latest release
	checksums, err := checksum.DownloadChecksums(ctx, 15*time.Second)
	if err != nil {
		t.Fatalf("Failed to download checkoutsums of github latest release: %v", err)
	}
	t.Logf("Get checksums of github latest release: %v", checksums)

	var originHash = ""
	for key, value := range checksums {
		// if strings.HasSuffix(value, os_arch_ext) {
		if value == asset.Name {
			originHash = key
			break
		}
	}
	if originHash == "" {
		t.Fatalf("sha256 hash of %v from github is empty", asset.Name)
	}
	t.Logf("sha256 hash of %v from github is %v", asset.Name, originHash)

	// 3. Download github latest release and check sha256 hash
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get program directory: %v", err)
	}
	filePath := filepath.Join(cwd, asset.Name)
	t.Logf("Get the filepath where the executable file is saved: %v", filePath)

	hashsum, err := sha256sum(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := asset.DownloadRelease(ctx, 5*time.Minute, filePath); err != nil {
				t.Fatalf("Failed to download github latest release: %v", err)
			}
			t.Logf("Download github latest release success from: %v, size: %v, save in: %v", asset.Url, asset.Size, filePath)

			hashsum, err = sha256sum(filePath)
			if err != nil {
				t.Fatalf("sha256sum %v error: %v", filePath, err)
			}
		} else {
			t.Fatalf("sha256sum %v error: %v", filePath, err)
		}
	}
	if hashsum == "" {
		t.Fatal("Failed to calcute sha256 hash of download file")
	}
	t.Logf("sha256sum %v -> %v", filePath, hashsum)

	if hashsum != originHash {
		// delete file
		os.Remove(filePath)
		t.Fatal("sha256 hash of download file not match")
	}
	t.Log("Check sha256 hash of download file success")

	// 4. Automatic restart during idle time
}

// go test -v -timeout 30s -count=1 -run TestFilePath AIComputingNode/pkg/selfupdate
func TestFilePath(t *testing.T) {
	path1, err := os.Executable()
	if err != nil {
		t.Fatalf("os.Executable failed: %v", err)
	}
	t.Log("os.Executable:", path1)
	t.Log("filepath.Dir:", filepath.Dir(path1))
	t.Log("filepath.Base:", filepath.Base(path1))
	t.Log("filepath.Ext:", filepath.Ext(path1))

	path2, err := os.Getwd()
	if err != nil {
		t.Fatalf("os.Getwd: %v", err)
	}
	t.Log("os.Getwd:", path2)
}

// go test -v -timeout 30s -count=1 -run TestCalculateSHA256 AIComputingNode/pkg/selfupdate
func TestCalculateSHA256(t *testing.T) {
	if hashstr, err := sha256sum("github_releases.go"); err != nil {
		t.Fatalf("sha256sum failed: %v", err)
	} else {
		t.Logf("sha256sum: %v", hashstr)
	}

	emptyhash, err := sha256sum("not_existed_file")
	t.Log("sha256sum not_existed_file:", emptyhash, err)
	t.Log(os.IsNotExist(err))

	file, err := os.Open("github_releases.go")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Fatalf("Failed to copy file: %v", err)
	}

	hb := hash.Sum(nil)
	t.Logf("Calcute sha256: %x", hb)
	t.Logf("Calcute sha256: %v", hex.EncodeToString(hb))
	t.Logf("Calcute sha256: %v", hb)
}

func TestChecksumsTxt(t *testing.T) {
	file, err := os.Open("checksums.txt")
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) < 2 {
			t.Log("invalid line format of checksums")
			continue
		}
		// t.Logf("checksums line %v: %v", len(fields), fields)
		t.Logf("sha256 %v of %v", fields[0], fields[1])
	}

	if err := scanner.Err(); err != nil {
		t.Fatalf("line scanner error: %v", err)
	}
}
