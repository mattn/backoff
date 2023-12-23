package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/opt"
)

const name = "backoff"

const version = "0.0.1"

var revision = "HEAD"

type status struct {
	Retry int       `json:"retry"`
	Time  time.Time `json:"time"`
}

func count(n int) string {
	switch n {
	case 1:
		return "1st"
	case 2:
		return "2nd"
	default:
		return fmt.Sprintf("%dth", n)
	}
}

func run() int {
	var f string
	var k string
	var off time.Duration
	var maxwait time.Duration
	var verbose bool
	var showVersion bool

	flag.StringVar(&f, "f", ".backoff", "storage file")
	flag.StringVar(&k, "k", "", "key (default is command-line)")
	flag.DurationVar(&off, "off", time.Minute, "offset")
	flag.DurationVar(&maxwait, "max", time.Hour, "max wait")
	flag.BoolVar(&verbose, "V", false, "verbose")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.Parse()

	if flag.NArg() == 0 {
		flag.Usage()
		os.Exit(2)
	}

	if showVersion {
		fmt.Println(version)
		os.Exit(0)
	}

	store, err := leveldb.OpenFile(f, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v: %v\n", os.Args[0], err)
		return 1
	}
	defer store.Close()

	cmd := exec.Command(flag.Arg(0), flag.Args()[1:]...)

	var bk []byte
	if k != "" {
		bk = []byte(k)
	} else {
		bk = []byte(cmd.String())
	}

	st := status{
		Time: time.Now(),
	}
	b, err := store.Get(bk, &opt.ReadOptions{DontFillCache: true})
	if err == nil {
		if err = json.Unmarshal(b, &st); err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", os.Args[0], err)
		} else {
			backoff := off * time.Duration(1<<(st.Retry-1))
			if maxwait == 0 || backoff > maxwait {
				backoff = maxwait
			}
			next := st.Time.Add(backoff)
			if time.Now().Before(next) {
				if verbose {
					fmt.Printf("%s retrying. waiting more %v\n", count(st.Retry), next.Sub(time.Now()))
				}
				return 0
			}
		}
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()

	code := cmd.ProcessState.ExitCode()
	if code == 0 {
		println(string(bk))
		if err = store.Delete(bk, &opt.WriteOptions{Sync: true}); err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", os.Args[0], err)
		}
	} else {
		st.Retry++
		st.Time = time.Now()
		b, err = json.Marshal(&st)
		if err == nil {
			err = store.Put(bk, b, &opt.WriteOptions{Sync: true})
		}
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v: %v\n", os.Args[0], err)
		}
	}
	return code
}

func main() {
	os.Exit(run())
}
