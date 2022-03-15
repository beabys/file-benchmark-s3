package cmd

import (
	"encoding/csv"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/briandowns/spinner"
	"github.com/spf13/cobra"
	"gitlab.com/beabys/file-benchmark-s3/adapters"
	"gitlab.com/beabys/file-benchmark-s3/file"
)

type Files []*file.File
type Proccess struct {
	StartTime       time.Time
	EndTime         time.Time
	Duration        time.Duration
	FilesProccessed int
	concurrentJobs  int
}

var concurrency, maxFileSize, minFileSize, filesNo int
var workingFolder, outputPath, path, originalPath, downloadPath string
var spinAnimation bool

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	// cobra.CheckErr(rootCmd.Execute())
}

// init function
func init() {
	//adding required flag parmeters
	rootCmd.Flags().StringVarP(&outputPath, "output", "o", "./results", "Output folder of the results.")
	rootCmd.Flags().IntVarP(&filesNo, "files-number", "f", 1, "Number of files to be created.")
	rootCmd.Flags().IntVarP(&concurrency, "concurrent-jobs", "c", 1, "Number of jobs excuted writing files. (Maximum 10)")
	rootCmd.Flags().IntVarP(&maxFileSize, "max-file-fize", "m", 2000, "Max size in Mb allocated for each test file.")
	rootCmd.Flags().IntVarP(&minFileSize, "min-file-fize", "i", 5, "Min size in Mb allocated for each test file.")
	rootCmd.Flags().StringVarP(&originalPath, "path", "p", "./", "Path where the test files will be created.")
	rootCmd.Flags().BoolVarP(&spinAnimation, "spinner", "s", false, "Show spinner, just to know if still alive the proccess, don't use if running on background")
}

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "file-benchmark-s3",
	Short: "benchmark for basic file operations on s3 object storage",
	Long: `
	file-benchmark is a tool to test speed on basic operations on s3 object storage.
	useful when need to test speed on S3 object.`,
	Run: func(cmd *cobra.Command, args []string) {
		proccess := &Proccess{
			StartTime:      time.Now(),
			concurrentJobs: concurrency,
		}
		workingFolder = randomString(6)
		originalPath = splitString(originalPath, "/", true)
		outputPath = splitString(outputPath, "/", true)
		path = splitString(originalPath+workingFolder, "/", true)
		downloadPath = splitString(fmt.Sprintf("%sdownload_%s", originalPath, workingFolder), "/", true)

		// Creating new S3 Config
		s3Config := &adapters.S3Config{
			Region:           "",
			AccessKeyID:      "",
			SecretAccessKey:  "",
			Endpoint:         "",
			Bucket:           "",
			DisableSSL:       true,
			S3ForcePathStyle: true,
		}

		// Creating new session
		session, err := adapters.NewS3(s3Config)
		if err != nil {
			log.Fatalf("unable to create S3 session, Please verify config values")
			os.Exit(1)
		}

		if err := session.Connect(); err != nil {
			log.Fatalf("unable to Connect with bucket %s, Please verify config values", session.Bucket)
			os.Exit(1)
		}

		fmt.Printf("Creating %s folders\n", outputPath)
		// Create output folder
		if _, err := os.Stat(outputPath); os.IsNotExist(err) {
			err := os.Mkdir(outputPath, 0777)
			if err != nil {
				log.Fatalf("unable to create folder %s", outputPath)
				os.Exit(1)
			}
		}
		fmt.Printf("Creating tmp testing folders on %s\n", originalPath)
		// Create working folder, this will be deletaed at the end of the execution
		if _, err := os.Stat(path); os.IsNotExist(err) {
			err := os.Mkdir(path, 0777)
			if err != nil {
				log.Fatalf("unable to create folder %s on %s", path, originalPath)
				os.Exit(1)
			}
		}

		// Create working folder, this will be deleted at the end of the execution
		if _, err := os.Stat(downloadPath); os.IsNotExist(err) {
			err := os.Mkdir(downloadPath, 0777)
			if err != nil {
				log.Fatalf("unable to create folder %s on %s", downloadPath, originalPath)
				os.Exit(1)
			}
		}

		fmt.Print("\nStarting to create files, this proccess can take a while\n")

		var filesSlice Files
		proccessFiles(&filesSlice, session)

		fmt.Println("Creating Report on " + outputPath)
		err = writeReport(&filesSlice, proccess)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		fmt.Printf("Execution total time : %v\n", time.Since(proccess.StartTime))
	},
}

