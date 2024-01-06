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
	downloadWelcomeURL := os.Getenv("GIF_WELCOME_URL")
	downloadRingURL := os.Getenv("GIF_RING_URL")

	if host == "" || username == "" || password == "" {
		log.Fatal("Please set SSH_HOST, SSH_USERNAME, and SSH_PASSWORD environment variables.")
	}

	for {
		err := runSSH(host, username, password, downloadWelcomeURL, downloadRingURL)
		if err != nil {
			log.Println("Error in SSH connection:", err)
		}

		// Wait for 5 seconds before attempting to reconnect
		time.Sleep(5 * time.Second)
	}
}

func runSSH(host, username, password, downloadWelcomeURL, downloadRingURL string) error {
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
	err = downloadFile(client, downloadWelcomeURL, "/mnt/log/welcome.png")
	if err != nil {
		log.Println("Error downloading file:", err)
	}

	// on the very first run download a new image, the mounting logic should be fine.
	err = downloadFile(client, downloadRingURL, "/mnt/log/ring.png")
	if err != nil {
		log.Println("Error downloading file:", err)
	}

	// Infinite loop to maintain the SSH connection
	for {
		// Check if the /usr/etc/gui/screen_240x240/Welcome_Anim_60.png is mounted
		isWelcomeMounted, err := isWelcomeMounted(client)
		if err != nil {
			log.Println("Error checking welcome mount status:", err)
			return err
		} else if isWelcomeMounted {
			fmt.Println("Welcome image is mounted!")
		} else {
			fmt.Println("Welcome image not mounted, mounting now.")
			_, err = mountWelcomeScreen(client)
			if err != nil {
				log.Println("Error mounting welcome image:", err)
				return err
			}
		}

		isRingMounted, err := isRingMounted(client)
		if err != nil {
			log.Println("Error checking ring mount status:", err)
			return err
		} else if isRingMounted {
			fmt.Println("Ring image is mounted!")
		} else {
			fmt.Println("Ring image not mounted, mounting now.")
			_, err = mountRingScreen(client)
			if err != nil {
				log.Println("Error mounting ring image:", err)
				return err
			}
		}

		// Wait for 30 seconds before checking again
		time.Sleep(30 * time.Second)
	}
}

func isWelcomeMounted(client *ssh.Client) (bool, error) {
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

func isRingMounted(client *ssh.Client) (bool, error) {
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
	return strings.Contains(string(output), "/usr/etc/gui/screen_240x240/Ringing_Anim_48.png"), nil
}

func mountWelcomeScreen(client *ssh.Client) (bool, error) {
	// Execute the 'mount' command on the remote server
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	_, err = session.CombinedOutput("mount -o bind /mnt/log/welcome.png /usr/etc/gui/screen_240x240/Welcome_Anim_60.png")
	if err != nil {
		return false, fmt.Errorf("Failed to execute command: %v", err)
	}
	return true, nil
}

func mountRingScreen(client *ssh.Client) (bool, error) {
	// Execute the 'mount' command on the remote server
	session, err := client.NewSession()
	if err != nil {
		return false, fmt.Errorf("Failed to create session: %v", err)
	}
	defer session.Close()

	_, err = session.CombinedOutput("mount -o bind /mnt/log/ring.png /usr/etc/gui/screen_240x240/Ringing_Anim_48.png")
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
