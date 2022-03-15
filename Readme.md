# file-benchmark

file-benchmark is a tool to test speed on basic operations on files
useful when need to test speed on S3 mounted folders using s3fs or any othwer similar tool.

## Usage:
```
file-benchmark [flags]

Flags:

```bash
-c, --concurrent-jobs int   Number of jobs excuted writing files. (Maximum 10) (default 1)
-f, --files-number int      Number of files to be created. (default 1)
-h, --help                  help for file-benchmark
-m, --max-file-fize int     Max size in Mb allocated for each test file. (default 2000)
-i, --min-file-fize int     Min size in Mb allocated for each test file. (default 5)
-o, --output string         Output folder of the results. (default "./results")
-p, --path string           Path where the test files will be created. (default "./")
-s, --spinner               Show spinner, just to know if still alive the proccess, don't use if running on background

```

Usage Example:

```bash
$ ./file-benchmark -c10 -f 100 -m10 -i5 -o ./100files -p /mnt/s3fs/
```

in this example will create 100 random files, between 5 Mb and 10Mb usign 10 concurent jobs.

The files will be created inside a the path /mnt/s3fs/{random name} and  /mnt/s3fs/copy_{random name}, after the test is complete, those folders will be removed.

The results will be written on the path ./100files, those results will be:

- result.csv  this file containt the following information
    - File path
    - File size in MB
    - MD5
    - Copy File path
    - MD5 of Copy
    - Creation time in Seconds
    - Copy time in Seconds
    - MD5 verification Copy time in Seconds
    - Delete time in Seconds
    - Delete Copy time in Seconds
- result.txt this file show more general information about the total execution
    - Start Time
    - End Time:
    - Duration:
    - Files Proccessed:
    - Concurrent Jobs