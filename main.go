package main

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func main() {
	host := os.Getenv("SSH_HOST")
	username := os.Getenv("SSH_USERNAME")
	password := os.Getenv("SSH_PASSWORD")
	downloadURL := os.Getenv("GIF_URL")

	if host == "" || username == "" || password == "" {
		log.Fatal("Please set SSH_HOST, SSH_USERNAME, and SSH_PASSWORD environment variables.")
	}

	for {
		err := runSSH(host, username, password, downloadURL)
		if err != nil {
			log.Println("Error in SSH connection:", err)
		}

		// Wait for 5 seconds before attempting to reconnect
		time.Sleep(5 * time.Second)
	}
}

func runSSH(host, username, password, downloadURL string) error {
	// Initialize SSH configuration
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	// Initial connection to SSH server
	client, err := ssh.Dial("tcp", host+":22", config)
	if err != nil {
		return fmt.Errorf("Failed to dial: %v", err)
	}
	defer client.Close()

	// on the very first run download a new image, the mounting logic should be fine.
	err = downloadFile(client, downloadURL, "/mnt/log/custom.png")
	if err != nil {
		log.Println("Error downloading file:", err)
	}

	// Infinite loop to maintain the SSH connection
	for {
		// Check if the /usr/etc/gui/screen_240x240/Welcome_Anim_60.png is mounted
		isMounted, err := isFileMounted(client)
		if err != nil {
			log.Println("Error checking mount status:", err)
		} else if isMounted {
			fmt.Println("File is mounted!")
		} else {
			fmt.Println("File is not mounted, download and mount.")

			err = downloadFile(client, downloadURL, "/mnt/log/custom.png")
			if err != nil {
				log.Println("Error downloading file:", err)
			}
			isMounted, err = mountFile(client)
			if err != nil {
				log.Println("Error mounting file:", err)
			}
		}
		// Wait for 30 seconds before checking again
		time.Sleep(30 * time.Second)
	}
}

func isFileMounted(client *ssh.Client) (bool, error) {
	// Execute the 'mount' command on the remote server
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput("mount")
	if err != nil {
		return false, fmt.Errorf("Failed to execute command: %v", err)
	}

	// Check if the file path is present in the 'mount' output
	return strings.Contains(string(output), "/usr/etc/gui/screen_240x240/Welcome_Anim_60.png"), nil
}

func mountFile(client *ssh.Client) (bool, error) {
	// Execute the 'mount' command on the remote server
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	_, err = session.CombinedOutput("mount -o bind /mnt/log/custom.png /usr/etc/gui/screen_240x240/Welcome_Anim_60.png")
	if err != nil {
		return false, fmt.Errorf("Failed to execute command: %v", err)
	}
	return true, nil
}

func umountFile(client *ssh.Client) (bool, error) {
	// Execute the 'mount' command on the remote server
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	_, err = session.CombinedOutput("umount /usr/etc/gui/screen_240x240/Welcome_Anim_60.png")
	if err != nil {
		return false, fmt.Errorf("Failed to execute command: %v", err)
	}
	return true, nil
}

func downloadFile(client *ssh.Client, url, remotePath string) error {
	// Execute the 'curl' command on the remote server to download the file
	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	cmd := fmt.Sprintf("curl -o %s %s --insecure", remotePath, url)
	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return fmt.Errorf("Failed to execute command: %v\nOutput: %s", err, string(output))
	}

	fmt.Printf("Downloaded file from %s to %s\n", url, remotePath)
	return nil
}
