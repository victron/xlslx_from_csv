package int_libs

import (
	"golang.org/x/crypto/ssh"
	"log"
	"strconv"
	"github.com/pkg/sftp"
	"fmt"
	"time"
	"errors"
	"strings"
	"bytes"
	//"unicode"
	//"github.com/alecthomas/gometalinter/_linters/src/honnef.co/go/tools/staticcheck"
	"os"
	"sort"
)

const RFC3339_short = "2006-01-02T15:04 UTC"
const OSI_date = "20060102"
const OSI_date_time = "20060102T1504"
const SAM_file_time = "A20060102.1504"

type SiteAuthen struct {
	Ip       string
	Port     int
	UserName string
	UserPass string
}

type Task struct {
	HomeDir string
	Site    string
	Start   time.Time
	Stop    time.Time
	Interval int
	// /opt/5620sam/lte/stats
	RootStatDIR string
	//  ["20171016", "20171019", "20171022", "20171025", "20171028"]
	DatesDirs []string
	// mme
	NE_Dir string
	// 172.20.44.5-kv-po-sgsn3_3
	SiteDir         string
	FileLastRecived string
	// full path to dir
	FileLastDir          string
	FilesToRecive        []string
	Dir_to_FilesToRecive string
	FileArch             string
	foldersTimes         map[string]time.Time
	foldersTimesT        map[time.Time]string
}

// constractor to init maps inside struct
func InitTask() *Task {
	var task Task
	task.foldersTimes = make(map[string]time.Time)
	task.foldersTimesT = make(map[time.Time]string)
	return &task
}

type Connection struct {
	SshConnect  *ssh.Client
	SftpConnect *sftp.Client
}

