# TV Show Finder
Search for shows via themoviedb API and show the summary of a selected episode.

This project is an exercise to learn the Go programming language and therefore non-commercial.

## Usage
To run the program you need to register at themoviedb.org and request an API key.
This API key has to be stored into an environment variable before the program is executed.

This program was tested with Go 1.18 and 1.20 under Ubuntu 22.04 and Windows 11.

### Linux
In Bash run:
```
export TMDB_API_KEY="abcd1234"
go run ./main.go
```

### Windows
In PowerShell run:
```
$env:TMDB_API_KEY="abcd1234"
go run .\main.go
```
