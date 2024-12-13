package selfupdate

import (
	"bufio"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"AIComputingNode/pkg/log"
)

/*
https://docs.github.com/zh/rest/releases?apiVersion=2022-11-28

curl -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/OWNER/REPO/releases/latest

curl -L \
  -H "Accept: application/vnd.github+json" \
  -H "Authorization: Bearer <YOUR-TOKEN>" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/OWNER/REPO/releases/assets/ASSET_ID

*/

var getGithubLatestReleaseUrl = "https://api.github.com/repos/DeepBrainChain/AIComputingNode/releases/latest"

type GithubReleaseAsset struct {
	Url                string `json:"url"`
	Id                 int64  `json:"id"`
	NodeId             string `json:"node_id"`
	Name               string `json:"name"`
	Label              string `json:"label"`
	ContentType        string `json:"content_type"`
	State              string `json:"state"`
	Size               int64  `json:"size"`
	BrowserDownloadUrl string `json:"browser_download_url"`
}

type GithubLatestRelease struct {
	Url         string               `json:"url"`
	Id          int64                `json:"id"`
	NodeId      string               `json:"node_id"`
	TagName     string               `json:"tag_name"`
	Name        string               `json:"name"`
	Draft       bool                 `json:"draft"`
	PreRelease  bool                 `json:"prerelease"`
	CreatedAt   string               `json:"created_at"`
	PublishedAt string               `json:"published_at"`
	Assets      []GithubReleaseAsset `json:"assets"`
	Body        string               `json:"body"`
}

func sha256sum(filepath string) (string, error) {
	file, err := os.Open(filepath)
	if err != nil {
		log.Logger.Errorf("Failed to open download file: %v", err)
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		log.Logger.Errorf("Failed to copy download data when calcute hash: %v", err)
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func DetectLatestGithubRelease(ctx context.Context, timeout time.Duration) (*GithubLatestRelease, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", getGithubLatestReleaseUrl, nil)
	if err != nil {
		log.Logger.Errorf("Create http request for getting latest release failed: %v", err)
		return nil, err
	}
	// req.Header.Set("Content-Type", "application/json")
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Errorf("Send getting latest release request failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	// application/json; charset=utf-8
	// if resp.Header.Get("Content-Type") == "application/json" {
	// if strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
	if strings.HasPrefix(resp.Header.Get("Content-Type"), "application/json") {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Logger.Errorf("Read response of getting latest release failed: %v", err)
			return nil, err
		}
		glr := &GithubLatestRelease{}
		if err := json.Unmarshal(body, glr); err != nil {
			log.Logger.Errorf("Unmarshal response of getting latest release failed: %v", err)
			return nil, err
		}
		return glr, nil
	} else if resp.StatusCode != 200 {
		log.Logger.Errorf("Getting latest release error: %s", resp.Status)
		return nil, err
	} else {
		log.Logger.Errorf("Response of getting latest release is not JSON")
		return nil, err
	}
}

func (asset *GithubReleaseAsset) DownloadRelease(ctx context.Context, downloadTimeout time.Duration, filepath string) error {
	file, err := os.OpenFile(filepath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0755)
	if err != nil {
		log.Logger.Errorf("Failed to create file: %v", err)
		return err
	}
	defer file.Close()
	req, err := http.NewRequestWithContext(ctx, "GET", asset.Url, nil)
	if err != nil {
		log.Logger.Errorf("Create http request for download latest release failed: %v", err)
		return err
	}
	req.Header.Set("Accept", asset.ContentType)
	client := &http.Client{
		Timeout: downloadTimeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Errorf("Send download latest release request failed: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Logger.Errorf("Failed to download latest release: %v", resp.Status)
		return fmt.Errorf("failed to download latest release: %v", resp.Status)
	}
	written, err := io.Copy(file, resp.Body)
	if err != nil {
		log.Logger.Errorf("Failed to copy file: %v", err)
		return err
	}
	// check file hash
	// if written != asset.Size {
	// 	log.Logger.Warn("The number of bytes copied from latest release is not equal with asset's size")
	// }
	log.Logger.Infof("Download %v bytes from: %v", written, asset.Url)
	return nil
}

func (asset *GithubReleaseAsset) DownloadChecksums(ctx context.Context, timeout time.Duration) (map[string]string, error) {
	hashs := make(map[string]string)
	req, err := http.NewRequestWithContext(ctx, "GET", asset.Url, nil)
	if err != nil {
		log.Logger.Errorf("Create http request for download checksums failed: %v", err)
		return hashs, err
	}
	// req.Header.Set("Accept", asset.ContentType)
	req.Header.Set("Accept", "application/octet-stream")
	client := &http.Client{
		Timeout: timeout,
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Logger.Errorf("Send download checksums request failed: %v", err)
		return hashs, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Logger.Errorf("Failed to download checksums: %v", resp.Status)
		return hashs, fmt.Errorf("failed to download checksums: %v", resp.Status)
	}
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 2 {
			log.Logger.Warn("Invalid line format of checksums")
			continue
		}
		hashs[fields[0]] = fields[1]
	}
	if err := scanner.Err(); err != nil {
		log.Logger.Warnf("line scanner error: %v", err)
	}
	return hashs, nil
}