func (server SiteAuthen) SSH() *ssh.Client {
	//var hostKey ssh.PublicKey
	config := &ssh.ClientConfig{
		User: server.UserName,
		Auth: []ssh.AuthMethod{
			ssh.Password(server.UserPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}
	// "localhost:22"
	socket := server.Ip + ":" + strconv.Itoa(server.Port)
	// Dial your ssh server.
	conn, err := ssh.Dial("tcp", socket, config)
	if err != nil {
		log.Fatal("[FATAL]", "unable to connect: ", err)
	}
	log.Println("[INFO]", "connected to:", socket)
	return conn
	//defer conn.Close()
}

func (conect *Connection) SFTP() {
	// open an SFTP session over an existing ssh connection.
	sftp_, err := sftp.NewClient(conect.SshConnect)
	if err != nil {
		log.Fatal(err)
	}
	conect.SftpConnect = sftp_
	//defer sftp_.Close()
}

// help method for GetSftpInfo
func (task *Task) getDirs(sftp_ *sftp.Client) error {
	out_err := errors.New("folders with stat not found")
	files, err := sftp_.ReadDir(task.RootStatDIR)
	if err != nil {
		log.Println("[ERROR]", "could not read dir:", task.RootStatDIR)
		return out_err
	}
	//var dir_time time.Time
	for _, file := range files {
		if file.IsDir() {
			dir_time, err := time.Parse(OSI_date, file.Name())
			task.DatesDirs = append(task.DatesDirs, file.Name())
			task.foldersTimes[file.Name()] = dir_time
			task.foldersTimesT[dir_time] = file.Name()
			if err != nil {
				out_err = nil
				continue
				//fmt.Print(file.Name() + "\t")
			}
		}
	}
	sort.Strings(task.DatesDirs)
	return out_err
}

// collect data for task
func (task *Task) GetSftpInfo(sftp_ *sftp.Client) {
	//stat_dir := task.RootStatDIR
	stat_dir_modif := ""
	loop := true

	// get STAT directory
	mainloop:
	for loop {
		task.DatesDirs = []string{}
		log.Println("[INFO]", "directory to search stat:\n"+task.RootStatDIR)
		fmt.Print("overite stat dirictory if need: ")
		fmt.Scanf("%s\n", &stat_dir_modif)
		if stat_dir_modif != "" {
			task.RootStatDIR = stat_dir_modif
		}
		// check if directory exists
		if file, err := sftp_.Stat(task.RootStatDIR); err != nil || !file.IsDir() {
			loop = true
			log.Printf("[ERROR] directory %s not exists!!!!", task.RootStatDIR)
			fmt.Print("enter PATH to stat dir: ")
			continue mainloop
		} else {
			log.Println("[INFO]", "stat dir found, colecting dates folders")
			task.getDirs(sftp_)
			loop = false
		}

		var err error
		fmt.Println("give START date/time")
		task.Start, err = task.checkDates()
		if err != nil {
			loop = true
			log.Println("[ERROR]", err)
			continue mainloop
		} else {
			log.Println("[DEBUG]", "start date:", task.Start)
			task.Start = task.Start.Round(15* time.Minute)
			log.Println("[DEBUG]", "start date (rounding):", task.Start)
			loop = false
		}

		fmt.Println("give STOP date/time")
		task.Stop, err = task.checkDates()
		if err != nil {
			loop = true
			log.Println("[ERROR]", err)
			continue mainloop
		} else {
			log.Println("[DEBUG]", "stop date:", task.Start)
			task.Stop = task.Stop.Round(15* time.Minute)
			log.Println("[DEBUG]", "stop date (rounding):", task.Stop)
			if task.Start.After(task.Stop) || task.Start.Equal(task.Stop) {
				log.Println("[ERROR]", "START date should be before STOP date")
				loop = true
				continue mainloop
			}
			loop = false
		}
		task.NE_Dir = "mme"
		var NE_folders []string
		startSearchDir := task.RootStatDIR + "/" + task.Start.Format(OSI_date) + "/" + task.NE_Dir
		log.Println("[DEBUG]", "start search site dir in:", startSearchDir)
		folders, err := sftp_.ReadDir(startSearchDir)
		if err != nil {
			log.Println("[ERROR]", "could not read dir:", task.RootStatDIR)
			//return out_err
		}
		for _, folder := range folders {
			if folder.IsDir() {
				NE_folders = append(NE_folders, folder.Name())
			}
		}
		fmt.Println("Site folders:", NE_folders)
		fmt.Println("give Site folder form available list:")
		fmt.Scanf("%s\n", &task.SiteDir)
	}

	log.Println("[DEBUG]", "created task:", task)
}

// check if folder with stat exist on server for requested by user time
func (task *Task) checkDates() (time.Time, error) {
	// request user start date to build statistics
	loop := true
	var output_time time.Time
	var user_input_time time.Time
	var err error
	var output_error error = errors.New("folder missing")

	for loop {
		fmt.Println("available dirs with stat:", task.DatesDirs)
		fmt.Print("enter date/time in format like \"20171123T1600\" : ")
		user_input := ""
		fmt.Scanf("%s\n", &user_input)
		log.Println("[DEBUG]", "user input:", user_input)
		user_input_time, err = time.Parse(OSI_date_time, user_input)
		if err != nil {
			loop = true
			log.Println("[DEBUG]", "parse error:", err)
			log.Println("[ERROR]", "wrong input format")
			continue
		}
		output_time = user_input_time
		// put zeros in hoour, minutes,....
		user_input_time_trun := user_input_time.Truncate(24 * time.Hour)
		dir, ok := task.foldersTimesT[user_input_time_trun]
		if ok {
			log.Println("[INFO]", "found directory:", dir)
			output_error = nil
			break
		} else {
			log.Println("[ERROR]", "no folder related to date", user_input_time_trun)
			continue
		}
	}
	return output_time, output_error
}

func (task *Task) GenerateFileList(sftp_ *sftp.Client) int {
	// need redoo for multiple folders
	task.Dir_to_FilesToRecive = ""
	var fileLastTime time.Time
	if task.FileLastRecived == "" {
		task.Dir_to_FilesToRecive = task.RootStatDIR + "/" + task.foldersTimesT[task.Start.Truncate(24*time.Hour)] +
			"/" + task.NE_Dir + "/" + task.SiteDir
		log.Println("[DEBUG]", "folder to get stat files:", task.Dir_to_FilesToRecive)
		fileLastTime = task.Start
	} else {
		_, err := time.Parse(SAM_file_time, strings.Split(task.FileLastRecived, "+")[0])
		log.Println("[DEBUG]", "FileLastRecived:", task.FileLastRecived)
		check(err)
		fileLastTime, _ = time.Parse(SAM_file_time, strings.Split(task.FileLastRecived, "+")[0])
		switch {
		case fileLastTime.Hour() == 23 && fileLastTime.Minute() == 45 && fileLastTime.Equal(task.Stop):
			log.Println("[DEBUG]", "stopping work in lastFile dir")
			task.Dir_to_FilesToRecive = task.RootStatDIR + "/" + fileLastTime.Format(OSI_date) + "/" + task.NE_Dir + "/" + task.SiteDir
			return 0
		case fileLastTime.Hour() == 23 && fileLastTime.Minute() == 45 && fileLastTime.Before(task.Stop):
			log.Println("[DEBUG]", "switch to next dir")
			newDirTime := fileLastTime.Add(time.Hour * 24)
			task.Dir_to_FilesToRecive = task.RootStatDIR + "/" + newDirTime.Format(OSI_date) + "/" + task.NE_Dir + "/" + task.SiteDir
		case fileLastTime.Hour() < 23 || fileLastTime.Minute() < 45:
			log.Println("[DEBUG]", "work in dir of lastFile")
			task.Dir_to_FilesToRecive = task.RootStatDIR + "/" + fileLastTime.Format(OSI_date) + "/" + task.NE_Dir + "/" + task.SiteDir
		default:
			log.Println("[DEBUG]", "default case, should not works")
			task.Dir_to_FilesToRecive = task.RootStatDIR + "/" + fileLastTime.Format(OSI_date) + "/" + task.NE_Dir + "/" + task.SiteDir
			return 0
		}
	}

	log.Println("[DEBUG]", "Dir_to_FilesToRecive:", task.Dir_to_FilesToRecive)
	files, err := sftp_.ReadDir(task.Dir_to_FilesToRecive)
	check(err)
	filesNumbers := 0
	for _, file := range files {
		fileName := file.Name()
		log.Println("[DEBUG]","testing file:", fileName)
		fileTimePart := strings.Split(fileName, "+")[0]
		fileTime, err := time.Parse(SAM_file_time, fileTimePart)
		if err == nil {
			log.Println("[DEBUG]", "fileTime:", fileTime, "task.Stop:", task.Stop, "fileLastTime:", fileLastTime)
			if fileTime.Before(task.Stop) && fileTime.After(fileLastTime) {
				task.FilesToRecive = append(task.FilesToRecive, fileName)
				filesNumbers ++
			}
		} else {
			log.Println("[ERROR]", "could not get time from file:", file.Name())
		}

	}
	log.Println("[DEBUG]", "file list to recieve:", task.FilesToRecive)
	return filesNumbers
}

//func (task *Task) GenFoldersList() error {
//	if task.Start.Truncate(24 * time.Hour) == task.Stop.Truncate(24 * time.Hour) {
//
//	}
//}

func (task *Task) CompressFiles(connect Connection) error {
	current_time := time.Now()
	task.FileArch = "robot_" + current_time.Format(OSI_date_time) + ".tgz"
	archFileName := task.HomeDir + "/" + task.FileArch
	switch {
	case len(task.FilesToRecive) == 0:
		return errors.New("no files in filelist")
		return errors.New("empty task.FilesToRecive, nothing to receive")
	case len(task.FilesToRecive) <= 10:
		filesList := strings.Join(task.FilesToRecive, " ")
		command := "tar cvzf " + archFileName + " -C " + task.Dir_to_FilesToRecive + " " + filesList
		return connect.RunOneCmd(command)
	case len(task.FilesToRecive) > 10:
		command := "tar cvf " + strings.Split(archFileName, ".")[0] + " -C " + task.Dir_to_FilesToRecive + " " + task.FilesToRecive[0]
		err := connect.RunOneCmd(command)
		if err != nil {
			return err
		}
		for _, file := range task.FilesToRecive[1:] {
			command := "tar rvf " + strings.Split(archFileName, ".")[0] + " -C " + task.Dir_to_FilesToRecive + " " + file
			err := connect.RunOneCmd(command)
			if err != nil {
				return err
			}
		}
		command = "gzip -S .tgz " + strings.Split(archFileName, ".")[0]
		return connect.RunOneCmd(command)

	default:
		command := "tar cvzf " + archFileName + " -C " + task.Dir_to_FilesToRecive + " " + "."
		return connect.RunOneCmd(command)
	}
}

// wraper on ssh.Session; create new session, execute one command, close session
func (connect *Connection) RunOneCmd(cmd string) error {
	var buffer bytes.Buffer
	session, err := connect.SshConnect.NewSession()
	if err != nil {
		log.Println("[FATAL]", "can't create session", err)
		return err
	}
	session.Stdout = &buffer
	defer session.Close()
	log.Println("[INFO]", "executing:", cmd)
	// with CombinedOutput method
	//b, err := session.CombinedOutput(command)
	//check(err)
	//log.Println("[DEBUG]", "create tar file", string(b))
	if err := session.Run(cmd); err != nil {
		log.Fatal("[FATAL]", "Failed to run: ", err)
		return err
	} else {
		log.Println("[DEBUG]", "result of\n", cmd, ":\n", buffer.String())
		return nil
	}
}

func (connect *Connection) SftpCopy(src, dst string) error {
	// Open the source file
	srcFile, err := connect.SftpConnect.Open(src)
	if err != nil {
		log.Println("[ERROR]", err)
		log.Fatal("[FATAL]", "could not open SRC file:", src)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		log.Println("[ERROR]", err)
		log.Fatal("[FATAL]", "could not create DST file", dst)
	}
	defer dstFile.Close()

	// Copy the file
	log.Println("[INFO]", "copy remote:", src, "to local:", dst)

	if _bytes, err := srcFile.WriteTo(dstFile); err != nil {
		log.Println("[ERROR]", "coppied bytes:", _bytes)
		log.Println("[ERROR]", "could not copy file from remote:", err)
		return errors.New("sftp copy failed!!!!!!!!!!!!!!!!")
	} else {
		log.Println("[INFO]", "coppied bytes:", _bytes)
		return nil
	}
}