// Create a new file struct
func newFile(s3Session *adapters.S3Session) *file.File {
	random := rand.New(rand.NewSource(time.Now().UnixNano()))
	minSize := 1024 * 1024 * minFileSize
	maxSize := 1024 * 1024 * maxFileSize
	size := random.Intn(maxSize-minSize) + minSize
	name := randomString(16)
	return &file.File{
		Size:         size,
		Path:         path,
		DownloadPath: downloadPath,
		Name:         name,
		S3Session:    s3Session,
		Execution:    &file.ExecutionTime{},
	}
}

// randomString return a random string
func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789")
	var b strings.Builder
	for i := 0; i < length; i++ {
		b.WriteRune(chars[rand.Intn(len(chars))])
	}
	return b.String()
}

// Start Procceessing files Proccess
func proccessFiles(fs *Files, s3session *adapters.S3Session) error {
	// creating wait group type
	var wg sync.WaitGroup

	// create the Spinner
	if spinAnimation {
		s := spinner.New(spinner.CharSets[43], 100*time.Millisecond)
		s.Start()
		time.Sleep(4 * time.Second)
		defer s.Stop()
	}

	// Create Jobs Pool
	files := make(chan *file.File)

	//appending files definitions into slice
	for i := 0; i < filesNo; i++ {
		*fs = append(*fs, newFile(s3session))
	}

	// need to iterate on the workers,
	// each free worker can take one job(file) from the pool
	for w := 1; w <= concurrency; w++ {
		wg.Add(1)
		// Consumer
		go workerConsumer(files, w, &wg)
	}

	// Create jobs pool
	for _, job := range *fs {
		files <- job
	}

	// after complete close files
	close(files)

	// add wait to stop until all the files get proccessed
	wg.Wait()

	fmt.Printf("\nRemoving tmp testing folders ... ")
	err := os.RemoveAll(path)
	if err != nil {
		return fmt.Errorf("fail \nUnable to remove folder %s", path)
	}
	err = os.RemoveAll(downloadPath)
	if err != nil {
		return fmt.Errorf("fail \nUnable to remove folder %s", downloadPath)
	}
	fmt.Print("done \n")
	return nil
}

// Worker executing the file creation
func workerConsumer(files <-chan *file.File, w int, wg *sync.WaitGroup) {
	defer wg.Done()
	for file := range files {
		execute(file, w)
	}
}

// execute the test per file (creation, copy and delete)
func execute(f *file.File, w int) {
	// Start creation of file
	start := time.Now()
	// file := newFile()
	f.Create()
	f.Execution.CreationDuration = time.Since(start)

	// Start Upload proccess of file
	startUpload := time.Now()
	src := fmt.Sprintf("%s%s", f.Path, f.Name)
	dst := fmt.Sprintf("%s/%s", workingFolder, f.Name)
	f.S3Session.UploadObject(src, dst)
	f.Execution.UploadDuration = time.Since(startUpload)

	// Start Download proccess of file
	startDownload := time.Now()
	srcDownload := fmt.Sprintf("%s/%s", workingFolder, f.Name)
	dstDownload := fmt.Sprintf("%s%s", f.DownloadPath, f.Name)
	f.S3Session.DownloadObject(srcDownload, dstDownload)
	f.Execution.DownloadDuration = time.Since(startDownload)

	// Delete S3Object storage file
	startDeleteS3 := time.Now()
	s3DeletePath := fmt.Sprintf("%s/%s", workingFolder, f.Name)
	err := f.S3Session.DeleteObject(s3DeletePath)
	if err != nil {
		fmt.Println(fmt.Sprintf("error removing file %s from bucket %s", s3DeletePath, f.S3Session.Bucket))
	}
	f.Execution.DeleteUploadDuration = time.Since(startDeleteS3)

	// Delete original file
	startDelete := time.Now()
	err = os.Remove(fmt.Sprintf("%s%s", f.Path, f.Name))
	if err != nil {
		fmt.Println("error removing file" + f.Path + f.Name)
	}
	f.Execution.DeleteLocalDuration = time.Since(startDelete)

	// Delete Downloaded file
	startDeleteDownload := time.Now()
	err = os.Remove(fmt.Sprintf("%s%s", f.DownloadPath, f.Name))
	if err != nil {
		fmt.Println("error removing file " + f.DownloadPath + f.Name)
	}
	f.Execution.DeleteDownloadDuration = time.Since(startDeleteDownload)
}

