package api

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kvmgo/lib"
	"kvmgo/utils"

	"libvirt.org/go/libvirt"
)

func CheckPoolExists(conn *libvirt.Connect, poolName string) bool {
	if _, err := conn.LookupStoragePoolByName(poolName); err != nil {
		return false
	}
	return true
}

func GetPoolPath(conn *libvirt.Connect, poolName string) (string, error) {
	log.Printf("Checking if Pool %s exists\n", poolName)

	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return "", err
	}
	defer pool.Free()

	xmlDesc, err := pool.GetXMLDesc(0)
	if err != nil {
		return "", err
	}

	// Parse the XML to get the path
	// This is a simple string search, you might want to use proper XML parsing for more complex scenarios
	pathStart := strings.Index(xmlDesc, "<path>")
	pathEnd := strings.Index(xmlDesc, "</path>")
	if pathStart == -1 || pathEnd == -1 {
		return "", fmt.Errorf("path not found in pool XML")
	}

	path := xmlDesc[pathStart+6 : pathEnd]
	return path, nil
}

// ListPoolVolumes lists the volumes associated with the Storage Pool such as ubuntu for Images
func ListPoolVolumes(conn *libvirt.Connect, poolName string) ([]*lib.Volume, error) {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return nil, fmt.Errorf("failed to lookup pool: %v", err)
	}
	defer pool.Free()

	// Refresh the pool to get the latest state
	if err := pool.Refresh(0); err != nil {
		return nil, fmt.Errorf("failed to refresh pool: %v", err)
	}

	volumes, err := pool.ListAllStorageVolumes(0)
	if err != nil {
		return nil, fmt.Errorf("failed to list volumes: %v", err)
	}

	var vols []*lib.Volume

	for _, vol := range volumes {
		// path, err := vol.GetPath()
		volume, err := lib.NewVolume(&vol)
		if err != nil {
			vol.Free()
			return nil, fmt.Errorf("failed to get volume path: %v", err)
		}
		// volumePaths = append(volumePaths, path)
		vols = append(vols, volume)
		vol.Free()
	}

	return vols, nil
}

func DeletePool(conn *libvirt.Connect, poolName string, deleteContents bool) error {
	pool, err := conn.LookupStoragePoolByName(poolName)
	if err != nil {
		return fmt.Errorf("failed to lookup pool: %v", err)
	}
	defer pool.Free()

	// Stop the pool
	err = pool.Destroy()
	if err != nil {
		return fmt.Errorf("failed to stop pool: %v", err)
	}

	if deleteContents {
		// Delete all volumes in the pool
		volumes, err := pool.ListAllStorageVolumes(0)
		if err != nil {
			return fmt.Errorf("failed to list volumes: %v", err)
		}
		for _, vol := range volumes {
			err = vol.Delete(0)
			if err != nil {
				vol.Free()
				return fmt.Errorf("failed to delete volume: %v", err)
			}
			vol.Free()
		}

		// Delete the pool itself (including on-disk data)
		err = pool.Delete(libvirt.STORAGE_POOL_DELETE_NORMAL)
		if err != nil {
			return fmt.Errorf("failed to delete pool data: %v", err)
		}
	}

	// Undefine the pool
	err = pool.Undefine()
	if err != nil {
		return fmt.Errorf("failed to undefine pool: %v", err)
	}

	return nil
}

// FetchImageUrl pulls an image using a URL to a Directory
// This .img file will be used by the VM's as a Base Image by defining it using the StorageCreateXML
func FetchImageUrl(url, dir string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("passed empty URL")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(500 * time.Second),
	}

	// Send GET request
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to GET from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	log.Printf("Download Started")

	// Extract the filename from the URL
	fileName := filepath.Base(url)
	if fileName == "." || fileName == "/" {
		fileName = "downloaded_image"
	}

	// If the filename doesn't have an extension, try to get it from the Content-Type
	if !strings.Contains(fileName, ".") {
		contentType := resp.Header.Get("Content-Type")
		ext := ""
		switch contentType {
		case "application/x-qemu-disk":
			ext = ".qcow2"
		case "application/x-raw-disk-image":
			ext = ".img"
		}
		fileName += ext
	}

	// Create the full file path
	filePath := filepath.Join(dir, fileName)

	log.Printf("Creating file at: %s", filePath)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer out.Close()

	fileSize := resp.ContentLength

	log.Printf("Starting to copy data to file. Size is %d\n", fileSize)

	written, err := io.Copy(out, resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to save file %s: %v", filePath, err)
	}
	log.Printf("Copied %d bytes to file", written)

	return filePath, nil
}

