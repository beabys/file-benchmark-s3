# file-benchmark-s3

file-benchmark-s3 is a tool to test speed on basic operations on files
useful when need to test speed on S3 Object Storage.

## Usage:

file-benchmark [flags]

Flags:

```bash
-h, --help                  help for file-benchmark
-g, --generate-config       Generate a config file to conect to S3
-c, --concurrent-jobs int   Number of jobs excuted writing files. (Maximum 10) (default 1)
-f, --files-number int      Number of files to be created. (default 1)
-m, --max-file-fize int     Max size in Mb allocated for each test file. (default 2000)
-i, --min-file-fize int     Min size in Mb allocated for each test file. (default 5)
-o, --output string         Output folder of the results. (default "./results")
-p, --path string           Path where the test files will be created. (default "./")
-s, --spinner               Show spinner, just to know if still alive the proccess, don't use if running on background

```
Config file Generation

file-bechmak-s3 requires a config file `s3Config.ini` this can be generated using the command:

```bash
$ ./file-benchmark-s3 -c10 -f 100 -m10 -i5 -o ./100files -p /mnt/s3fs/
```
after the file is generated, only need to fill the required variables


Usage Example:

```bash
$ ./file-benchmark-s3 -c10 -f 100 -m10 -i5 -o ./100files -p /mnt/s3fs/
```

In this example will create 100 random files, between 5 Mb and 10Mb usign 10 concurent jobs.

The files will be created inside a the path /mnt/s3fs/{random name} and  /mnt/s3fs/download_{random name}, after the test is complete, those folders will be removed.

The results will be written on the path ./100files, those results will be:

- result.csv  this file containt the following information
    - File path
	- File size in MB
	- MD5
	- Download File path
	- Creation time in Seconds
	- Upload time in Seconds
	- Download time in Seconds
	- Delete time in Seconds
	- Delete Downloaded Copy time in Seconds
	- Delete Object on S3 time in Seconds

- result.txt this file show more general information about the total execution
    - Start Time
    - End Time:
    - Duration:
    - Files Proccessed:
    - Concurrent Jobs