// writeReport create the file report in csv format
func writeReport(FS *Files, p *Proccess) error {
	file, err := os.Create(outputPath + "result.csv")
	if err != nil {
		return err
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// set Headers
	header := []string{
		"File path",
		"File size in MB",
		"MD5",
		"Download File path",
		"Creation time in Seconds",
		"Upload time in Seconds",
		"Download time in Seconds",
		"Delete time in Seconds",
		"Delete Downloaded Copy time in Seconds",
		"Delete Object on S3 time in Seconds",
	}
	err = writer.Write(header)
	if err != nil {
		return err
	}
	for _, file := range *FS {
		var sizeMB float32 = (float32(file.Size) / 1024) / 1024
		value := []string{
			file.Path + file.Name,
			fmt.Sprintf("%v", sizeMB),
			file.MD5,
			file.DownloadPath + file.Name,
			fmt.Sprintf("%v", file.Execution.CreationDuration.Seconds()),
			fmt.Sprintf("%v", file.Execution.UploadDuration.Seconds()),
			fmt.Sprintf("%v", file.Execution.DownloadDuration.Seconds()),
			fmt.Sprintf("%v", file.Execution.DeleteLocalDuration.Seconds()),
			fmt.Sprintf("%v", file.Execution.DeleteDownloadDuration.Seconds()),
			fmt.Sprintf("%v", file.Execution.DeleteUploadDuration.Seconds()),
		}

		err := writer.Write(value)
		if err != nil {
			return err
		}
	}

	// write txt file
	file2, err := os.Create(outputPath + "result.txt")
	if err != nil {
		return err
	}
	defer file2.Close()

	p.EndTime = time.Now()
	p.Duration = time.Since(p.StartTime)
	p.FilesProccessed = len(*FS)
	value2 := ""
	value2 += fmt.Sprintf("Start Time: %s \n", p.StartTime.String())
	value2 += fmt.Sprintf("End Time: %s \n", p.EndTime.String())
	value2 += fmt.Sprintf("Duration: %v \n", p.Duration)
	value2 += fmt.Sprintf("Files Proccessed: %d \n", p.FilesProccessed)
	value2 += fmt.Sprintf("Concurrent Jobs : %d \n", p.concurrentJobs)
	_, err = file2.WriteString(value2)
	if err != nil {
		return err
	}
	file2.Sync()
	return nil
}

func splitString(s string, separator string, allowPrepend bool) string {
	prepend := true
	var newSlice []string
	var newString string
	dataSliced := strings.Split(s, "/")
	if len(dataSliced) == 1 && dataSliced[0] != "." {
		dataSliced = append([]string{"."}, dataSliced...)
	}
	for key, value := range dataSliced {
		if (value == "." || value == "..") && key == 0 {
			prepend = false
		}
		if value != "" {
			newSlice = append(newSlice, value)
		}
	}
	if prepend && allowPrepend {
		newString = "/"
	}
	newString += strings.Join(newSlice, "/")
	newString += "/"
	return newString

}