// Downloads an Image with Progress logs
func DownloadImage(url, dir string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("passed empty URL")
	}

	if err := utils.CreateDirIfNotExist(dir); err != nil {
		return "", fmt.Errorf("failed to create the dir %s", dir)
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(500 * time.Second),
	}

	// Send GET request
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to GET from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	log.Printf("Download Started")

	// Extract the filename from the URL
	fileName := filepath.Base(url)
	if fileName == "." || fileName == "/" {
		fileName = "downloaded_image"
	}

	// If the filename doesn't have an extension, try to get it from the Content-Type
	if !strings.Contains(fileName, ".") {
		contentType := resp.Header.Get("Content-Type")
		ext := ""
		switch contentType {
		case "application/x-qemu-disk":
			ext = ".qcow2"
		case "application/x-raw-disk-image":
			ext = ".img"
		}
		fileName += ext
	}

	// Create the full file path
	filePath := filepath.Join(dir, fileName)

	log.Printf("Creating file at: %s", filePath)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer out.Close()

	fileSize := resp.ContentLength
	log.Printf("Starting to copy data to file. Size is %d bytes\n", fileSize)

	// Progress tracking variables
	var written int64
	buf := make([]byte, 32*1024) // 32 KB buffer
	startTime := time.Now()

	// Function to log progress
	logProgress := func() {
		elapsed := time.Since(startTime).Seconds()
		speed := float64(written) / elapsed / (1024 * 1024) // Speed in MB/s
		percent := float64(written) / float64(fileSize) * 100
		log.Printf("Progress: %.2f%%, Speed: %.2f MB/s, Written: %d bytes\n", percent, speed, written)
	}

	// Read from response body and write to file with progress logging
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				logProgress()
			}
			if ew != nil {
				return "", fmt.Errorf("failed to write to file %s: %v", filePath, ew)
			}
			if nr != nw {
				return "", fmt.Errorf("failed to write all bytes to file %s", filePath)
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read from response body: %v", er)
		}
	}

	log.Printf("Download completed. Total bytes written: %d\n", written)

	return filePath, nil
}

// Downloads an Image with Progress logs
func DownloadImageProgress(url, dir string) (string, error) {
	if url == "" {
		return "", fmt.Errorf("passed empty URL")
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: time.Duration(2600 * time.Second),
	}

	// Send GET request
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("failed to GET from %s: %v", url, err)
	}
	defer resp.Body.Close()

	// Check server response
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("bad status: %s", resp.Status)
	}

	log.Printf("Download Started")

	// Extract the filename from the URL
	fileName := filepath.Base(url)
	if fileName == "." || fileName == "/" {
		fileName = "downloaded_image"
	}

	// If the filename doesn't have an extension, try to get it from the Content-Type
	if !strings.Contains(fileName, ".") {
		contentType := resp.Header.Get("Content-Type")
		ext := ""
		switch contentType {
		case "application/x-qemu-disk":
			ext = ".qcow2"
		case "application/x-raw-disk-image":
			ext = ".img"
		}
		fileName += ext
	}

	// Create the full file path
	filePath := filepath.Join(dir, fileName)

	log.Printf("Creating file at: %s", filePath)
	out, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file %s: %v", filePath, err)
	}
	defer out.Close()

	fileSize := resp.ContentLength
	log.Printf("Starting to copy data to file. Size is %d bytes\n", fileSize)

	// Progress tracking variables
	var written int64
	buf := make([]byte, 128*1024) // 64 KB buffer
	startTime := time.Now()
	lastLoggedTime := time.Now()

	// Function to log progress
	logProgress := func() {
		elapsed := time.Since(startTime).Seconds()
		speed := float64(written) / elapsed / (1024 * 1024) // Speed in MB/s
		percent := float64(written) / float64(fileSize) * 100
		fmt.Printf("\rProgress: %.2f%%, Speed: %.2f MB/s, Written: %d bytes", percent, speed, written)
	}

	// Read from response body and write to file with progress logging
	for {
		nr, er := resp.Body.Read(buf)
		if nr > 0 {
			nw, ew := out.Write(buf[0:nr])
			if nw > 0 {
				written += int64(nw)
				if time.Since(lastLoggedTime).Seconds() >= 2 {
					logProgress()
					lastLoggedTime = time.Now()
				}
			}
			if ew != nil {
				return "", fmt.Errorf("failed to write to file %s: %v", filePath, ew)
			}
			if nr != nw {
				return "", fmt.Errorf("failed to write all bytes to file %s", filePath)
			}
		}
		if er != nil {
			if er == io.EOF {
				break
			}
			return "", fmt.Errorf("failed to read from response body: %v", er)
		}
	}

	logProgress() // Log final progress

	totalTime := time.Since(startTime).Seconds()
	fmt.Printf("\nDownload completed in %.2f seconds.\n", totalTime)

	fmt.Println("\nDownload completed.")
	log.Printf("Download completed. Total bytes written: %d\n", written)

	return filePath, nil
}
