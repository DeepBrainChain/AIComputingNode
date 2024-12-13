package selfupdate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync/atomic"
	"time"

	"AIComputingNode/pkg/log"
	"AIComputingNode/pkg/model"
)

func UpdateGithubLatestRelease(ctx context.Context, cur_version string, activeReqs *int32) {
	// 1. Detect github latest release
	glr, err := DetectLatestGithubRelease(ctx, 15*time.Second)
	if err != nil {
		log.Logger.Errorf("Failed to detect github latest release: %v", err)
		return
	}

	log.Logger.Infof("Getting latest release: %v", glr)
	log.Logger.Infof("Current version %v, github latest version %v, tag_name %v, prerelease %v, published at %v",
		cur_version, glr.Name, glr.TagName, glr.PreRelease, glr.PublishedAt)
	if cur_version == glr.TagName {
		log.Logger.Info("Already latest, no need to upgrade")
		return
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
		log.Logger.Errorf("github latest release is empty: %v", asset)
		return
	}
	log.Logger.Infof("Get latest asset: %v", asset)

	if checksum.Url == "" {
		log.Logger.Error("Url of github latest release's checksum is empty")
		return
	}
	log.Logger.Infof("Get checksum asset: %v", checksum)

	// 2. Get sha256sum value of asset from github latest release
	checksums, err := checksum.DownloadChecksums(ctx, 15*time.Second)
	if err != nil {
		log.Logger.Errorf("Failed to download checkoutsums of github latest release: %v", err)
		return
	}
	log.Logger.Infof("Get checksums of github latest release: %v", checksums)

	var originHash = ""
	for key, value := range checksums {
		// if strings.HasSuffix(value, os_arch_ext) {
		if value == asset.Name {
			originHash = key
			break
		}
	}
	if originHash == "" {
		log.Logger.Errorf("sha256 hash of %v from github is empty", asset.Name)
		return
	}
	log.Logger.Infof("sha256 hash of %v from github is %v", asset.Name, originHash)

	// 3. Download github latest release and check sha256 hash
	execPath, err := os.Executable()
	if err != nil {
		log.Logger.Errorf("Failed to get executable filepath: %v", err)
		return
	}
	log.Logger.Infof("Get executable filepath: %v", execPath)
	filePath := filepath.Join(filepath.Dir(execPath), asset.Name)
	log.Logger.Infof("Get the filepath where the executable file is saved: %v", filePath)

	hashsum, err := sha256sum(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := asset.DownloadRelease(ctx, 30*time.Minute, filePath); err != nil {
				log.Logger.Errorf("Failed to download github latest release: %v", err)
				return
			}
			log.Logger.Infof(
				"Download github latest release success from: %v, size: %v, save in: %v",
				asset.Url,
				asset.Size,
				filePath,
			)

			hashsum, err = sha256sum(filePath)
			if err != nil {
				log.Logger.Errorf("sha256sum %v error: %v", filePath, err)
				return
			}
		} else {
			log.Logger.Errorf("sha256sum %v error: %v", filePath, err)
			return
		}
	}
	if hashsum == "" {
		log.Logger.Error("Failed to calcute sha256 hash of download file")
		return
	}
	log.Logger.Infof("sha256sum %v -> %v", filePath, hashsum)

	if hashsum != originHash {
		log.Logger.Error("sha256 hash of download file not match")
		// delete file
		os.Remove(filePath)
		return
	}
	log.Logger.Info("Check sha256 hash of download file success")

	// 4. Automatic restart during idle time
	activeHttpReqs := atomic.LoadInt32(activeReqs)
	activeModelReqs := model.IdleCount()
	log.Logger.Infof(
		"Currently active http requests %v, model idle count %v",
		activeHttpReqs,
		activeModelReqs,
	)
	if activeHttpReqs != 0 || activeModelReqs != 0 {
		log.Logger.Warnf("Wait for idle time to automatically restart")
		return
	}
	// os.Rename(execPath, execPath+".bak")
	backupOld := filepath.Join(
		filepath.Dir(execPath),
		fmt.Sprintf("host_%v_%v.old", cur_version, os_arch_ext),
	)
	if err := os.Rename(execPath, backupOld); err != nil {
		log.Logger.Errorf("Failed to mv %v -> %v: %v", execPath, backupOld, err)
		return
	}
	if err := os.Rename(filePath, execPath); err != nil {
		log.Logger.Errorf("Failed to replace the program file using mv: %v", err)
		return
	}
	// os.Remove(execPath + ".bak")
	log.Logger.Info("Begin to restart program......")
	os.Exit(0)
}
